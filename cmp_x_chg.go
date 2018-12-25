package ravendb

// CmpXchg represents data for cmp xchg method
type CmpXchg struct {
	MethodCallData
}

// CmpXchgValue returns CmpXchg for a given key
func CmpXchgValue(key string) *CmpXchg {
	cmpXchg := &CmpXchg{}
	cmpXchg.args = []interface{}{key}
	return cmpXchg
}
