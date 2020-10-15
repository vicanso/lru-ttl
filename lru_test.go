/*
Copyright 2013 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lruttl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simpleStruct struct {
	int
	string
}

type complexStruct struct {
	int
	simpleStruct
}

var getTests = []struct {
	name       string
	keyToAdd   interface{}
	keyToGet   interface{}
	expectedOk bool
}{
	{"string_hit", "myKey", "myKey", true},
	{"string_miss", "myKey", "nonsense", false},
	{"simple_struct_hit", simpleStruct{1, "two"}, simpleStruct{1, "two"}, true},
	{"simple_struct_miss", simpleStruct{1, "two"}, simpleStruct{0, "noway"}, false},
	{"complex_struct_hit", complexStruct{1, simpleStruct{2, "three"}},
		complexStruct{1, simpleStruct{2, "three"}}, true},
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	for _, tt := range getTests {
		lru := NewLRU(0)
		lru.cache = nil
		lru.Add(tt.keyToAdd, 1234)
		lru.Add(tt.keyToAdd, 1234)
		val, ok := lru.Get(tt.keyToGet)
		assert.Equal(tt.expectedOk, ok)
		if ok {
			assert.Equal(1234, val)
		}
	}
}

func TestRemove(t *testing.T) {
	assert := assert.New(t)
	lru := NewLRU(0)
	lru.Add("myKey", 1234)
	val, ok := lru.Get("myKey")
	assert.True(ok)
	assert.Equal(1234, val)
	assert.Equal(1, lru.Len())

	lru.Remove("myKey")
	_, ok = lru.Get("myKey")
	assert.False(ok)
}

func TestEvict(t *testing.T) {
	assert := assert.New(t)
	evictedKeys := make([]Key, 0)
	onEvictedFun := func(key Key, value interface{}) {
		evictedKeys = append(evictedKeys, key)
	}

	lru := NewLRU(20)
	lru.OnEvicted = onEvictedFun
	for i := 0; i < 22; i++ {
		lru.Add(fmt.Sprintf("myKey%d", i), 1234)
	}

	assert.Equal(2, len(evictedKeys))
	assert.Equal(Key("myKey0"), evictedKeys[0])
	assert.Equal(Key("myKey1"), evictedKeys[1])
	lru.Clear()
	assert.Equal(0, lru.Len())
}
