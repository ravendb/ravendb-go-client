package ravendb

import "strings"

// TODO: make it more efficient by modifying the array in-place
func StringArrayRemove(pa *[]string, s string) bool {
	if len(*pa) == 0 {
		return false
	}
	var res []string
	removed := false
	for _, s2 := range *pa {
		if s2 == s {
			removed = true
			continue
		}
		res = append(res, s2)
	}
	*pa = res
	return removed
}

// TODO: make it more efficient by modifying the array in-place
func StringArrayRemoveCustomCompare(pa *[]string, s string, cmp func(string, string) bool) bool {
	if len(*pa) == 0 {
		return false
	}
	var res []string
	removed := false
	for _, s2 := range *pa {
		if cmp(s2, s) {
			removed = true
			continue
		}
		res = append(res, s2)
	}
	*pa = res
	return removed
}

func StringArrayCopy(a []string) []string {
	n := len(a)
	if n == 0 {
		return nil
	}
	res := make([]string, n, n)
	for i := 0; i < n; i++ {
		res[i] = a[i]
	}
	return res
}

// return a1 - a2
func StringArraySubtract(a1, a2 []string) []string {
	if len(a2) == 0 {
		return a1
	}
	if len(a1) == 0 {
		return nil
	}
	diff := make(map[string]struct{})
	for _, k := range a1 {
		diff[k] = struct{}{}
	}
	for _, k := range a2 {
		delete(diff, k)
	}
	if len(diff) == 0 {
		return nil
	}
	// TODO: pre-allocate
	var res []string
	for k := range diff {
		res = append(res, k)
	}
	return res
}

func StringArrayContains(a []string, s string) bool {
	for _, el := range a {
		if el == s {
			return true
		}
	}
	return false
}

func StringArrayEq(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	if len(a1) == 0 {
		return true
	}
	// TODO: could be faster if used map
	for _, s := range a1 {
		if !StringArrayContains(a2, s) {
			return false
		}
	}
	return true
}

const (
	unlikelySep = "\x02\x01\x03"
)

// equivalent of Java's containsSequence http://joel-costigliola.github.io/assertj/core/api/org/assertj/core/api/ListAssert.html#containsSequence(ELEMENT...)
// checks if a1 contains sub-sequence a2
func StringArrayContainsSequence(a1, a2 []string) bool {
	// TODO: technically it's possible for this to have false positive
	// but it's very unlikely
	s1 := strings.Join(a1, unlikelySep)
	s2 := strings.Join(a2, unlikelySep)
	return strings.Contains(s1, s2)
}

func StringArrayContainsExactly(a1, a2 []string) bool {
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

func StringArrayReverse(a []string) {
	n := len(a)
	for i := 0; i < n/2; i++ {
		a[i], a[n-1-i] = a[n-1-i], a[i]
	}
}
