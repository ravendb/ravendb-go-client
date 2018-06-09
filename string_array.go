package ravendb

// TODO: make it more efficient by modifying the array in-place
func removeStringFromArray(pa *[]string, s string) {
	var res []string
	for _, s2 := range *pa {
		if s2 == s {
			continue
		}
		res = append(res, s2)
	}
	*pa = res
}

func stringArrayCopy(a []string) []string {
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
