package ravendb

import (
	"errors"
	"fmt"
)

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

func isValidDbNameChar(c rune) bool {
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	switch c {
	case '_', '-', '.':
		return true
	}
	return false
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/tools/utils.py#L47
// returns nil if db name is ok
func isDatabaseNameValid(dbName string) error {
	if dbName == "" {
		return errors.New("database name cannot be empty")
	}
	for _, c := range dbName {
		if !isValidDbNameChar(c) {
			return fmt.Errorf(`Database name can only contain only A-Z, a-z, _, . or - but was: %s`, dbName)
		}
	}
	return nil
}
