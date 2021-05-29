package lruttl

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRingCache(t *testing.T) {
	assert := assert.New(t)

	ringCache := NewRing(RingCacheParams{
		Size:       10,
		MaxEntries: 1000,
		DefaultTTL: time.Minute,
	})

	key := "test"
	c := ringCache.Get(key)
	assert.NotNil(c)
	assert.Equal(c, ringCache.Get(key))

	for i := 0; i < 1000; i++ {
		str := strconv.Itoa(int(time.Now().UnixNano()))
		assert.NotNil(ringCache.Get(str))
	}
}
