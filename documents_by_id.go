package ravendb

import "strings"

type documentsByID struct {
	inner map[string]*documentInfo
}

func newDocumentsByID() *documentsByID {
	return &documentsByID{
		inner: map[string]*documentInfo{},
	}
}

func (d *documentsByID) getValue(id string) *documentInfo {
	id = strings.ToLower(id)
	return d.inner[id]
}

func (d *documentsByID) add(info *documentInfo) {
	id := strings.ToLower(info.id)

	if _, ok := d.inner[id]; ok {
		return
	}

	d.inner[id] = info
}

func (d *documentsByID) remove(id string) bool {
	id = strings.ToLower(id)
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
