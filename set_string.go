package ravendb

import "strings"

// TODO: write tests

// Set_String is a Go equivalent of Java's Set<String> for easy porting
type Set_String struct {
	strings []string
	cmp     func(string, string) bool
}

func NewSet_String() *Set_String {
	return &Set_String{
		cmp: String_defaultCompare,
	}
}

func String_defaultCompare(s1, s2 string) bool {
	return s1 == s2
}

func String_compareToIgnoreCase(s1, s2 string) bool {
	return strings.EqualFold(s1, s2)
}

func (s *Set_String) exist(str string) bool {
	for _, el := range s.strings {
		if s.cmp(el, str) {
			return true
		}
	}
	return false
}

func (s *Set_String) add(str string) {
	if s.exist(str) {
		return
	}
	s.strings = append(s.strings, str)
}
