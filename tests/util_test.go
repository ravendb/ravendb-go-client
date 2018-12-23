package tests

import (
	"sort"
	"strings"
)

const (
	unlikelySep = "\x02\x01\x03"
)

func stringArrayCopy(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	return append([]string{}, a...)
}

func stringArrayContains(a []string, s string) bool {
	for _, el := range a {
		if el == s {
			return true
		}
	}
	return false
}

// equivalent of Java's containsSequence http://joel-costigliola.github.io/assertj/core/api/org/assertj/core/api/ListAssert.html#containsSequence(ELEMENT...)
// checks if a1 contains sub-sequence a2
func stringArrayContainsSequence(a1, a2 []string) bool {
	// TODO: technically it's possible for this to have false positive
	// but it's very unlikely
	s1 := strings.Join(a1, unlikelySep)
	s2 := strings.Join(a2, unlikelySep)
	return strings.Contains(s1, s2)
}

func stringArrayContainsExactly(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, s := range a1 {
		if s != a2[i] {
			return false
		}
	}
	return true
}

// stringArrayEq returns true if arrays have the same content, ignoring order
func stringArrayEq(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	if len(a1) == 0 {
		return true
	}
	a1c := stringArrayCopy(a1)
	a2c := stringArrayCopy(a2)
	sort.Strings(a1c)
	sort.Strings(a2c)
	for i, s := range a1c {
		if s != a2c[i] {
			return false
		}
	}
	return true
}

func stringArrayReverse(a []string) {
	n := len(a)
	for i := 0; i < n/2; i++ {
		a[i], a[n-1-i] = a[n-1-i], a[i]
	}
}
