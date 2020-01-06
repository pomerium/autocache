package autocache // import "github.com/pomerium/autocache"

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/golang/groupcache"
	"github.com/hashicorp/memberlist"
)

var _ memberlist.EventDelegate = &Autocache{}

// Options are the configurations of a Autocache.
type Options struct {
	// Groupcache related
	//
	// Transport optionally specifies an http.RoundTripper for the client
	// to use when it makes a request to another groupcache node.
	// If nil, the client uses http.DefaultTransport.
	TransportFn func(context.Context) http.RoundTripper
	PoolOptions *groupcache.HTTPPoolOptions

	// Memberlist related
	//
	// SeedNodes is a slice of addresses we use to bootstrap peer discovery
	// Seed nodes should contain a list of valid URLs including scheme and port
	// if those are used to connect to your group cache cluster. (e.g. "https://example.net:8000")
	// Memberlist will be bootstrapped using just the hostname of those seed URLS.
	SeedNodes []string
	// MemberlistConfig ist he memberlist configuration to use.
	// If empty, `DefaultLANConfig` is used.
	MemberlistConfig *memberlist.Config

	// Logger is a custom logger which you provide.
	Logger *log.Logger
}

func (o *Options) validate() error {
	if len(o.SeedNodes) == 0 {
		return errors.New("must supply at least one seed node")
	}
	u, err := url.Parse(o.SeedNodes[0])
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		return fmt.Errorf("%s has no scheme", u.String())
	}
	return nil
}

type Autocache struct {
	Pool *groupcache.HTTPPool

	self   string
	peers  []string
	scheme string
	port   string

	logger *log.Logger
}

// New creates a new Autocache instance, setups memberlist, and
// invokes groupcache's peer pooling handlers.
func New(o *Options) (*Autocache, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	var ac Autocache

	u, _ := url.Parse(o.SeedNodes[0]) // err checked in validate
	ac.scheme = u.Scheme
	ac.port = u.Port()

	ac.logger = o.Logger
	if ac.logger == nil {
		ac.logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	mlConfig := o.MemberlistConfig
	if mlConfig == nil {
		ac.logger.Println("defaulting to lan configuration")
		mlConfig = memberlist.DefaultLANConfig()
	}
	mlConfig.Events = &ac
	mlConfig.Logger = ac.logger
	list, err := memberlist.Create(mlConfig)
	if err != nil {
		return nil, err
	}
	if len(list.Members()) == 0 {
		return nil, errors.New("memberlist can't find self")
	}
	if list.Members()[0].Addr == nil {
		return nil, errors.New("memberlist self addr cannot be nil")
	}
	ac.self = list.Members()[0].Addr.String()
	poolOptions := &groupcache.HTTPPoolOptions{}
	if o.PoolOptions != nil {
		poolOptions = o.PoolOptions
	}
	ac.Pool = groupcache.NewHTTPPoolOpts(ac.groupcacheURL(ac.self), poolOptions)
	if o.TransportFn != nil {
		ac.Pool.Transport = o.TransportFn
	}

	seeds := make([]string, len(o.SeedNodes))
	for k, v := range o.SeedNodes {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		seeds[k] = u.Hostname()
	}

	if _, err := list.Join(seeds); err != nil {
		return nil, fmt.Errorf("couldn't join memberlist cluster: %w", err)
	}
	return &ac, nil
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
	if ac.Pool != nil {
		ac.Pool.Set(ac.peers...)
		ac.logger.Printf("Autocache/NotifyJoin: %s peers: %v", uri, len(ac.peers))
	}
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified. Implements memberlist's
// EventDelegate's interface.
func (ac *Autocache) NotifyLeave(node *memberlist.Node) {
	uri := ac.groupcacheURL(node.Addr.String())
	ac.removePeer(uri)
	ac.Pool.Set(ac.peers...)
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
	if ac.Pool == nil {
		http.Error(w, "pool not initialized", http.StatusInternalServerError)
		return
	}
	ac.Pool.ServeHTTP(w, r)
}
