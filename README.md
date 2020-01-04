[![pomerium chat](https://img.shields.io/badge/chat-on%20slack-blue.svg?style=flat&logo=slack)](http://slack.pomerium.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/pomerium/autocache)](https://goreportcard.com/report/github.com/pomerium/autocache)
[![GoDoc](https://godoc.org/github.com/pomerium/autocache?status.svg)](https://godoc.org/github.com/pomerium/autocache)
[![LICENSE](https://img.shields.io/github/license/pomerium/autocache.svg)](https://github.com/pomerium/autocache/blob/master/LICENSE)

# Autocache

[Groupcache](https://github.com/golang/groupcache) enhanced with [memberlist](https://github.com/hashicorp/memberlist) for distributed peer discovery.

## TL;DR

Run `docker-compose up --scale autocache=5`

### Client

```bash
for i in`seq 10`; do curl "http://autocache.localhost/get/?key=hunter42";echo; done

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
autocache_4  | bcryptKey/key:"hunter42"	time:920.3141ms
autocache_2  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[921.3286ms]
autocache_1  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[1.3112ms]
autocache_5  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[783.4µs]
autocache_4  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[11.5µs]
autocache_3  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[1.2833ms]
autocache_2  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[735.6µs]
autocache_1  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[539.8µs]
autocache_5  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[902.2µs]
autocache_4  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[11.7µs]
autocache_3  | cacheHandler: group[bcryptKey]	key["hunter42"]	time[697.2µs]
```
