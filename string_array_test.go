package ravendb

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArraySubtract(t *testing.T) {
	var tests = []struct {
		a1, a2 []string
		exp    []string
	}{
		{nil, nil, nil},
		{[]string{}, nil, nil},
		{[]string{"a"}, nil, []string{"a"}},
		{[]string{"a", "b"}, []string{"a"}, []string{"b"}},
	}
	for _, test := range tests {
		got := StringArraySubtract(test.a1, test.a2)
		sort.Strings(got)
		if !StringArrayEq(test.exp, got) {
			t.Fatalf("got: %#v, exp: %#v", got, test.exp)
		}
	}
}

func TestStringArrayContains(t *testing.T) {
	var tests = []struct {
		a   []string
		s   string
		exp bool
	}{
		{nil, "a", false},
		{[]string{}, "a", false},
		{[]string{"a"}, "a", true},
		{[]string{"b", "a"}, "a", true},
		{[]string{"b", "a", "c", "d"}, "a", true},
		{[]string{"a"}, "A", false},
		{[]string{"a", "a"}, "a", true},
		{[]string{"a", ""}, "", true},
		{[]string{}, "", false},
	}
	for _, test := range tests {
		got := StringArrayContains(test.a, test.s)
		assert.Equal(t, test.exp, got)
	}
}
