package ravendb

// CompareExchangeValue represents value for compare exchange
type CompareExchangeValue struct {
	Key   string
	Index int
	Value interface{}
}

// NewCompareExchangeValue returns new CompareExchangeValue
func NewCompareExchangeValue(key string, index int, value interface{}) *CompareExchangeValue {
	return &CompareExchangeValue{
		Key:   key,
		Index: index,
		Value: value,
	}
}
