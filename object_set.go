package ravendb

// TODO: possibly use []interface{} to be more efficient, assuming this doesn't
// grow to large number of items
type ObjectSet struct {
	items map[interface{}]struct{}
}

func NewObjectSet() *ObjectSet {
	return &ObjectSet{
		items: map[interface{}]struct{}{},
	}
}
func (s *ObjectSet) isEmpty() bool {
	return len(s.items) == 0
}

func (s *ObjectSet) add(o interface{}) {
	s.items[o] = struct{}{}
}

func (s *ObjectSet) remove(o interface{}) {
	delete(s.items, o)
}

func (s *ObjectSet) contains(o interface{}) bool {
	_, ok := s.items[o]
	return ok
}

func (s *ObjectSet) clear() {
	s.items = map[interface{}]struct{}{}
}
