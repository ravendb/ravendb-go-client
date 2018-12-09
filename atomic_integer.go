package ravendb

import "sync/atomic"

// atomicInteger makes porting Java code easier
type atomicInteger struct {
	N int32
}

func (i *atomicInteger) incrementAndGet() int {
	res := atomic.AddInt32(&i.N, 1)
	return int(res)
}

func (i *atomicInteger) get() int {
	res := atomic.LoadInt32(&i.N)
	return int(res)
}

func (i *atomicInteger) Get() int {
	res := atomic.LoadInt32(&i.N)
	return int(res)
}

func (i *atomicInteger) set(n int) {
	atomic.StoreInt32(&i.N, int32(n))
}

func (i *atomicInteger) compareAndSet(old, new int) bool {
	return atomic.CompareAndSwapInt32(&i.N, int32(old), int32(new))
}

func (i *atomicInteger) decrementAndGet() int {
	res := atomic.AddInt32(&i.N, -1)
	return int(res)
}
