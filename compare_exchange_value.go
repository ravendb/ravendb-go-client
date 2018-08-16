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

func (v *CompareExchangeValue) GetKey() string {
	return v.key
}

func (v *CompareExchangeValue) SetKey(key string) {
	v.key = key
}

func (v *CompareExchangeValue) GetIndex() int {
	return v.index
}

func (v *CompareExchangeValue) SetIndex(index int) {
	v.index = index
}

func (v *CompareExchangeValue) GetValue() interface{} {
	return v.value
}

func (v *CompareExchangeValue) SetValue(value interface{}) {
	v.value = value
}
