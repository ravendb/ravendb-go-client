package ravendb

// TODO: possibly use []Object to be more efficient, assuming this doesn't
// grow to large number of items
type ObjectSet struct {
	items map[Object]struct{}
}

func NewObjectSet() *ObjectSet {
	return &ObjectSet{
		items: map[Object]struct{}{},
	}
}
func (s *ObjectSet) isEmpty() bool {
	return len(s.items) == 0
}

func (s *ObjectSet) add(o Object) {
	s.items[o] = struct{}{}
}

func (s *ObjectSet) remove(o Object) {
	delete(s.items, o)
}

func (s *ObjectSet) contains(o Object) bool {
	_, ok := s.items[o]
	return ok
}

func (s *ObjectSet) clear() {
	s.items = map[Object]struct{}{}
}
