package ravendb

import "sync/atomic"

type atomicBool int32

func (b *atomicBool) set(v bool) {
	val := int32(0)
	if v {
		val = 1
	}
	atomic.StoreInt32((*int32)(b), val)
}

func (b *atomicBool) get() bool {
	v := atomic.LoadInt32((*int32)(b))
	return v == 1
}

func (b *atomicBool) isTrue() bool {
	v := atomic.LoadInt32((*int32)(b))
	return v == 1
}
