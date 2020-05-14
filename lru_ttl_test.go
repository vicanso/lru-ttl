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

	key1 := "foo1"
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
}
