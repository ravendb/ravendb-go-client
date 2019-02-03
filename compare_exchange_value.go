package ravendb

// CompareExchangeValue represents value for compare exchange
type CompareExchangeValue struct {
	Key   string
	Index int64
	Value interface{}
}

// NewCompareExchangeValue returns new CompareExchangeValue
func NewCompareExchangeValue(key string, index int64, value interface{}) *CompareExchangeValue {
	return &CompareExchangeValue{
		Key:   key,
		Index: index,
		Value: value,
	}
}
