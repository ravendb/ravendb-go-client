package ravendb

import (
	"sort"
	"testing"
)

func stringInArray(a []string, s string) bool {
	for _, s2 := range a {
		if s == s2 {
			return true
		}
	}
	return false
}

func stringArrayEq(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	if len(a1) == 0 {
		return true
	}
	// TODO: could be faster if used map
	for _, s := range a1 {
		if !stringInArray(a2, s) {
			return false
		}
	}
	return true
}

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
		got := stringArraySubtract(test.a1, test.a2)
		sort.Strings(got)
		if !stringArrayEq(test.exp, got) {
			t.Fatalf("got: %#v, exp: %#v", got, test.exp)
		}
	}
}
