package ravendb

// IMetadataDictionary describes metadata for a document
// Note: in Java there's only one subclass of IMetadataDictionary, so for
// easy porting we alias its name to the implementation
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

func NewMetadataAsDictionary(metadata ObjectNode, parent *IMetadataDictionary, parentKey String) *MetadataAsDictionary {
	panicIf(parent == nil, "Parent cannot be null")
	panicIf(parentKey == "", "ParentKey cannot be empty")
	return &MetadataAsDictionary{
		_source:    metadata,
		_parent:    parent,
		_parentKey: parentKey,
	}
}

func (d *MetadataAsDictionary) isDirty() bool {
	return d.dirty
}

func (d *MetadataAsDictionary) keySet() []string {
	if d._metadata == nil {
		d.init()
	}
	// TODO: pre-allocate res
	var res []string
	for k := range d._metadata {
		res = append(res, k)
	}
	return res
}

func (d *MetadataAsDictionary) init() {
	d.dirty = true
	d._metadata = map[string]Object{}

	for k, v := range d._source {
		val := d.convertValue(k, v)
		d._metadata[k] = val
	}

	if d._parent != nil {
		d._parent.put(d._parentKey, d)
	}
}

func (d *MetadataAsDictionary) put(key String, value Object) Object {
	if d._metadata == nil {
		d.init()
	}
	d.dirty = true

	d._metadata[key] = value
	return value
}

func (d *MetadataAsDictionary) convertValue(key string, value Object) Object {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case int, bool, string, float32, float64: // TODO: more int types?
		return value
	case ObjectNode:
		return NewMetadataAsDictionary(v, d, key)
	case []interface{}:
		// TODO: not sure this will work
		// TODO: pre-allocate result
		var res []interface{}
		for _, el := range v {
			newEl := d.convertValue(key, el)
			res = append(res, newEl)
		}
		return res
	default:
		panicIf(true, "usuppoted type %T", value)
	}

	return nil
}

func (d *MetadataAsDictionary) clear() {
	if d._metadata == nil {
		d.init()
	}
	d.dirty = true

	d._metadata = map[string]Object{} // TODO: can it be nil?
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
    public Object get(Object key) {
        if (_metadata != null) {
            return _metadata.get(key);
        }

        return convertValue((String) key, _source.get((String) key));
    }

    @Override
    public boolean isEmpty() {
        return size() == 0;
    }

    @Override
    public void putAll(Map<? extends String, ?> m) {
        if (_metadata == null) {
            init();
        }
        dirty = true;

        _metadata.putAll(m);
    }

    @Override

    @Override
    public boolean containsKey(Object key) {
        if (_metadata != null) {
            return _metadata.containsKey(key);
        }

        return _source.has((String)key);
    }

    @Override
    public Set<Entry<String, Object>> entrySet() {
        if (_metadata == null) {
            init();
        }

        return _metadata.entrySet();
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
