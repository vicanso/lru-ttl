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
	"sync"
	"time"
)

type Cache struct {
	mutex      *sync.RWMutex
	MaxEntries int
	TTL        time.Duration
	lru        *LRUCache
}

type cacheItem struct {
	expiredAt int64
	value     interface{}
}

func (item *cacheItem) isExpired() bool {
	return item.expiredAt < time.Now().UnixNano()
}

// New creates a new cache with rw mutex
func New(maxEntries int, defaultTTL time.Duration) *Cache {
	c := NewWithoutRWMutex(maxEntries, defaultTTL)
	c.mutex = new(sync.RWMutex)
	return c
}

// NewWithoutRWMutex create a new cache without rw mutex
func NewWithoutRWMutex(maxEntries int, defaultTTL time.Duration) *Cache {
	if maxEntries <= 0 || defaultTTL <= 0 {
		panic(errors.New("maxEntries and default ttl must be gt 0"))
	}
	return &Cache{
		MaxEntries: maxEntries,
		TTL:        defaultTTL,
		lru:        NewLRU(maxEntries),
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value interface{}, ttl ...time.Duration) {
	expiredAt := time.Now().UnixNano()
	if len(ttl) != 0 {
		expiredAt += ttl[0].Nanoseconds()
	} else {
		expiredAt += c.TTL.Nanoseconds()
	}
	if c.mutex != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
	c.lru.Add(key, &cacheItem{
		expiredAt: expiredAt,
		value:     value,
	})
}

// Get gets a key's value from the cache.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.mutex != nil {
		c.mutex.RLock()
	}
	data, ok := c.lru.Get(key)
	// release lock asap
	if c.mutex != nil {
		c.mutex.RUnlock()
	}
	if !ok {
		return
	}
	item, ok := data.(*cacheItem)
	if !ok {
		return
	}
	if item.isExpired() {
		ok = false
		return
	}
	value = item.value
	ok = true
	return
}

// Remove removes the key's value from the cache.
func (c *Cache) Remove(key Key) {
	if c.mutex != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
	c.lru.Remove(key)
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	return len(c.Keys())
}

// ForEach for each the items of cache
func (c *Cache) ForEach(fn func(key Key, value interface{})) {
	if c.mutex != nil {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
	}
	for _, e := range c.lru.cache {
		kv := e.Value.(*entry)
		item, ok := kv.value.(*cacheItem)
		if !ok || item.isExpired() {
			continue
		}
		// 返回key与value
		fn(kv.key, item.value)
	}
}

// Keys get all keys of cache
func (c *Cache) Keys() []Key {
	result := make([]Key, 0)
	c.ForEach(func(key Key, _ interface{}) {
		result = append(result, key)
	})
	return result
}
