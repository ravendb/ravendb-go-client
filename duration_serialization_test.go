package ravendb

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCanSerializeDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d   time.Duration
		exp string
	}{
		{time.Hour*24*5 + time.Hour*2, `"5.02:00:00"`},
		{time.Millisecond * 5, `"00:00:00.0050000"`},
	}
	for _, test := range tests {
		d2 := Duration(test.d)
		d, err := json.Marshal(d2)
		assert.NoError(t, err)
		got := string(d)
		assert.Equal(t, test.exp, got)
	}
}

func TestCanDeserializeDuration(t *testing.T) {
	tests := []struct {
		s   string
		exp time.Duration
	}{
		{`"5.02:00:00"`, time.Hour*24*5 + time.Hour*2},
		{`"00:00:00.0050000"`, time.Millisecond * 5},
		{`"00:00:00.005000"`, time.Millisecond * 5},
		{`"00:00:00.00500"`, time.Millisecond * 5},
		{`"00:00:00.1"`, time.Millisecond * 100},
	}
	for _, test := range tests {
		var got Duration
		d := []byte(test.s)
		err := json.Unmarshal(d, &got)
		assert.NoError(t, err)
		exp := Duration(test.exp)
		assert.Equal(t, exp, got)
	}
}
