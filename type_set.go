package ravendb

import "reflect"

// Note: this uses identity of object equality test,
// not value of object equality test.
type TypeSet struct {
	a []reflect.Type
}

func NewTypeSet() *TypeSet {
	return &TypeSet{}
}

func NewTypeSetWithType(t reflect.Type) *TypeSet {
	return &TypeSet{
		a: []reflect.Type{t},
	}
}

func (s *TypeSet) exists(t reflect.Type) bool {
	for _, el := range s.a {
		if el == t {
			return true
		}
	}
	return false
}

func (s *TypeSet) add(t reflect.Type) {
	if !s.exists(t) {
		s.a = append(s.a, t)
	}
}
