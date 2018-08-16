package ravendb

// IMetadataDictionary describes metadata for a document
// Note: in Java there's only one subclass of IMetadataDictionary, so for
// easy porting we alias its name to the implementation
// TODO: remove the alias when porting is done
type IMetadataDictionary = MetadataAsDictionary

type MetadataAsDictionary struct {
	_parent    *MetadataAsDictionary
	_parentKey string

	// the actual metadata
	_metadata map[string]Object
	_source   ObjectNode

	dirty bool
}

func NewMetadataAsDictionaryWithSource(metadata ObjectNode) *MetadataAsDictionary {
	return &MetadataAsDictionary{
		_source: metadata,
	}
}

func NewMetadataAsDictionaryWithMetadata(metadata map[string]interface{}) *MetadataAsDictionary {
	return &MetadataAsDictionary{
		_metadata: metadata,
	}
}

func NewMetadataAsDictionary(metadata ObjectNode, parent *IMetadataDictionary, parentKey string) *MetadataAsDictionary {
	panicIf(parent == nil, "Parent cannot be null")
	panicIf(parentKey == "", "ParentKey cannot be empty")
	return &MetadataAsDictionary{
		_source:    metadata,
		_parent:    parent,
		_parentKey: parentKey,
	}
}

func (d *MetadataAsDictionary) IsDirty() bool {
	return d.dirty
}

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

func (d *MetadataAsDictionary) Init() {
	d.dirty = true
	d._metadata = map[string]Object{}

	for k, v := range d._source {
		val := d.ConvertValue(k, v)
		d._metadata[k] = val
	}

	if d._parent != nil {
		d._parent.Put(d._parentKey, d)
	}
}

func (d *MetadataAsDictionary) Put(key string, value Object) Object {
	if d._metadata == nil {
		d.Init()
	}
	d.dirty = true

	d._metadata[key] = value
	return value
}

func (d *MetadataAsDictionary) ConvertValue(key string, value Object) Object {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case int, bool, string, float32, float64: // TODO: more int types?
		return value
	case ObjectNode:
		return NewMetadataAsDictionary(v, d, key)
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

func (d *MetadataAsDictionary) Clear() {
	if d._metadata == nil {
		d.Init()
	}
	d.dirty = true

	d._metadata = map[string]Object{} // TODO: can it be nil?
}

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

func (d *MetadataAsDictionary) EntrySet() map[string]Object {
	if d._metadata == nil {
		d.Init()
	}

	return d._metadata
}

func (d *MetadataAsDictionary) ContainsKey(key string) bool {
	if d._metadata != nil {
		_, ok := d._metadata[key]
		return ok
	}

	_, ok := d._source[key]
	return ok
}

// TODO: return an error instead of panicking on cast failures
func (d *MetadataAsDictionary) GetObjects(key string) []*IMetadataDictionary {
	objI, ok := d.Get(key)
	if !ok || objI == nil {
		return nil
	}
	obj := objI.([]interface{})
	n := len(obj)
	if n == 0 {
		return nil
	}
	list := make([]*IMetadataDictionary, n, n)
	for i := 0; i < n; i++ {
		v := obj[i].(*IMetadataDictionary)
		list[i] = v
	}
	return list
}

/*
    @Override
    public int size() {
        if (_metadata != null) {
            return _metadata.size();
        }

        return _source.size();
    }


    @Override
    public boolean isEmpty() {
        return size() == 0;
    }

    @Override
    public void putAll(Map<? extends string, ?> m) {
        if (_metadata == null) {
            init();
        }
        dirty = true;

        _metadata.putAll(m);
    }

    @Override
    public Object remove(Object key) {
        if (_metadata == null) {
            init();
        }
        dirty = true;

        return _metadata.remove(key);
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
