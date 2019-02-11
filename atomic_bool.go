package ravendb

import "sync/atomic"

type atomicBool int32

func atomicBoolSet(av *atomicBool, v bool) {
	var n int32
	if v {
		n = 1
	}
	atomic.StoreInt32((*int32)(av), n)
}

func atomicBoolIsSet(av *atomicBool) bool {
	return atomic.LoadInt32((*int32)(av)) != 0
}
