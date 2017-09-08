package gocsi

import "time"

// MutexWithTryLock is a lock object that implements the semantics
// of a sync.Mutex in addition to a TryLock function.
type MutexWithTryLock interface {
	// Lock locks the mutex.
	Lock()
	// Unlock unlocks the mutex.
	Unlock()
	// TryLock attempts to obtain a lock but times out if no lock
	// can be obtained in the specified duration. A flag is returned
	// indicating whether or not the lock was obtained.
	TryLock(timeout time.Duration) bool
}

type mutex struct {
	c chan int
}

// NewMutexWithTryLock returns a new mutex that implements TryLock.
func NewMutexWithTryLock() MutexWithTryLock {
	return &mutex{c: make(chan int, 1)}
}

func (m *mutex) Lock() {
	m.c <- 1
}

func (m *mutex) Unlock() {
	<-m.c
}

func (m *mutex) TryLock(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	select {
	case m.c <- 1:
		timer.Stop()
		return true
	case <-time.After(timeout):
	}
	return false
}
