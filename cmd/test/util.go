package main

import "fmt"

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		err := fmt.Errorf(format, args...)
		must(err)
	}
}

func stringInArray(a []string, s string) bool {
	for _, s2 := range a {
		if s2 == s {
			return true
		}
	}
	return false
}
