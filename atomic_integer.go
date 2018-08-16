package ravendb

import "sync/atomic"

// AtomicInteger makes porting Java code easier
// TODO: make it be int32 on 32-bit and int64 on 64-bit
type AtomicInteger struct {
	N int32
}

func (i *AtomicInteger) IncrementAndGet() int {
	res := atomic.AddInt32(&i.N, 1)
	return int(res)
}

func (i *AtomicInteger) Get() int {
	res := atomic.LoadInt32(&i.N)
	return int(res)
}

func (i *AtomicInteger) Set(n int) {
	atomic.StoreInt32(&i.N, int32(n))
}

func (i *AtomicInteger) CompareAndSet(old, new int) bool {
	return atomic.CompareAndSwapInt32(&i.N, int32(old), int32(new))
}

func (i *AtomicInteger) DecrementAndGet() int {
	res := atomic.AddInt32(&i.N, -1)
	return int(res)
}
