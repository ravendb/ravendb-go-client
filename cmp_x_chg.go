package ravendb

type CmpXchg struct {
	MethodCallData
}

func CmpXchg_value(key string) *CmpXchg {
	cmpXchg := &CmpXchg{}
	cmpXchg.args = []Object{key}
	return cmpXchg
}
