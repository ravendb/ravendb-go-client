package ravendb

type CmpXchg struct {
	MethodCallData
}

func CmpXchg_value(key string) *CmpXchg {
	cmpXchg := &CmpXchg{}
	cmpXchg.args = []interface{}{key}
	return cmpXchg
}
