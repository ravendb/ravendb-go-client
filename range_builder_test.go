package ravendb

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestRangeBuilder(t *testing.T) {
	{
		b := NewRangeBuilder("foo")
		b = b.IsLessThan(3)
		b = b.IsLessThan(4)
		assert.Error(t, b.Err())
		assert.True(t, strings.Contains(b.Err().Error(), "Less bound was already set"))
	}

	{
		b := NewRangeBuilder("foo")
		b = b.IsLessThan(3)
		b = b.IsLessThanOrEqualTo(18)
		assert.Error(t, b.Err())
		assert.True(t, strings.Contains(b.Err().Error(), "Less bound was already set"))
	}

	{
		b := NewRangeBuilder("foo")
		b = b.IsGreaterThan(3)
		b = b.IsGreaterThan(18)
		assert.Error(t, b.Err())
		assert.True(t, strings.Contains(b.Err().Error(), "Greater bound was already set"))
	}

	{
		b := NewRangeBuilder("foo")
		b = b.IsGreaterThanOrEqualTo(3)
		b = b.IsGreaterThan(18)
		assert.Error(t, b.Err())
		assert.True(t, strings.Contains(b.Err().Error(), "Greater bound was already set"))
	}

	{
		b := NewRangeBuilder("foo")
		addQueryParameter := func(interface{}) string {
			return ""
		}
		s, err := b.GetStringRepresentation(addQueryParameter)
		assert.Equal(t, s, "")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "Bounds were not set"))
	}
}
