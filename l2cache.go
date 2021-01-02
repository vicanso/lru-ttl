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
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

type SlowCache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
}

// L2Cache l2cache, use lru cache for the first cache, and slow cache for the second cache
// lru cache should be set max entries for less memory usage but faster,
// slow cache is slower but more space
type L2Cache struct {
	prefix    string
	ttl       time.Duration
	ttlCache  *Cache
	slowCache SlowCache
	marshal   func(v interface{}) ([]byte, error)
	unmarshal func(data []byte, v interface{}) error
}

var ErrIsNil = errors.New("cache is nil")
var ErrInvalidType = errors.New("invalid type")

// BufferMarshal buffer marshal
func BufferMarshal(v interface{}) ([]byte, error) {
	buf, ok := v.(*bytes.Buffer)
	if !ok {
		return nil, ErrInvalidType
	}
	return buf.Bytes(), nil
}

// BufferUnmarshal buffer unmarshal
func BufferUnmarshal(data []byte, v interface{}) error {
	buf, ok := v.(*bytes.Buffer)
	if !ok {
		return ErrInvalidType
	}
	buf.Write(data)
	return nil
}

// NewL2Cache create a new l2cache
func NewL2Cache(slowCache SlowCache, maxEntries int, defaultTTL time.Duration) *L2Cache {
	return &L2Cache{
		ttl:       defaultTTL,
		ttlCache:  New(maxEntries, defaultTTL),
		slowCache: slowCache,
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}
}

// SetMarshal set marshal function, default is json.Marshal
func (l2 *L2Cache) SetMarshal(fn func(v interface{}) ([]byte, error)) {
	l2.marshal = fn
}

// SetUnmarshal set unmarshal function, default is json.Un
func (l2 *L2Cache) SetUnmarshal(fn func(data []byte, v interface{}) error) {
	l2.unmarshal = fn
}

// SetPrefix set prefix for l2cache key
func (l2 *L2Cache) SetPrefix(prefix string) {
	l2.prefix = prefix
}

func (l2 *L2Cache) getKey(key string) string {
	return l2.prefix + key
}

// Get get value from cache
func (l2 *L2Cache) Get(key string, value interface{}) (err error) {
	key = l2.getKey(key)
	v, ok := l2.ttlCache.Get(key)
	// 如果获取到的不为空，但是ok为false
	// 表示数据已过期
	if v != nil && !ok {
		err = ErrIsNil
		return
	}
	var buf []byte
	if ok {
		buf, _ = v.([]byte)
	}
	if len(buf) == 0 {
		buf, err = l2.slowCache.Get(key)
		if err != nil {
			return
		}
	}
	err = l2.unmarshal(buf, value)
	if err != nil {
		return
	}
	return
}

// Set set value to cache
func (l2 *L2Cache) Set(key string, value interface{}) (err error) {
	key = l2.getKey(key)
	buf, err := l2.marshal(value)
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
