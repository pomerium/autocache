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
	// Groupcache uses normal HTTP for RPC. We need to know the scheme
	// and port to use to build the RPC request.
	Scheme string
	Port   int
	// Transport optionally specifies an http.RoundTripper for the client
	// to use when it makes a request to another groupcache node.
	// If nil, the client uses http.DefaultTransport.
	TransportFn func(context.Context) http.RoundTripper
	PoolOptions *groupcache.HTTPPoolOptions

	// Memberlist related
	//
	// SeedNodes is a slice of addresses we use to bootstrap peer discovery
	SeedNodes []string
	// MemberlistConfig ist he memberlist configuration to use.
	// If empty, `DefaultLANConfig` is used.
	MemberlistConfig *memberlist.Config

	// Logger is a custom logger which you provide.
	Logger *log.Logger
}

func (o *Options) validate() error {
	if o.Scheme == "" {
		return errors.New("scheme is required")
	}
	if o.Port == 0 {
		return errors.New("port is required")
	}
	if len(o.SeedNodes) == 0 {
		return errors.New("must supply at least one seed node")
	}
	return nil
}

type Autocache struct {
	Pool *groupcache.HTTPPool

	self   string
	peers  []string
	scheme string
	port   int

	logger *log.Logger
}

// New creates a new Autocache instance, setups memberlist, and
// invokes groupcache's peer pooling handlers.
func New(o *Options) (*Autocache, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	ac := Autocache{
		scheme: o.Scheme,
		port:   o.Port,
		logger: o.Logger,
	}
	if o.Logger == nil {
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

	if _, err := list.Join(o.SeedNodes); err != nil {
		return nil, err
	}
	return &ac, nil
}

// groupcacheURL builds a groupcache friendly RPC url from an address
func (ac *Autocache) groupcacheURL(addr string) string {
	return fmt.Sprintf("%s://%s:%d", ac.scheme, addr, ac.port)
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
		ac.logger.Printf("NotifyJoin:%s peers: %v", uri, len(ac.peers))
	}
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified. Implements memberlist's
// EventDelegate's interface.
func (ac *Autocache) NotifyLeave(node *memberlist.Node) {
	uri := ac.groupcacheURL(node.Addr.String())
	ac.removePeer(uri)
	ac.Pool.Set(ac.peers...)
	ac.logger.Printf("NotifyLeave:%s peers: %v", uri, len(ac.peers))
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified. Implements memberlist EventDelegate's interface.
func (ac *Autocache) NotifyUpdate(node *memberlist.Node) {
	ac.logger.Printf("NotifyUpdate: %+v\n", node)
}

func (ac *Autocache) removePeer(uri string) {
	for i := 0; i < len(ac.peers); i++ {
		if ac.peers[i] == uri {
			ac.peers = append(ac.peers[:i], ac.peers[i+1:]...)
			i--
		}
	}
}
