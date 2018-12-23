package ravendb

import (
	"sort"
	"strings"
)

const (
	unlikelySep = "\x02\x01\x03"
)

func stringArrayRemoveNoCase(a []string, s string) []string {
	n := len(a)
	if n == 0 {
		return a
	}
	var toRemove []int
	for i, s1 := range a {
		if strings.EqualFold(s1, s) {
			toRemove = append(toRemove, i)
		}
	}
	return stringArrayRemoveAtIndexes(a, toRemove)
}

func stringArrayRemove(pa *[]string, s string) bool {
	a := *pa
	n := len(a)
	if n == 0 {
		return false
	}

	var toRemove []int
	for i, s1 := range a {
		if s1 == s {
			toRemove = append(toRemove, i)
		}
	}
	if len(toRemove) == 0 {
		return false
	}
	*pa = stringArrayRemoveAtIndexes(a, toRemove)
	return true
}

func stringArrayCopy(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	return append([]string{}, a...)
}

// return a1 - a2
func stringArraySubtract(a1, a2 []string) []string {
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

func stringArrayContains(a []string, s string) bool {
	for _, el := range a {
		if el == s {
			return true
		}
	}
	return false
}

// stringArrayContainsNoCase returns true if a contains s using case-insensitive
// comparison
func stringArrayContainsNoCase(a []string, s string) bool {
	for _, el := range a {
		if strings.EqualFold(el, s) {
			return true
		}
	}
	return false
}

func stringArrayRemoveAtIndexes(a []string, toRemove []int) []string {
	if len(toRemove) == 0 {
		return a
	}
	// remove from the end so that index in toRemove isn't invalidated
	// by changing the array
	n := len(a)
	lastIdx := n - 1
	for i := len(toRemove) - 1; i >= 0; i-- {
		idx := toRemove[i]
		// remove by replacing with element from end of array
		a[idx] = a[lastIdx]
		lastIdx--
	}
	return a[:n-len(toRemove)]
}

// stringArrayRemoveDuplicates removes duplicate strings from a
func stringArrayRemoveDuplicates(a []string) []string {
	n := len(a)
	if n < 2 {
		return a
	}
	sort.Strings(a)
	var toRemove []int
	prev := a[0]
	for i := 1; i < n; i++ {
		if a[i] == prev {
			toRemove = append(toRemove, i)
			continue
		}
		prev = a[i]
	}
	return stringArrayRemoveAtIndexes(a, toRemove)
}

// stringArrayRemoveDuplicatesNoCase removes duplicate strings from a, ignoring case
func stringArrayRemoveDuplicatesNoCase(a []string) []string {
	n := len(a)
	if n < 2 {
		return a
	}
	sort.Slice(a, func(i, j int) bool {
		s1 := strings.ToLower(a[i])
		s2 := strings.ToLower(a[j])
		return s1 < s2
	})
	var toRemove []int
	prev := strings.ToLower(a[0])
	for i := 1; i < n; i++ {
		s := strings.ToLower(a[i])
		if s == prev {
			toRemove = append(toRemove, i)
			continue
		}
		prev = s
	}
	return stringArrayRemoveAtIndexes(a, toRemove)
}
