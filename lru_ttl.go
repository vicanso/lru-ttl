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

	"github.com/golang/groupcache/lru"
)

type Cache struct {
	MaxEntries int
	TTL        time.Duration
	c          *lru.Cache
}

type cacheItem struct {
	expiredAt int64
	value     interface{}
}

// New creates a new Cache.
func New(maxEntries int, defaultTTL time.Duration) *Cache {
	if maxEntries <= 0 || defaultTTL < 0 {
		panic(errors.New("maxEntries and default ttl must be gt 0"))
	}
	return &Cache{
		MaxEntries: maxEntries,
		TTL:        defaultTTL,
		c:          lru.New(maxEntries),
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value interface{}, ttl ...time.Duration) {
	expiredAt := time.Now().UnixNano()
	if len(ttl) != 0 {
		expiredAt += ttl[0].Nanoseconds()
	} else {
		expiredAt += c.TTL.Nanoseconds()
	}
	c.c.Add(key, &cacheItem{
		expiredAt: expiredAt,
		value:     value,
	})
}

// Get gets a key's value from the cache.
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	data, ok := c.c.Get(key)
	if !ok {
		return
	}
	item, ok := data.(*cacheItem)
	if !ok {
		return
	}
	if item.expiredAt < time.Now().UnixNano() {
		ok = false
		return
	}
	value = item.value
	ok = true
	return
}

// Remove removes the key's value from the cache.
func (c *Cache) Remove(key string) {
	c.c.Remove(key)
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	return c.c.Len()
}
