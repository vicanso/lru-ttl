# lru-ttl

[![Build Status](https://github.com/vicanso/lru-ttl/workflows/Test/badge.svg)](https://github.com/vicanso/lru-ttl/actions)

LRU cache with ttl. It's useful for short ttl cache. 
L2Cache use lru cache for the first cache, and slow cache for the second cache. Lru cache should be set max entries for less memory usage but faster, slow cache is slower but more space.

## LRU TTL


```go
cache := lruttl.New(1000, 60 * time.Second)
cache.Add("tree.xie", "my data")
data, ok := cache.Get("tree.xie")
cache.Remove("tree.xie")
cache.Add("tree.xie", "my data", time.Second)
```

## L2Cache

```go
// redisCache
l2 := lruttl.NewL2Cache(redisCache, 200, 10 * time.Minute)
ctx := context.Background()
err := l2.Set(ctx, "key", &map[string]string{
    "name": "test",
})
fmt.Println(err)
m := make(map[string]string)
err = l2.Get(ctx, "key", &m)
fmt.Println(err)
fmt.Println(m)
```

## Ring

```go
ringCache := lruttl.NewRing(lruttl.RingCacheParams{
    Size:       10,
    MaxEntries: 1000,
    DefaultTTL: time.Minute,
})
lruCache := ringCache("key")
```