package ravendb

// Semaphore is a Go implementation of Java's Semaphore
type Semaphore struct {
	c chan struct{}
}

func NewSemaphore(cap int) *Semaphore {
	return &Semaphore{
		c: make(chan struct{}, cap),
	}
}

func (s *Semaphore) acquire() {
	s.c <- struct{}{}
}

func (s *Semaphore) release() {
	<-s.c
}
