package ravendb

// Semaphore is a Go implementation of Java's Semaphore
type Semaphore struct {
	c chan bool
}

func NewSemaphore(cap int) *Semaphore {
	return &Semaphore{
		c: make(chan bool, cap),
	}
}

func (s *Semaphore) acquire() {
	s.c <- true
}

func (s *Semaphore) release() {
	<-s.c
}
