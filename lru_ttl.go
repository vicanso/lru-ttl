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
	mutex     *sync.RWMutex
	ttl       time.Duration
	lru       *LRUCache
	onEvicted func(key Key, value interface{})
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
	c.mutex = &sync.RWMutex{}
	return c
}

// NewWithoutRWMutex create a new cache without rw mutex
func NewWithoutRWMutex(maxEntries int, defaultTTL time.Duration) *Cache {
	if maxEntries <= 0 || defaultTTL <= 0 {
		panic(errors.New("maxEntries and default ttl must be gt 0"))
	}
	return &Cache{
		ttl: defaultTTL,
		lru: NewLRU(maxEntries),
	}
}

// SetOnEvicted set on evicted function
func (c *Cache) SetOnEvicted(fn func(key Key, value interface{})) {
	c.onEvicted = fn
	c.lru.OnEvicted = fn
}

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value interface{}, ttl ...time.Duration) {
	expiredAt := time.Now().UnixNano()
	if len(ttl) != 0 {
		expiredAt += ttl[0].Nanoseconds()
	} else {
		expiredAt += c.ttl.Nanoseconds()
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

// Get get a key's value from the cache.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	// 因为会更新其顺序，因此需要使用写锁
	if c.mutex != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
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
		c.lru.Remove(key)
		return
	}
	ok = true
	return
}

// Peek get a key's value from the cache, but move to front.
func (c *Cache) Peek(key Key) (value interface{}, ok bool) {
	// 因为不会更新其顺序，因此可以使用读锁
	if c.mutex != nil {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
	}
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
	if c.mutex != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
	c.lru.Remove(key)
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.mutex != nil {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
	}
	count := 0
	c.forEach(func(key Key, _ interface{}) {
		count++
	})
	return count
}

// forEach for each function
// 避免public的方法之间的调用，public的方法只调用private的方法
// 由于只有public的方法会使用锁，这样可以避免代码有误导致死锁
func (c *Cache) forEach(fn func(key Key, value interface{})) {
	c.lru.ForEach(func(lruKey Key, lruValue interface{}) {
		item, ok := lruValue.(*cacheItem)
		if !ok || item.isExpired() {
			return
		}
		// 返回key与value
		fn(lruKey, item.value)
	})
}

// ForEach for each function
func (c *Cache) ForEach(fn func(key Key, value interface{})) {
	if c.mutex != nil {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
	}
	c.forEach(fn)
}

// Keys get all keys of cache
func (c *Cache) Keys() []Key {
	if c.mutex != nil {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
	}
	result := make([]Key, 0)
	c.forEach(func(key Key, _ interface{}) {
		result = append(result, key)
	})
	return result
}
