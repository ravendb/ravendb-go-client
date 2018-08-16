package ravendb

// TODO: make key lookups case-insensitive
type DocumentsById struct {
	inner map[string]*DocumentInfo
}

func NewDocumentsById() *DocumentsById {
	return &DocumentsById{
		inner: map[string]*DocumentInfo{},
	}
}

func (d *DocumentsById) getValue(id string) *DocumentInfo {
	return d.inner[id]
}

func (d *DocumentsById) add(info *DocumentInfo) {
	// TODO: this happens in Java but in Go it breaks change tracking
	// e.g. crudTest_crudOperationsWithNull
	// Not sure if should restore this and the bug is caused by some
	// other code path
	// see investigation notes in https://www.notion.so/Investigation-crudTest_crudOperationsWithNull-failure-a37969d8118e4eb39635e458a369ab09
	/*
		if _, ok := d.inner[info.id]; ok {
			return
		}
	*/

	d.inner[info.id] = info
}

func (d *DocumentsById) remove(id string) bool {
	if _, ok := d.inner[id]; ok {
		return false
	}
	delete(d.inner, id)
	return true
}

func (d *DocumentsById) clear() {
	d.inner = map[string]*DocumentInfo{}
}

func (d *DocumentsById) getCount() int {
	return len(d.inner)
}
