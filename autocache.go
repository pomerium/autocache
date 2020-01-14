package autocache // import "github.com/pomerium/autocache"

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang/groupcache"
	"github.com/hashicorp/memberlist"
)

var _ memberlist.EventDelegate = &Autocache{}

// Options are the configurations of a Autocache.
type Options struct {
	// Groupcache related
	//
	// Groupcache's pool is a HTTP handler. Scheme and port should be set
	// such that group cache's internal http client, used to fetch, distributed
	// keys, knows how to build the request URL.
	PoolOptions *groupcache.HTTPPoolOptions
	PoolScheme  string
	PoolPort    int
	// Transport optionally specifies an http.RoundTripper for the client
	// to use when it makes a request to another groupcache node.
	// If nil, the client uses http.DefaultTransport.
	PoolTransportFn func(context.Context) http.RoundTripper
	// Context optionally specifies a context for the server to use when it
	// receives a request.
	// If nil, the server uses the request's context
	PoolContext func(*http.Request) context.Context

	// Memberlist related
	//
	// MemberlistConfig ist he memberlist configuration to use.
	// If empty, `DefaultLANConfig` is used.
	MemberlistConfig *memberlist.Config

	// Logger is a custom logger which you provide.
	Logger *log.Logger
}

// Autocache implements automatic, distributed membership for a cluster
// of cache pool peers.
type Autocache struct {
	GroupcachePool *groupcache.HTTPPool
	Memberlist     *memberlist.Memberlist

	self   string
	peers  []string
	scheme string
	port   string

	logger *log.Logger
}

// New creates a new Autocache instance, setups memberlist, and
// invokes groupcache's peer pooling handlers. Note, by design a groupcache
// pool can only be made _once_.
func New(o *Options) (*Autocache, error) {
	var err error
	ac := Autocache{
		scheme: o.PoolScheme,
		port:   fmt.Sprintf("%d", o.PoolPort),
		logger: o.Logger,
	}
	if ac.logger == nil {
		ac.logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	ac.logger.Printf("autocache: with options: %+v", o)

	if ac.scheme == "" {
		ac.logger.Printf("autocache: pool scheme not set, assuming http://")
		ac.scheme = "http"
	}
	if ac.port == "0" {
		ac.logger.Printf("autocache: pool port not set, assuming empty")
		ac.port = ""
	}

	mlConfig := o.MemberlistConfig
	if mlConfig == nil {
		ac.logger.Println("autocache: defaulting to lan configuration")
		mlConfig = memberlist.DefaultLANConfig()
	}
	mlConfig.Events = &ac
	mlConfig.Logger = ac.logger
	if ac.Memberlist, err = memberlist.Create(mlConfig); err != nil {
		return nil, fmt.Errorf("autocache: can't create memberlist: %w", err)
	}
	// the only way memberlist would be empty here, following create is if
	// the current node suddenly died. Still, we check to be safe.
	if len(ac.Memberlist.Members()) == 0 {
		return nil, errors.New("memberlist can't find self")
	}
	self := ac.Memberlist.Members()[0]
	if self.Addr == nil {
		return nil, errors.New("self addr cannot be nil")
	}
	ac.self = self.Addr.String()
	ac.logger.Printf("autocache: self addr is: %s", ac.self)
	poolOptions := &groupcache.HTTPPoolOptions{}
	if o.PoolOptions != nil {
		poolOptions = o.PoolOptions
	}
	gcSelf := ac.groupcacheURL(ac.self)
	ac.logger.Printf("autocache groupcache self: %s options: %+v", gcSelf, poolOptions)
	ac.GroupcachePool = groupcache.NewHTTPPoolOpts(gcSelf, poolOptions)
	if o.PoolTransportFn != nil {
		ac.GroupcachePool.Transport = o.PoolTransportFn
	}
	if o.PoolContext != nil {
		ac.GroupcachePool.Context = o.PoolContext
	}
	return &ac, nil
}

// Join is used to take an existing Memberlist and attempt to join a cluster
// by contacting all the given hosts and performing a state sync. Initially,
// the Memberlist only contains our own state, so doing this will cause
// remote nodes to become aware of the existence of this node, effectively
// joining the cluster.
//
// This returns the number of hosts successfully contacted and an error if
// none could be reached. If an error is returned, the node did not successfully
// join the cluster.
func (ac *Autocache) Join(existing []string) (int, error) {
	if ac.Memberlist == nil {
		return 0, errors.New("memberlist cannot be nil")
	}
	return ac.Memberlist.Join(existing)
}

// groupcacheURL builds a groupcache friendly RPC url from an address
func (ac *Autocache) groupcacheURL(addr string) string {
	u := fmt.Sprintf("%s://%s", ac.scheme, addr)
	if ac.port != "" {
		u = fmt.Sprintf("%s:%s", u, ac.port)
	}
	return u
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified. Implements memberlist's
// EventDelegate's interface.
func (ac *Autocache) NotifyJoin(node *memberlist.Node) {
	uri := ac.groupcacheURL(node.Addr.String())
	ac.removePeer(uri)
	ac.peers = append(ac.peers, uri)
	if ac.GroupcachePool != nil {
		ac.GroupcachePool.Set(ac.peers...)
		ac.logger.Printf("Autocache/NotifyJoin: %s peers: %v", uri, len(ac.peers))
	}
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified. Implements memberlist's
// EventDelegate's interface.
func (ac *Autocache) NotifyLeave(node *memberlist.Node) {
	uri := ac.groupcacheURL(node.Addr.String())
	ac.removePeer(uri)
	ac.GroupcachePool.Set(ac.peers...)
	ac.logger.Printf("Autocache/NotifyLeave: %s peers: %v", uri, len(ac.peers))
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified. Implements memberlist EventDelegate's interface.
func (ac *Autocache) NotifyUpdate(node *memberlist.Node) {
	ac.logger.Printf("Autocache/NotifyUpdate: %+v", node)
}

func (ac *Autocache) removePeer(uri string) {
	for i := 0; i < len(ac.peers); i++ {
		if ac.peers[i] == uri {
			ac.peers = append(ac.peers[:i], ac.peers[i+1:]...)
			i--
		}
	}
}

func (ac *Autocache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ac.GroupcachePool == nil {
		http.Error(w, "pool not initialized", http.StatusInternalServerError)
		return
	}
	ac.GroupcachePool.ServeHTTP(w, r)
}
