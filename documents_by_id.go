package ravendb

// TODO: make key lookups case-insensitive
type documentsByID struct {
	inner map[string]*documentInfo
}

func newDocumentsByID() *documentsByID {
	return &documentsByID{
		inner: map[string]*documentInfo{},
	}
}

func (d *documentsByID) getValue(id string) *documentInfo {
	return d.inner[id]
}

func (d *documentsByID) add(info *documentInfo) {
	if _, ok := d.inner[info.id]; ok {
		return
	}

	d.inner[info.id] = info
}

func (d *documentsByID) remove(id string) bool {
	if _, ok := d.inner[id]; !ok {
		return false
	}
	delete(d.inner, id)
	return true
}

func (d *documentsByID) clear() {
	d.inner = map[string]*documentInfo{}
}

func (d *documentsByID) getCount() int {
	return len(d.inner)
}
