package ravendb

import "fmt"

func must(err error) {
	if err != nil {
		panic(err.Error)
	}
}

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		err := fmt.Errorf(format, args...)
		must(err)
	}
}
