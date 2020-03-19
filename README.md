# lru-ttl

[![Build Status](https://img.shields.io/travis/vicanso/lru-ttl.svg?label=linux+build)](https://travis-ci.org/vicanso/lru-ttl)

LRU cache with ttl. It's useful for short ttl cache. 

```go
cache := New(1000, 60 * time.Second)
cache.Add("tree.xie", "my data")
data, ok := cache.Get("tree.xie")
cache.Remove("tree.xie")
```
