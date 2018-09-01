package ravendb

import "unicode"

// Go port of org.apache.commons.lang3.StringUtils

// TODO: replace with direct code
func stringIsEmpty(s string) bool {
	return s == ""
}

// TODO: replace with direct code
func stringIsNotEmpty(s string) bool {
	return s != ""
}

func stringIsWhitespace(s string) bool {
	for _, c := range s {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func stringIsBlank(s string) bool {
	for _, c := range s {
		if c != ' ' {
			return false
		}
	}
	return true
}

func stringIsNotBlank(s string) bool {
	return !stringIsBlank(s)
}
