package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicInteger(t *testing.T) {
	var i AtomicInteger
	v := i.Get()
	assert.Equal(t, 0, v)
	i.Set(5)
	v = i.Get()
	assert.Equal(t, 5, v)
	v = i.DecrementAndGet()
	assert.Equal(t, 4, v)
	v = i.DecrementAndGet()
	assert.Equal(t, 3, v)

	didSet := i.CompareAndSet(4, 8)
	assert.False(t, didSet)
	didSet = i.CompareAndSet(3, 8)
	assert.True(t, didSet)
	v = i.Get()
	assert.Equal(t, 8, v)
	v = i.IncrementAndGet()
	assert.Equal(t, 9, v)
}
