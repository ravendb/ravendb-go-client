package ravendb

import (
	"encoding/json"
	"strings"
)

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

func NewStringSetFromStrings(strings ...string) *StringSet {
	set := NewStringSet()
	for _, s := range strings {
		set.Add(s)
	}
	return set
}

// NewStringSetNoCase creates a string set which ignores case where comparing strings
func NewStringSetNoCase() *StringSet {
	return &StringSet{
		cmp: strings.EqualFold,
	}
}

func (s *StringSet) Contains(str string) bool {
	for _, el := range s.strings {
		if s.cmp(el, str) {
			return true
		}
	}
	return false
}

func (s *StringSet) Size() int {
	return len(s.strings)
}

func (s *StringSet) IsEmpty() bool {
	// for convenience, make it work with nil receiver
	if s == nil {
		return true
	}
	return len(s.strings) == 0
}

func (s *StringSet) Strings() []string {
	// for convenience, make it work with nil receiver
	if s == nil {
		return nil
	}
	return s.strings
}

func (s *StringSet) Add(str string) {
	if s.Contains(str) {
		return
	}
	s.strings = append(s.strings, str)
}

func (s *StringSet) Remove(str string) {
	StringArrayRemoveCustomCompare(&s.strings, str, s.cmp)
}

func (s *StringSet) Clear() {
	s.strings = nil
}

// MarshalJSON marshals StringSet as JSON in the form of array
// of string ["str1", "str2"]
func (s *StringSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.strings)
}

// UnmarshalJSON decodes JSON as
func (s *StringSet) UnmarshalJSON(d []byte) error {
	if len(d) == 0 {
		s.strings = nil
		return nil
	}
	var strings []string
	err := json.Unmarshal(d, &strings)
	if err != nil {
		return err
	}
	if len(strings) == 0 {
		strings = nil
	}
	s.strings = strings
	return nil
}
