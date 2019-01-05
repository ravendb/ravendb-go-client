package ravendb

// TODO: objectSet is only used for deletedEntities and probably
// doesn't need to be a set i.e. duplicates are ok

// remove duplicate objects from a. It's somewhat expensive O(n^2) but
// so is every other way of doing this.
// we could use a hash for a one-time pass but that would restrict
// possible values to only hashables (e.g. not map)
// TODO: write tests
/*
func removeDuplicatesFromObjectSet(a []interface{}) []interface{} {
	n := len(a)
	for i := 0; i < n; i++ {
		el := a[i]
		found := false
		n := len(a)
		for j := i; !found && j < n; j++ {
			el2 := a[j]
			if el == el2 {
				found = true
				a[j] = a[n-1]
				a[n-1] = nil
				a = a[:n-1]
				n--
				if i == j {
					i--
				}
			}
		}
	}
	return a
}
*/

// TODO: possibly use []interface{} to be more efficient, assuming this doesn't
// grow to large number of items
type objectSet struct {
	items map[interface{}]struct{}
}

func newObjectSet() *objectSet {
	return &objectSet{
		items: map[interface{}]struct{}{},
	}
}

func (s *objectSet) add(o interface{}) {
	s.items[o] = struct{}{}
}

func (s *objectSet) remove(o interface{}) {
	delete(s.items, o)
}

func (s *objectSet) contains(o interface{}) bool {
	_, ok := s.items[o]
	return ok
}

func (s *objectSet) clear() {
	s.items = map[interface{}]struct{}{}
}
