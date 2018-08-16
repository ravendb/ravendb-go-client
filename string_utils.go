package ravendb

import "unicode"

// Go port of org.apache.commons.lang3.StringUtils

// TODO: replace with direct code
func StringUtils_isEmpty(s string) bool {
	return s == ""
}

// TODO: replace with direct code
func StringUtils_isNotEmpty(s string) bool {
	return s != ""
}

func StringUtils_isWhitespace(s string) bool {
	for _, c := range s {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func StringUtils_isBlank(s string) bool {
	for _, c := range s {
		if c != ' ' {
			return false
		}
	}
	return true
}

func StringUtils_isNotBlank(s string) bool {
	return !StringUtils_isBlank(s)
}
