package ravendb

import "sort"

func IntArrayHasDuplicates(a []int) bool {
	if len(a) == 0 {
		return false
	}
	sort.Ints(a)
	prev := a[0]
	a = a[1:]
	for _, el := range a {
		if el == prev {
			return true
		}
		prev = el
	}
	return false
}

func IntArrayContains(a []int, n int) bool {
	for _, el := range a {
		if el == n {
			return true
		}
	}
	return false
}
