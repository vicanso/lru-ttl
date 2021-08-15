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
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testSlowCache struct {
	data map[string][]byte
}

const slowCacheTTL = 101 * time.Millisecond

func (sc *testSlowCache) Get(_ context.Context, key string) ([]byte, error) {
	buf, ok := sc.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	time.Sleep(time.Second)
	return buf, nil
}

func (sc *testSlowCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	sc.data[key] = value
	return nil
}
func (sc *testSlowCache) TTL(_ context.Context, key string) (time.Duration, error) {
	return slowCacheTTL, nil
}

type testData struct {
	Name string `json:"name,omitempty"`
}

func TestL2Cache(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

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

	err := l2.Get(ctx, key, &data)
	assert.NotNil(err)
	assert.Equal("not found", err.Error())

	err = l2.Set(ctx, key, &testData{
		Name: name,
	})
	assert.Nil(err)

	// 成功获取
	err = l2.Get(ctx, key, &data)
	assert.Nil(err)
	assert.Equal(name, data.Name)

	// 由于lru的大小令为1，因此会导致lru中清除了原有的key
	err = l2.Set(ctx, "ab", &testData{})
	assert.Nil(err)

	// 从slow cache中获取数据并更新lru缓存
	err = l2.Get(ctx, key, &data)
	assert.Nil(err)
	assert.Equal(name, data.Name)

	// 从lru获取缓存（时间较快)
	start := time.Now()
	err = l2.Get(ctx, key, &data)
	assert.Nil(err)
	assert.Equal(name, data.Name)
	// 从lru获取耗时少于10ms
	assert.True(time.Since(start) < 10*time.Millisecond)

	err = l2.Set(ctx, key, &map[string]string{
		"name": "newName",
	})
	assert.Nil(err)
	m := make(map[string]string)
	err = l2.Get(ctx, key, &m)
	assert.Nil(err)
	assert.Equal("newName", m["name"])
}

func TestL2CacheTTL(t *testing.T) {
	assert := assert.New(t)
	sc := testSlowCache{
		data: make(map[string][]byte),
	}
	ctx := context.Background()
	l2 := NewL2Cache(&sc, 10, 10*time.Second)
	key := "test"
	err := l2.Set(ctx, key, "value", 2*time.Second)
	assert.Nil(err)

	ttl, err := l2.TTL(ctx, key)
	assert.Nil(err)
	// 从lru中获取
	assert.True(ttl > time.Second && ttl <= 2*time.Second)

	l2.ttlCache.Remove(key)

	ttl, err = l2.TTL(ctx, key)
	assert.Nil(err)
	// 从slow cache中获取，slow cache获取ttl为固定值
	assert.Equal(slowCacheTTL, ttl)
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
