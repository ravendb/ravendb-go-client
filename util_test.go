package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceArrayContains(t *testing.T) {
	var tests = []struct {
		a   []interface{}
		v   interface{}
		exp bool
	}{
		{nil, "a", false},
		{[]interface{}{}, "a", false},
		{[]interface{}{""}, "a", false},
		{[]interface{}{"b", "a"}, "a", true},
		{[]interface{}{"b", "a", "c", "d"}, "a", true},
		{[]interface{}{"a"}, "A", false},
		{[]interface{}{"a", "a"}, "a", true},
		{[]interface{}{"a", ""}, "", true},
		{[]interface{}{}, "", false},
	}
	for idx, test := range tests {
		got := interfaceArrayContains(test.a, test.v)
		assert.Equal(t, test.exp, got, "a: %v, v: %v, idx: %d", test.a, test.v, idx)
	}
}
