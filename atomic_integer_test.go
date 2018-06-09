package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicInteger(t *testing.T) {
	var i AtomicInteger
	v := i.get()
	assert.Equal(t, 0, v)
	i.set(5)
	v = i.get()
	assert.Equal(t, 5, v)
	v = i.decrementAndGet()
	assert.Equal(t, 4, v)
	v = i.decrementAndGet()
	assert.Equal(t, 3, v)

	didSet := i.compareAndSet(4, 8)
	assert.False(t, didSet)
	didSet = i.compareAndSet(3, 8)
	assert.True(t, didSet)
	v = i.get()
	assert.Equal(t, 8, v)
	v = i.incrementAndGet()
	assert.Equal(t, 9, v)
}
