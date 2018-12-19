package ravendb

// Note: Java has IMetadataAsDictionary which is not needed in Go
// so we use concrete type MetadataAsDictionary

// MetadataAsDictionary describes metadata for a document
type MetadataAsDictionary struct {
	_parent    *MetadataAsDictionary
	_parentKey string

	// the actual metadata
	_metadata map[string]interface{}
	_source   ObjectNode

	dirty bool
}

// NewMetadataAsDictionaryWithSource returns MetadataAsDictionary based on a given source
func NewMetadataAsDictionaryWithSource(metadata ObjectNode) *MetadataAsDictionary {
	return &MetadataAsDictionary{
		_source: metadata,
	}
}

// NewMetadataAsDictionaryWithMetadata returns MetadataAsDictionary based on a given metadata
func NewMetadataAsDictionaryWithMetadata(metadata map[string]interface{}) *MetadataAsDictionary {
	return &MetadataAsDictionary{
		_metadata: metadata,
	}
}

// NewMetadataAsDictionary returns MetadataAsDictionary based on a given metadata and parent
func NewMetadataAsDictionary(metadata ObjectNode, parent *MetadataAsDictionary, parentKey string) *MetadataAsDictionary {
	panicIf(parent == nil, "Parent cannot be null")
	panicIf(parentKey == "", "ParentKey cannot be empty")
	return &MetadataAsDictionary{
		_source:    metadata,
		_parent:    parent,
		_parentKey: parentKey,
	}
}

// MarkDirty marks us as dirty
func (d *MetadataAsDictionary) MarkDirty() {
	d.dirty = true
}

// IsDirty returns if we're dirty
func (d *MetadataAsDictionary) IsDirty() bool {
	return d.dirty
}

// KeySet returns all keys
func (d *MetadataAsDictionary) KeySet() []string {
	if d._metadata == nil {
		d.Init()
	}
	// TODO: pre-allocate res
	var res []string
	for k := range d._metadata {
		res = append(res, k)
	}
	return res
}

// Init initializes metadata
func (d *MetadataAsDictionary) Init() {
	d.dirty = true
	d._metadata = map[string]interface{}{}

	for k, v := range d._source {
		val := d.ConvertValue(k, v)
		d._metadata[k] = val
	}

	if d._parent != nil {
		d._parent.Put(d._parentKey, d)
	}
}

// Put inserts a given value with a given key
func (d *MetadataAsDictionary) Put(key string, value interface{}) interface{} {
	if d._metadata == nil {
		d.Init()
	}
	d.dirty = true

	d._metadata[key] = value
	return value
}

// ConvertValue converts value with a given key to a desired type
func (d *MetadataAsDictionary) ConvertValue(key string, value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case int, bool, string, float32, float64: // TODO: more int types?
		return value
	case ObjectNode:
		// TODO: not sure what to do here. Relevant test case: TestRavenDB10641
		//return NewMetadataAsDictionary(v, d, key)
		return v
	case []interface{}:
		n := len(v)
		res := make([]interface{}, n, n)
		for i, el := range v {
			newEl := d.ConvertValue(key, el)
			res[i] = newEl
		}
		return res
	default:
		panicIf(true, "unsuppoted type %T", value)
	}

	return nil
}

// Clear removes all metadata
func (d *MetadataAsDictionary) Clear() {
	if d._metadata == nil {
		d.Init()
	}
	d.dirty = true

	d._metadata = map[string]interface{}{} // TODO: can it be nil?
}

// Get returns metadata value with a given key
func (d *MetadataAsDictionary) Get(key string) (interface{}, bool) {
	if d._metadata != nil {
		v, ok := d._metadata[key]
		return v, ok
	}

	v, ok := d._source[key]
	if !ok {
		return v, ok
	}
	return d.ConvertValue(key, v), ok
}

// EntrySet returns metadata as map[string]interface{}
func (d *MetadataAsDictionary) EntrySet() map[string]interface{} {
	if d._metadata == nil {
		d.Init()
	}

	return d._metadata
}

// ContainsKey returns true if we have metadata value with a given key
func (d *MetadataAsDictionary) ContainsKey(key string) bool {
	if d._metadata != nil {
		_, ok := d._metadata[key]
		return ok
	}

	_, ok := d._source[key]
	return ok
}

// GetObjects returns metadata info for a given key
// TODO: return an error instead of panicking on cast failures?
func (d *MetadataAsDictionary) GetObjects(key string) []*MetadataAsDictionary {
	objI, ok := d.Get(key)
	if !ok || objI == nil {
		return nil
	}
	obj := objI.([]interface{})
	n := len(obj)
	if n == 0 {
		return nil
	}
	list := make([]*MetadataAsDictionary, n, n)
	for i := 0; i < n; i++ {
		if d, ok := obj[i].(map[string]interface{}); ok {
			list[i] = NewMetadataAsDictionaryWithMetadata(d)
			continue
		}
		v := obj[i].(*MetadataAsDictionary)
		list[i] = v
	}
	return list
}

// Size returns number of metadata items
func (d *MetadataAsDictionary) Size() int {
	if d._metadata != nil {
		return len(d._metadata)
	}

	return len(d._source)
}

func (d *MetadataAsDictionary) IsEmpty() bool {
	return d.Size() == 0
}

func (d *MetadataAsDictionary) Remove(key string) {
	if d._metadata == nil {
		return
	}
	d.dirty = true

	delete(d._metadata, key)
}

/*
    @Override
    public void putAll(Map<? extends string, ?> m) {
        if (_metadata == null) {
            init();
        }
        dirty = true;

        _metadata.putAll(m);
    }

    @Override
    public boolean containsValue(Object value) {
        if (_metadata == null) {
            init();
        }

        return _metadata.containsValue(value);
    }

    @Override
    public Collection<Object> values() {
        if (_metadata == null) {
            init();
        }

        return _metadata.values();
    }

    @Override

}
*/
