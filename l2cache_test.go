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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testSlowCache struct {
	data map[string][]byte
}

func (sc *testSlowCache) Get(key string) ([]byte, error) {
	buf, ok := sc.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return buf, nil
}

func (sc *testSlowCache) Set(key string, value []byte, ttl time.Duration) error {
	sc.data[key] = value
	return nil
}

type testData struct {
	Name string `json:"name,omitempty"`
}

func TestL2Cache(t *testing.T) {
	assert := assert.New(t)

	sc := testSlowCache{
		data: make(map[string][]byte),
	}
	opts := []L2CacheOption{
		L2CacheMarshalOption(json.Marshal),
		L2CacheUnmarshalOption(json.Unmarshal),
		L2CachePrefixOption("prefix:"),
	}
	l2 := NewL2Cache(&sc, 1, time.Second, opts...)

	assert.Equal("prefix:1", l2.getKey("1"))

	key := "abcd"
	name := "test"
	data := testData{}

	err := l2.Get(key, &data)
	assert.NotNil(err)
	assert.Equal("not found", err.Error())

	err = l2.Set(key, &testData{
		Name: name,
	})
	assert.Nil(err)

	// 成功获取
	err = l2.Get(key, &data)
	assert.Nil(err)
	assert.Equal(name, data.Name)

	// 由于lru的大小令为1，因此会导致lru中清除了原有的key
	err = l2.Set("ab", &testData{})
	assert.Nil(err)

	// 从slow cache中获取
	err = l2.Get(key, &data)
	assert.Nil(err)
	assert.Equal(name, data.Name)

	err = l2.Set(key, &map[string]string{
		"name": "newName",
	})
	assert.Nil(err)
	m := make(map[string]string)
	err = l2.Get(key, &m)
	assert.Nil(err)
	assert.Equal("newName", m["name"])
}

func TestBufferMarshalUnmarshal(t *testing.T) {
	assert := assert.New(t)
	buf := bytes.NewBufferString("abc")
	result, err := BufferMarshal(buf)
	assert.Nil(err)
	assert.Equal(buf.Bytes(), result)

	newBuf := &bytes.Buffer{}
	err = BufferUnmarshal(result, newBuf)
	assert.Nil(err)
	assert.Equal(buf, newBuf)
}
