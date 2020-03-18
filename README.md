# lru-ttl

LRU cache with ttl. It's useful for short ttl cache.

```go
cache := New(1000, 60 * time.Second)
cache.Add("tree.xie", "my data")
data, ok := cache.Get("tree.xie")
cache.Remove("tree.xie")
```