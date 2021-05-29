// Copyright 2021 tree xie
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

import "time"

type ringCache struct {
	// lru cache list
	lruCaches []*Cache
	size      uint64
}

type RingCacheParams struct {
	// ring size
	Size int
	// max entries
	MaxEntries int
	// default ttl
	DefaultTTL time.Duration
}

// NewRing returns a new ring cache
func NewRing(params RingCacheParams, opts ...CacheOption) *ringCache {
	if params.DefaultTTL <= 0 || params.Size <= 0 || params.MaxEntries <= params.Size {
		panic("default ttl, size and max entries must be gt 0")
	}
	lruCacheCount := params.MaxEntries/params.Size + 1
	lruCaches := make([]*Cache, params.Size)
	for i := 0; i < params.Size; i++ {
		lruCaches[i] = New(lruCacheCount, params.DefaultTTL, opts...)
	}
	return &ringCache{
		lruCaches: lruCaches,
		size:      uint64(params.Size),
	}
}

// Get returns the lru ttl cache by key
func (rc *ringCache) Get(key string) *Cache {
	value := MemHashString(key)
	index := int(value % rc.size)
	return rc.lruCaches[index]
}
