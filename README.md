[![Build](https://github.com/pomerium/autocache/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/pomerium/autocache/actions/workflows/build.yml?query=branch%3Amain)
[![codecov](https://img.shields.io/codecov/c/github/pomerium/autocache.svg?style=flat)](https://codecov.io/gh/pomerium/autocache)
[![Go Report Card](https://goreportcard.com/badge/github.com/pomerium/autocache)](https://goreportcard.com/report/github.com/pomerium/autocache)
[![Go Reference](https://pkg.go.dev/badge/github.com/pomerium/autocache.svg)](https://pkg.go.dev/github.com/pomerium/autocache)
[![LICENSE](https://img.shields.io/github/license/pomerium/autocache.svg)](https://github.com/pomerium/autocache/blob/main/LICENSE)
[![discuss](https://img.shields.io/discourse/posts?server=https%3A%2F%2Fdiscuss.pomerium.com%2F&label=discuss)](https://discuss.pomerium.com/)

# Autocache

Autocache pairs [groupcache](https://github.com/golang/groupcache) with [memberlist](https://github.com/hashicorp/memberlist) for distributed peer discovery.

> **Note**: This project is experimental, not intended for production use, and may be archived in the future.

## Quick start

See `_example/` for usage.

### Run

```sh
docker-compose -f _example/docker-compose.yaml up --scale autocache=5
```

### Client

```sh
for i in `seq 10`; do curl "http://autocache.localhost/get/?key=hunter2"; echo; done
```

### Server

```text
autocache_2  | 2020/01/06 06:10:51 bcryptKey/key:"hunter2"      time:969.8645ms
autocache_2  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[969.9474ms]
autocache_1  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[1.3559ms]
autocache_3  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[1.1236ms]
autocache_4  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[1.2935ms]
autocache_5  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[985.2µs]
autocache_6  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[1.2163ms]
autocache_2  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[23.3µs]
autocache_1  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[495.3µs]
autocache_3  | 2020/01/06 06:10:51 cacheHandler: group[bcrypt]  key["hunter2"]  time[497.3µs]
autocache_4  | 2020/01/06 06:10:52 cacheHandler: group[bcrypt]  key["hunter2"]  time[770.5µs]
```

## License

Autocache is licensed under the [Apache 2.0 License](./LICENSE).
