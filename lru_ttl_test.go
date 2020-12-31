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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)
	cache := New(1, 300*time.Millisecond)
	key := "foo"
	value := []byte("bar")

	cache.Add(key, value)
	assert.Equal(1, cache.Len())
	data, ok := cache.Get(key)
	assert.True(ok)
	assert.Equal(value, data)

	key1 := 123
	value1 := []byte("bar1")
	cache.Add(key1, value1, 100*time.Millisecond)
	_, ok = cache.Get(key)
	assert.False(ok)
	assert.Equal(1, cache.Len())

	_, ok = cache.Get(key1)
	assert.True(ok)
	time.Sleep(500 * time.Millisecond)
	_, ok = cache.Get(key1)
	assert.False(ok)

	cache.Add(key, value)
	assert.Equal(1, cache.Len())
	cache.Remove(key)
	assert.Equal(0, cache.Len())

	max := 10
	cache = New(max, time.Minute)
	for i := 0; i < 2*max; i++ {
		cache.Add(i, i)
	}
	assert.Equal(max, cache.Len())
	assert.Equal(max, len(cache.Keys()))
	cache.ForEach(func(key Key, v interface{}) {
		index, ok := key.(int)
		assert.True(ok)
		value, ok := v.(int)
		assert.True(ok)
		assert.Equal(index, value)
	})
}

func TestCacheOnEvicted(t *testing.T) {
	assert := assert.New(t)

	cache := New(1, 300*time.Millisecond)

	evictedCount := 0
	evictedKeys := []string{
		"test1",
		"test2",
	}
	cache.SetOnEvicted(func(key Key, value interface{}) {
		assert.Equal(key, evictedKeys[evictedCount])
		evictedCount++
	})

	cache.Add("test1", "value1")
	cache.Add("test2", "value2")

	time.Sleep(500 * time.Millisecond)
	// 此时再次获取，该key已过期，也会触发evicted
	cache.Get("test2")

	assert.Equal(2, evictedCount)

}
