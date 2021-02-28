// Copyright 2020 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lruttl

import (
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type Key interface{}

type Cache struct {
	// mutex     *sync.RWMutex
	ttl       time.Duration
	lru       *lru.Cache
	onEvicted func(key Key, value interface{})
}

// CacheOption cache option
type CacheOption func(c *Cache)

type cacheItem struct {
	expiredAt int64
	value     interface{}
}

func (item *cacheItem) isExpired() bool {
	return item.expiredAt < time.Now().UnixNano()
}

// New returns a new lru cache with ttl
func New(maxEntries int, defaultTTL time.Duration, opts ...CacheOption) *Cache {
	if maxEntries <= 0 || defaultTTL <= 0 {
		panic(errors.New("maxEntries and default ttl must be gt 0"))
	}
	c := &Cache{
		ttl: defaultTTL,
	}
	l, err := lru.NewWithEvict(maxEntries, func(key, value interface{}) {
		if c.onEvicted != nil {
			c.onEvicted(key, value)
		}
	})
	if err != nil {
		panic(err)
	}
	c.lru = l
	for _, opt := range opts {
		opt(c)
	}
	return c

}

// CacheEvictedOption sets evicted function to cache
func CacheEvictedOption(fn func(key Key, value interface{})) CacheOption {
	return func(c *Cache) {
		c.onEvicted = fn
	}
}

// Add adds a value to the cache, it will use default ttl if the ttl is nil.
func (c *Cache) Add(key Key, value interface{}, ttl ...time.Duration) {
	expiredAt := time.Now().UnixNano()
	if len(ttl) != 0 {
		expiredAt += ttl[0].Nanoseconds()
	} else {
		expiredAt += c.ttl.Nanoseconds()
	}
	c.lru.Add(key, &cacheItem{
		expiredAt: expiredAt,
		value:     value,
	})
}

// Get returns value and exists from the cache by key, if value is expired then remove it.
// If the value is expired, value is not nil but exists is false.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	data, ok := c.lru.Get(key)
	if !ok {
		return
	}
	item, ok := data.(*cacheItem)
	if !ok {
		return
	}
	// 过期的元素数据也返回，但ok为false
	value = item.value
	if item.isExpired() {
		ok = false
		// 过期的元素删除
		c.lru.Remove(key)
		return
	}
	ok = true
	return
}

// Peek get a key's value from the cache, but not move to front.
// The performance is better than get.
// It will not remove it if the cache is expired.
func (c *Cache) Peek(key Key) (value interface{}, ok bool) {
	data, ok := c.lru.Peek(key)
	if !ok {
		return
	}
	item, ok := data.(*cacheItem)
	if !ok {
		return
	}
	// 过期的元素数据也返回，但ok为false
	value = item.value
	if item.isExpired() {
		ok = false
		// 过期不清除
		return
	}
	ok = true
	return
}

// Remove removes the key's value from the cache.
func (c *Cache) Remove(key Key) {
	c.lru.Remove(key)
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	return c.lru.Len()
}

// Keys gets all keys of cache
func (c *Cache) Keys() []Key {
	keys := c.lru.Keys()
	result := make([]Key, len(keys))
	for i, k := range keys {
		result[i] = k
	}
	return result
}
