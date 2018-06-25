package ravendb

import "strings"

// TODO: write tests

// Note: if it handles large number of strings, should use map[string]struct{}

// StringSet is a Go equivalent of Java's Set<string> for easy porting
type StringSet struct {
	strings []string
	cmp     func(string, string) bool
}

func String_defaultCompare(s1, s2 string) bool {
	return s1 == s2
}

func String_compareToIgnoreCase(s1, s2 string) bool {
	return strings.EqualFold(s1, s2)
}

func NewStringSet() *StringSet {
	return &StringSet{
		cmp: String_defaultCompare,
	}
}

// NewStringSetNoCase creates a string set which ignores case where comparing strings
func NewStringSetNoCase() *StringSet {
	return &StringSet{
		cmp: strings.EqualFold,
	}
}

func (s *StringSet) contains(str string) bool {
	for _, el := range s.strings {
		if s.cmp(el, str) {
			return true
		}
	}
	return false
}

func (s *StringSet) add(str string) {
	if s.contains(str) {
		return
	}
	s.strings = append(s.strings, str)
}

func (s *StringSet) remove(str string) {
	stringArrayRemoveCustomCompare(&s.strings, str, s.cmp)
}

func (s *StringSet) clear() {
	s.strings = nil
}
