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

// L2Cache use lru cache for the first cache, and slow cache for the second cache.
// LRU cache should be set max entries for less memory usage but faster,
// slow cache is slower and using more space, but it can store more data

package lruttl

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

type SlowCache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
}

// L2CacheOption l2cache option
type L2CacheOption func(c *L2Cache)

// A l2cache for frequently visited  data

type L2Cache struct {
	// prefix is the prefix of all key, it will auto prepend to the key
	prefix string
	// ttl is the duration for cache
	ttl time.Duration
	// ttlCache is the ttl lru cache
	ttlCache *Cache
	// slowCache is the slow cache for more data
	slowCache SlowCache
	// marshal is custom marshal function.
	// It will be json.Marshal if not set
	marshal func(v interface{}) ([]byte, error)
	// unmarshal is custom unmarshal function.
	// It will be json.Unmarshal if not set
	unmarshal func(data []byte, v interface{}) error
}

// ErrIsNil is the error of nil cache
var ErrIsNil = errors.New("cache is nil")

// ErrInvalidType is the error of invalid type
var ErrInvalidType = errors.New("invalid type")

// BufferMarshal converts *bytes.Buffer to bytes,
// it returns a ErrInvalidType if restult is not *bytes.Buffer
func BufferMarshal(result interface{}) ([]byte, error) {
	buf, ok := result.(*bytes.Buffer)
	if !ok {
		return nil, ErrInvalidType
	}
	return buf.Bytes(), nil
}

// BufferUnmarshal writes the data to buffer,
// it returns a ErrInvalidType if restult is not *bytes.Buffer
func BufferUnmarshal(data []byte, result interface{}) error {
	buf, ok := result.(*bytes.Buffer)
	if !ok {
		return ErrInvalidType
	}
	buf.Write(data)
	return nil
}

// NewL2Cache return a new L2Cache,
// it returns panic if maxEntries or defaultTTL is nil
func NewL2Cache(slowCache SlowCache, maxEntries int, defaultTTL time.Duration, opts ...L2CacheOption) *L2Cache {
	c := &L2Cache{
		ttl:       defaultTTL,
		ttlCache:  New(maxEntries, defaultTTL),
		slowCache: slowCache,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// L2CacheMarshalOption set custom marshal function for l2cache
func L2CacheMarshalOption(fn func(v interface{}) ([]byte, error)) L2CacheOption {
	return func(c *L2Cache) {
		c.marshal = fn
	}
}

// L2CacheUnmarshalOption set custom unmarshal function for l2cache
func L2CacheUnmarshalOption(fn func(data []byte, v interface{}) error) L2CacheOption {
	return func(c *L2Cache) {
		c.unmarshal = fn
	}
}

// L2CachePrefixOption set prefix for l2cache
func L2CachePrefixOption(prefix string) L2CacheOption {
	return func(c *L2Cache) {
		c.prefix = prefix
	}
}

func (l2 *L2Cache) getKey(key string) string {
	return l2.prefix + key
}

// Get first get cache from lru, if not exists,
// then get the data from slow cache.
// Use unmarshal function covert the data to result
func (l2 *L2Cache) Get(key string, result interface{}) (err error) {
	key = l2.getKey(key)
	v, ok := l2.ttlCache.Get(key)
	// 如果获取到数据不为空，但是ok为false
	// 表示数据已过期，数据已过期则返回ErrIsNil
	if v != nil && !ok {
		err = ErrIsNil
		return
	}
	var buf []byte
	// 从lru中获取到可用数据
	if ok {
		buf, _ = v.([]byte)
	}
	// 从lru中数据不存在
	// 有可能数据未过期但lru空间较小，因此被删除
	if len(buf) == 0 {
		buf, err = l2.slowCache.Get(key)
		if err != nil {
			return
		}
	}
	fn := l2.unmarshal
	if fn == nil {
		fn = json.Unmarshal
	}
	err = fn(buf, result)
	if err != nil {
		return
	}
	return
}

// Set converts the value to bytes, then set it to lru cache and slow cache
func (l2 *L2Cache) Set(key string, value interface{}) (err error) {
	key = l2.getKey(key)
	fn := l2.marshal
	if fn == nil {
		fn = json.Marshal
	}
	buf, err := fn(value)
	if err != nil {
		return
	}
	// 先设置较慢的缓存
	err = l2.slowCache.Set(key, buf, l2.ttl)
	if err != nil {
		return
	}
	l2.ttlCache.Add(key, buf)
	return
}
