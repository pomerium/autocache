package autocache // import "github.com/pomerium/autocache"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang/groupcache"
	"github.com/hashicorp/memberlist"
)

var _ memberlist.EventDelegate = &Autocache{}

// Options are the configurations of a Autocache.
type Options struct {
	// Groupcache related
	Scheme      string
	Port        int
	CacheSize   int64
	GroupName   string
	EnableStats bool
	GetterFn    groupcache.Getter
	// Transport optionally specifies an http.RoundTripper for the client
	// to use when it makes a request to another groupcache node.
	// If nil, the client uses http.DefaultTransport.
	TransportFn func(context.Context) http.RoundTripper

	// Memberlist related
	SeedNodes        []string
	MemberlistConfig *memberlist.Config

	// Logger is a custom logger which you provide.
	Logger *log.Logger
}

func (o *Options) validate() error {
	if o.GetterFn == nil {
		return errors.New("groupcache requires getter fn")
	}
	if o.GroupName == "" {
		return errors.New("group name is required")
	}
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
	self  string
	peers []string
	// Handler are
	Handler http.Handler

	// groupcache related
	scheme    string
	port      int
	cacheSize int64
	Cache     *groupcache.Group
	pool      *groupcache.HTTPPool

	logger *log.Logger
}

// New creates a new Autocache instance, setups memberlist, and
// invokes groupcache's peer pooling handlers.
//
// NB: By default
func New(o *Options) (*Autocache, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	ac := Autocache{
		scheme:    o.Scheme,
		port:      o.Port,
		cacheSize: 4 << 20,
		logger:    o.Logger,
	}
	if o.Logger == nil {
		ac.logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if o.CacheSize != 0 {
		ac.cacheSize = o.CacheSize
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
	ac.Cache = groupcache.NewGroup(o.GroupName, ac.cacheSize, o.GetterFn)
	ac.pool = groupcache.NewHTTPPoolOpts(
		ac.groupcacheURL(ac.self),
		&groupcache.HTTPPoolOptions{
			BasePath: "/",
		},
	)
	if o.TransportFn != nil {
		ac.pool.Transport = o.TransportFn
	}

	mux := http.NewServeMux()
	if o.EnableStats {
		mux.HandleFunc("/stats/", ac.statsHandler)
	}
	mux.HandleFunc("/get/", ac.Get)
	mux.Handle("/", ac.pool)
	ac.Handler = mux
	if _, err := list.Join(o.SeedNodes); err != nil {
		return nil, err
	}
	return &ac, nil
}

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
	if ac.pool != nil {
		ac.pool.Set(ac.peers...)
		ac.logger.Printf("NotifyJoin:%s peers: %v", uri, len(ac.peers))
	}
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified. Implements memberlist's
// EventDelegate's interface.
func (ac *Autocache) NotifyLeave(node *memberlist.Node) {
	uri := ac.groupcacheURL(node.Addr.String())
	ac.removePeer(uri)
	ac.pool.Set(ac.peers...)
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

// Get attempts to retrieve the query param value of `key` from the cache.
func (ac *Autocache) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	now := time.Now()
	defer func() {
		log.Printf("cacheHandler: group[%s]\tkey[%q]\ttime[%v]", ac.Cache.Name(), key, time.Since(now))
	}()
	var respBody []byte
	if err := ac.Cache.Get(r.Context(), key, groupcache.AllocatingByteSliceSink(&respBody)); err != nil {
		ac.logger.Printf("Get/cache.Get error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (ac *Autocache) statsHandler(w http.ResponseWriter, r *http.Request) {
	respBody, err := json.Marshal(&ac.Cache.Stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
