package ravendb

import "sync/atomic"

// AtomicInteger makes porting Java code easier
// TODO: make it be int32 on 32-bit and int64 on 64-bit
type AtomicInteger struct {
	N int32
}

func (i *AtomicInteger) incrementAndGet() int {
	res := atomic.AddInt32(&i.N, 1)
	return int(res)
}

func (i *AtomicInteger) get() int {
	res := atomic.LoadInt32(&i.N)
	return int(res)
}

func (i *AtomicInteger) set(n int) {
	atomic.StoreInt32(&i.N, int32(n))
}

func (i *AtomicInteger) compareAndSet(old, new int) bool {
	return atomic.CompareAndSwapInt32(&i.N, int32(old), int32(new))
}

func (i *AtomicInteger) decrementAndGet() int {
	res := atomic.AddInt32(&i.N, -1)
	return int(res)
}
