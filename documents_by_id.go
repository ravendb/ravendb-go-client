package ravendb

import (
	"strings"
	"sync"
)

// TODO: change to an alias:
//  type documentsByID map[string]*documentInfo

type documentsByID struct {
	inner sync.Map
}

func newDocumentsByID() *documentsByID {
	return &documentsByID{
		inner: sync.Map{},
	}
}

func (d *documentsByID) getValue(id string) *documentInfo {
	id = strings.ToLower(id)
	value, ok := d.inner.Load(id)
	if !ok {
		return nil
	}
	return value.(*documentInfo)
}

func (d *documentsByID) add(info *documentInfo) {
	id := strings.ToLower(info.id)

	_, ok := d.inner.Load(id)

	if ok {
		return
	}

	d.inner.Store(id, info)
}

func (d *documentsByID) remove(id string) bool {
	id = strings.ToLower(id)

	_, ok := d.inner.Load(id)
	if !ok {
		return false
	}
	d.inner.Delete(id)
	return true
}
