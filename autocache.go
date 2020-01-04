package autocache // import "github.com/pomerium/autocache"

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	GetterFn    groupcache.Getter
	EnableStats bool

	// Memberlist related
	SeedNodes        []string
	MemberlistConfig *memberlist.Config
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
	self    string
	peers   []string
	Handler http.Handler

	// groupcache related
	scheme    string
	port      int
	cacheSize int64
	cache     *groupcache.Group
	pool      *groupcache.HTTPPool

	// todo(bdd): support custom logger
	// Logger *log.Logger
}

func New(o *Options) (*Autocache, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	ac := Autocache{
		scheme:    o.Scheme,
		port:      o.Port,
		cacheSize: 4 << 20,
	}
	if o.CacheSize != 0 {
		ac.cacheSize = o.CacheSize
	}
	mlConfig := memberlist.DefaultLANConfig()
	if o.MemberlistConfig != nil {
		mlConfig = o.MemberlistConfig
	}
	mlConfig.Events = &ac
	list, err := memberlist.Create(mlConfig)
	if err != nil {
		return nil, err
	}
	if len(list.Members()) == 0 {
		return nil, errors.New("memberlist can't find self")
	}
	if list.Members()[0].Addr == nil {
		return nil, errors.New("self addr cannot be nil")
	}
	ac.self = list.Members()[0].Addr.String()
	ac.cache = groupcache.NewGroup(o.GroupName, ac.cacheSize, o.GetterFn)
	ac.pool = groupcache.NewHTTPPoolOpts(ac.groupcacheURL(ac.self), &groupcache.HTTPPoolOptions{})

	mux := http.NewServeMux()
	if o.EnableStats {
		mux.HandleFunc("/stats/", ac.statsHandler)
	}
	// todo(bdd): add signature verify middleware
	//
	mux.HandleFunc("/get/", ac.cacheHandler)
	// in a real app you probably want this served from a different listener
	// the default handler actually panics(!?!) on unknown route
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
		log.Printf("NotifyJoin: %s\tpeers: %v", uri, len(ac.peers))
	}
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified. Implements memberlist's
// EventDelegate's interface.
func (ac *Autocache) NotifyLeave(node *memberlist.Node) {
	uri := ac.groupcacheURL(node.Addr.String())
	ac.removePeer(uri)
	ac.pool.Set(ac.peers...)
	log.Printf("NotifyLeave: %s\tpeers: %v", uri, len(ac.peers))
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified. Implements memberlist EventDelegate's interface.
func (ac *Autocache) NotifyUpdate(node *memberlist.Node) {
	log.Printf("NotifyUpdate: %+v\n", node)
}

func (ac *Autocache) removePeer(uri string) {
	for i := 0; i < len(ac.peers); i++ {
		if ac.peers[i] == uri {
			ac.peers = append(ac.peers[:i], ac.peers[i+1:]...)
			i--
		}
	}
}

func (ac *Autocache) cacheHandler(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	now := time.Now()
	defer func() {
		log.Printf("cacheHandler: group[%s]\tkey[%q]\ttime[%v]", ac.cache.Name(), key, time.Since(now))
	}()
	var respBody []byte
	if err := ac.cache.Get(r.Context(), key, groupcache.AllocatingByteSliceSink(&respBody)); err != nil {
		log.Printf("cacheHandler/cache.Get: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (ac *Autocache) statsHandler(w http.ResponseWriter, r *http.Request) {
	respBody, err := json.Marshal(&ac.cache.Stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
