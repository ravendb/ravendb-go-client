package ravendb

type CompareExchangeValue struct {
	key   string
	index int
	value interface{}
}

func NewCompareExchangeValue(key string, index int, value interface{}) *CompareExchangeValue {
	return &CompareExchangeValue{
		key:   key,
		index: index,
		value: value,
	}
}

func (v *CompareExchangeValue) getKey() string {
	return v.key
}

func (v *CompareExchangeValue) setKey(key string) {
	v.key = key
}

func (v *CompareExchangeValue) getIndex() int {
	return v.index
}

func (v *CompareExchangeValue) setIndex(index int) {
	v.index = index
}

func (v *CompareExchangeValue) getValue() interface{} {
	return v.value
}

func (v *CompareExchangeValue) setValue(value interface{}) {
	v.value = value
}
