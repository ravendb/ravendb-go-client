package ravendb

func IntArrayContains(a []int, n int) bool {
	for _, el := range a {
		if el == n {
			return true
		}
	}
	return false
}
