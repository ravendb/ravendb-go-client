package ravendb

func dupMapStringString(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	ret := make(map[string]string)
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func dupMapStringFloat64(m map[string]float64) map[string]float64 {
	if m == nil {
		return nil
	}
	ret := make(map[string]float64)
	for k, v := range m {
		ret[k] = v
	}
	return ret
}
