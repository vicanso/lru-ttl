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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUTTL(t *testing.T) {
	assert := assert.New(t)
	cache := New(1, 300*time.Millisecond)
	key := "foo"
	value := []byte("bar")

	cache.Add(key, value)
	assert.Equal(1, cache.Len())
	data, ok := cache.Get(key)
	assert.True(ok)
	assert.Equal(value, data)

	data, ok = cache.Peek(key)
	assert.True(ok)
	assert.Equal(value, data)

	// 添加新的数据
	key1 := 123
	value1 := []byte("bar1")
	cache.Add(key1, value1, 100*time.Millisecond)
	// 由于长度限制为1，因此原来的数据被清除
	_, ok = cache.Get(key)
	assert.False(ok)
	assert.Equal(1, cache.Len())

	_, ok = cache.Get(key1)
	assert.True(ok)
	time.Sleep(500 * time.Millisecond)
	// 数据过期，peek不清除，数据有返回
	item, ok := cache.Peek(key1)
	assert.False(ok)
	assert.NotNil(item)
	// 数据过期，get时清除，数据有返回
	item, ok = cache.Get(key1)
	assert.False(ok)
	assert.NotNil(item)

	// 添加数据
	cache.Add(key, value)
	assert.Equal(1, cache.Len())
	// 清除数据
	cache.Remove(key)
	assert.Equal(0, cache.Len())

	max := 10
	cache = New(max, time.Minute)
	for i := 0; i < 2*max; i++ {
		cache.Add(i, i)
	}
	assert.Equal(max, cache.Len())
	assert.Equal(max, len(cache.Keys()))
}

func TestLRUTTLGetTTL(t *testing.T) {
	assert := assert.New(t)

	cache := New(10, time.Minute)
	key := "test"
	itemTTL := 100 * time.Millisecond

	cache.Add(key, "a", itemTTL)
	ttl := cache.TTL(key)
	assert.True(ttl > (itemTTL/2) && ttl <= itemTTL)
	// 等待过期
	time.Sleep(2 * itemTTL)
	ttl = cache.TTL(key)
	assert.Equal(time.Duration(-1), ttl)
}

func TestLRUTTLParallelAdd(t *testing.T) {
	assert := assert.New(t)
	cache := New(10, time.Second)
	key1 := "1"
	value1 := "value1"
	key2 := "2"
	value2 := "value2"
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		if i%2 == 0 {
			go func() {
				cache.Add(key1, value1)
				wg.Done()
			}()
		} else {
			go func() {
				cache.Add(key2, value2)
				wg.Done()
			}()
		}
	}
	wg.Wait()
	value, ok := cache.Get(key1)
	assert.True(ok)
	assert.Equal(value1, value)

	value, ok = cache.Get(key2)
	assert.True(ok)
	assert.Equal(value2, value)
}

func TestLRUTTLParallelGet(t *testing.T) {
	assert := assert.New(t)
	cache := New(10, time.Second)
	key1 := "1"
	value1 := "value1"
	cache.Add(key1, value1)
	key2 := "2"
	value2 := "value2"
	cache.Add(key2, value2)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		if i%2 == 0 {
			go func() {
				value, ok := cache.Get(key1)
				assert.True(ok)
				assert.Equal(value1, value)
				wg.Done()
			}()
		} else {
			go func() {
				value, ok := cache.Get(key2)
				assert.True(ok)
				assert.Equal(value2, value)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func TestLRUTTLParallelPeek(t *testing.T) {
	assert := assert.New(t)
	cache := New(10, time.Second)
	key1 := "1"
	value1 := "value1"
	cache.Add(key1, value1)
	key2 := "2"
	value2 := "value2"
	cache.Add(key2, value2)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		if i%2 == 0 {
			go func() {
				value, ok := cache.Peek(key1)
				assert.True(ok)
				assert.Equal(value1, value)
				wg.Done()
			}()
		} else {
			go func() {
				value, ok := cache.Peek(key2)
				assert.True(ok)
				assert.Equal(value2, value)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func TestLRUTTLCacheOnEvicted(t *testing.T) {
	assert := assert.New(t)

	evictedCount := 0
	evictedKeys := []string{
		"test1",
		"test2",
	}
	cache := New(1, 300*time.Millisecond, CacheEvictedOption(func(key Key, value interface{}) {
		assert.Equal(key, evictedKeys[evictedCount])
		evictedCount++
	}))

	cache.Add("test1", "value1")
	cache.Add("test2", "value2")

	time.Sleep(500 * time.Millisecond)
	// 此时再次获取，该key已过期，也会触发evicted
	cache.Get("test2")

	assert.Equal(2, evictedCount)

}

func TestGetAndPeekBytes(t *testing.T) {
	assert := assert.New(t)

	cache := New(10, time.Minute)
	key := "key"

	cache.Add(key, []byte("abc"))

	data, ok := cache.GetBytes(key)
	assert.True(ok)
	assert.Equal([]byte("abc"), data)

	data, ok = cache.PeekBytes(key)
	assert.True(ok)
	assert.Equal([]byte("abc"), data)
}
