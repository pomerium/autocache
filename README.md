[![pomerium chat](https://img.shields.io/badge/chat-on%20slack-blue.svg?style=flat&logo=slack)](http://slack.pomerium.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/pomerium/autocache)](https://goreportcard.com/report/github.com/pomerium/autocache)
[![GoDoc](https://godoc.org/github.com/pomerium/autocache?status.svg)](https://godoc.org/github.com/pomerium/autocache)
[![LICENSE](https://img.shields.io/github/license/pomerium/autocache.svg)](https://github.com/pomerium/autocache/blob/master/LICENSE)

# Autocache

[Groupcache](https://github.com/golang/groupcache) enhanced with [memberlist](https://github.com/hashicorp/memberlist) for distributed peer discovery.

## TL;DR

See `cmd/autocache/main.go` for usage.

### Run

`docker-compose up --scale autocache=5`

### Client

```bash
for i in`seq 10`; do curl "http://autocache.localhost/get/?key=hunter2";echo; done

```

```
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
$2a$14$1CCq.8WOxEmLY3jdkwZKIeR1bN/B0jnWwwSKc1VTf60A57VOXKblC
```

### Server

```
autocache_5  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[1.4064ms]
autocache_3  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[1.1171ms]
autocache_4  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[12.9µs]
autocache_1  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[916.9µs]
autocache_2  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[903.5µs]
autocache_5  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[544µs]
autocache_3  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[534.6µs]
autocache_4  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[19.2µs]
autocache_1  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[796.1µs]
autocache_2  | 2020/01/04 23:46:43 cacheHandler: group[bcrypt]	key["hunter2"]	time[626.9µs]
```
