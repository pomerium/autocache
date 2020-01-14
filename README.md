[![Build Status](https://github.com/pomerium/autocache/workflows/build/badge.svg)](https://github.com/pomerium/autocache/actions?workflow=build)
[![codecov](https://img.shields.io/codecov/c/github/pomerium/autocache.svg?style=flat)](https://codecov.io/gh/pomerium/autocache)
[![Go Report Card](https://goreportcard.com/badge/github.com/pomerium/autocache)](https://goreportcard.com/report/github.com/pomerium/autocache)
[![GoDoc](https://godoc.org/github.com/pomerium/autocache?status.svg)](https://godoc.org/github.com/pomerium/autocache)
[![LICENSE](https://img.shields.io/github/license/pomerium/autocache.svg)](https://github.com/pomerium/autocache/blob/master/LICENSE)
[![pomerium chat](https://img.shields.io/badge/chat-on%20slack-blue.svg?style=flat&logo=slack)](http://slack.pomerium.io)

# Autocache

[Groupcache](https://github.com/golang/groupcache) enhanced with [memberlist](https://github.com/hashicorp/memberlist) for distributed peer discovery.

## TL;DR

See `/_example/` for usage.

### Run

`docker-compose up -f _example/docker-compose.yaml --scale autocache=5`

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
autocache_2  | 2020/01/06 06:10:51 bcryptKey/key:"hunter2"	time:969.8645ms
autocache_2  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[969.9474ms]
autocache_1  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[1.3559ms]
autocache_3  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[1.1236ms]
autocache_4  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[1.2935ms]
autocache_5  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[985.2µs]
autocache_6  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[1.2163ms]
autocache_2  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[23.3µs]
autocache_1  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[495.3µs]
autocache_3  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]	key["hunter2"]	time[497.3µs]
autocache_4  | 2020/01/06 06:10:52 cacheHandler: group[bcrypt]	key["hunter2"]	time[770.5µs]
```
