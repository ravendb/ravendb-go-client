package ravendb

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	{
		var tt time.Time
		d, err := json.Marshal(Time(tt))
		assert.NoError(t, err)
		s := string(d)
		assert.Equal(t, `"0001-01-01T00:00:00.0000000Z"`, s)
	}

	{
		tt := time.Now()
		d, err := json.Marshal(Time(tt))
		assert.NoError(t, err)
		s := string(d)
		parts := strings.Split(s, ".")
		fracPart := parts[len(parts)-1]
		assert.Equal(t, 9, len(fracPart)) // 8 = 7 digits + Z + "
		assert.True(t, strings.HasSuffix(s, `Z"`))
	}

}
