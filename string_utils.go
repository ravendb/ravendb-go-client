package ravendb

import "unicode"

// Go port of org.apache.commons.lang3.StringUtils

func StringUtils_isEmpty(s string) bool {
	return s == ""
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
