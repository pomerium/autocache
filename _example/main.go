package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/groupcache"
	"github.com/pomerium/autocache"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultAddr = ":http"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	existing := []string{}
	if nodes := os.Getenv("NODES"); nodes != "" {
		existing = strings.Split(nodes, ",")
	}

	ac, err := autocache.New(&autocache.Options{})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := ac.Join(existing); err != nil {
		log.Fatal(err)
	}
	var exampleCache cache
	exampleCache.group = groupcache.NewGroup("bcrypt", 1<<20, exampleCache)

	mux := http.NewServeMux()
	mux.Handle("/get/", exampleCache)
	mux.Handle("/_groupcache/", ac)
	log.Fatal(http.ListenAndServe(addr, mux))

}

type cache struct {
	group *groupcache.Group
}

// Get is am arbitrary getter function. Bcrypt is nice here because, it:
//	1) takes a long time
//	2) uses a random seed so non-cache results for the same key are obvious
func (ac cache) Get(ctx context.Context, key string, dst groupcache.Sink) error {
	now := time.Now()
	defer func() {
		log.Printf("bcryptKey/key:%q\ttime:%v", key, time.Since(now))
	}()
	out, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	if err != nil {
		return err
	}
	if err := dst.SetBytes(out); err != nil {
		return err
	}
	return nil
}

func (ac cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("cacheHandler: group[%s]\tkey[%q]\ttime[%v]", ac.group.Name(), key, time.Since(now))
	}()
	var respBody []byte
	if err := ac.group.Get(r.Context(), key, groupcache.AllocatingByteSliceSink(&respBody)); err != nil {
		log.Printf("Get/cache.Get error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(respBody))
}
