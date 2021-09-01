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
	"context"
	"encoding/json"
	"errors"
	"time"
)

type SlowCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Del(ctx context.Context, key string) (int64, error)
}

// L2CacheOption l2cache option
type L2CacheOption func(c *L2Cache)

type L2CacheMarshal func(v interface{}) ([]byte, error)
type L2CacheUnmarshal func(data []byte, v interface{}) error

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
	marshal L2CacheMarshal
	// unmarshal is custom unmarshal function.
	// It will be json.Unmarshal if not set
	unmarshal L2CacheUnmarshal
}

// ErrIsNil is the error of nil cache
var ErrIsNil = errors.New("cache is nil")

// ErrKeyIsNil is the error of nil key
var ErrKeyIsNil = errors.New("key is nil")

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

// L2CacheMarshalOption sets custom marshal function for l2cache
func L2CacheMarshalOption(fn L2CacheMarshal) L2CacheOption {
	return func(c *L2Cache) {
		c.marshal = fn
	}
}

// L2CacheUnmarshalOption sets custom unmarshal function for l2cache
func L2CacheUnmarshalOption(fn L2CacheUnmarshal) L2CacheOption {
	return func(c *L2Cache) {
		c.unmarshal = fn
	}
}

// L2CachePrefixOption sets prefix for l2cache
func L2CachePrefixOption(prefix string) L2CacheOption {
	return func(c *L2Cache) {
		c.prefix = prefix
	}
}

func (l2 *L2Cache) getKey(key string) (string, error) {
	if key == "" {
		return "", ErrKeyIsNil
	}
	return l2.prefix + key, nil
}

// TTL returns the ttl for key
func (l2 *L2Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	key, err := l2.getKey(key)
	if err != nil {
		return 0, err
	}
	d := l2.ttlCache.TTL(key)
	// 小于0的表示不存在
	// 由于lru有大小限制，不存在时则从slow cache获取
	if d >= 0 {
		return d, nil
	}
	return l2.slowCache.TTL(ctx, key)
}

// GetBytes gets data from lur cache first, if not exists,
// then gets the data from slow cache.
func (l2 *L2Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	key, err := l2.getKey(key)
	if err != nil {
		return nil, err
	}
	v, ok := l2.ttlCache.Get(key)
	var buf []byte
	// 获取成功，而数据不为nil
	// 如果ok为false时，数据也可能不为空（已过期）
	if ok && v != nil {
		buf, _ = v.([]byte)
	}
	// 从lru中获取到可用数据
	// 从lru中数据不存在（数据不存在或过期都有可能）
	// 有可能数据未过期但lru空间较小，因此被删除
	// 也有可能lru中数据过期但 slow cache中数据已更新
	if len(buf) == 0 {
		b, err := l2.slowCache.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		buf = b
		// 成功从slowcache获取缓存，则将数据设置回lru ttl
		if len(buf) != 0 {
			ttl, _ := l2.slowCache.TTL(ctx, key)
			if ttl != 0 {
				l2.ttlCache.Add(key, buf, ttl)
			}
		}
	}
	return buf, nil
}

// SetBytes sets data to lru cache and slow cache
func (l2 *L2Cache) SetBytes(ctx context.Context, key string, value []byte, ttl ...time.Duration) error {
	t := l2.ttl
	if len(ttl) != 0 && ttl[0] != 0 {
		t = ttl[0]
	}
	// 先设置较慢的缓存
	err := l2.slowCache.Set(ctx, key, value, t)
	if err != nil {
		return err
	}
	l2.ttlCache.Add(key, value, t)
	return nil
}

// Get gets data from lru cache first, if not exists,
// then gets the data from slow cache.
// Use unmarshal function coverts the data to result
func (l2 *L2Cache) Get(ctx context.Context, key string, result interface{}) error {
	buf, err := l2.GetBytes(ctx, key)
	if err != nil {
		return err
	}

	fn := l2.unmarshal
	if fn == nil {
		fn = json.Unmarshal
	}
	err = fn(buf, result)
	if err != nil {
		return err
	}
	return nil
}

// Set converts the value to bytes, then sets it to lru cache and slow cache
func (l2 *L2Cache) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	key, err := l2.getKey(key)
	if err != nil {
		return err
	}
	fn := l2.marshal
	if fn == nil {
		fn = json.Marshal
	}
	buf, err := fn(value)
	if err != nil {
		return err
	}
	return l2.SetBytes(ctx, key, buf, ttl...)
}

// Del deletes data from lru cache and slow cache
func (l2 *L2Cache) Del(ctx context.Context, key string) (int64, error) {
	key, err := l2.getKey(key)
	if err != nil {
		return 0, err
	}
	l2.ttlCache.Remove(key)
	return l2.slowCache.Del(ctx, key)
}
