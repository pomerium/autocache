package main

import (
	"context"
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

	o := autocache.Options{
		Scheme:    "http",
		Port:      80,
		SeedNodes: existing,
		GroupName: "bcryptKey",
		GetterFn:  groupcache.GetterFunc(bcryptKey)}
	ac, err := autocache.New(&o)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(addr, ac.Handler))

}

// bcryptKey is am arbitrary getter function. In this example, we simply bcrypt
// the key which is useful because bcrypt:
// 		1) takes a long time
//		2) uses a random seed so non-cache results for the same key are obvious
func bcryptKey(ctx context.Context, key string, dst groupcache.Sink) error {
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
