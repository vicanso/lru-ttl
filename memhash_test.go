package lruttl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemhash(t *testing.T) {
	assert := assert.New(t)
	value := MemHashString("abc")

	assert.Equal(value, MemHashString("abc"))
	assert.Equal(value, MemHashString("abc"))
	assert.NotEqual(value, MemHashString("bcd"))
}
