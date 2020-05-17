package rowlock

import (
	"sync"
)

// NewLocker defines a type of function that can be used to create a new Locker.
type NewLocker func() sync.Locker

// Row is the type of a row.
//
// It must be hashable.
type Row interface{}

// RWLocker is the abstracted interface of sync.RWMutex.
type RWLocker interface {
	sync.Locker

	RLocker() sync.Locker
}

// Make sure that sync.RWMutex is compatible with RWLocker interface.
var _ RWLocker = (*sync.RWMutex)(nil)

// MutexNewLocker is a NewLocker using sync.Mutex.
func MutexNewLocker() sync.Locker {
	return new(sync.Mutex)
}

// RWMutexNewLocker is a NewLocker using sync.RWMutex.
func RWMutexNewLocker() sync.Locker {
	return new(sync.RWMutex)
}

// RowLock defines a set of locks.
//
// When you do Lock/Unlock operations, you don't do them on a global scale.
// Instead, a Lock/Unlock operation is operated on a given row.
//
// If NewLocker returns an implementation of RWLocker in NewRowLock,
// the RowLock can be locked separately for read in RLock and RUnlock functions.
// Otherwise, RLock is the same as Lock and RUnlock is the same as Unlock.
type RowLock struct {
	locks      sync.Map
	lockerPool sync.Pool
}

// NewRowLock creates a new RowLock with the given NewLocker.
func NewRowLock(f NewLocker) *RowLock {
	return &RowLock{
		lockerPool: sync.Pool{
			New: func() interface{} {
				return f()
			},
		},
	}
}

// Lock locks a row.
//
// If this is a new row,
// a new locker will be created using the NewLocker specified in NewRowLock.
func (rl *RowLock) Lock(row Row) {
	rl.getLocker(row).Lock()
}

// Unlock unlocks a row.
func (rl *RowLock) Unlock(row Row) {
	rl.getLocker(row).Unlock()
}

// RLock locks a row for read.
//
// It only works as expected when NewLocker specified in NewRowLock returns an
// implementation of RWLocker. Otherwise, it's the same as Lock.
func (rl *RowLock) RLock(row Row) {
	rl.getRLocker(row).Lock()
}

// RUnlock unlocks a row for read.
//
// It only works as expected when NewLocker specified in NewRowLock returns an
// implementation of RWLocker. Otherwise, it's the same as Unlock.
func (rl *RowLock) RUnlock(row Row) {
	rl.getRLocker(row).Unlock()
}

// getLocker returns the lock for the given row.
//
// If this is a new row,
// a new locker will be created using the NewLocker specified in NewRowLock.
func (rl *RowLock) getLocker(row Row) sync.Locker {
	newLocker := rl.lockerPool.Get()
	locker, loaded := rl.locks.LoadOrStore(row, newLocker)
	if loaded {
		rl.lockerPool.Put(newLocker)
	}
	return locker.(sync.Locker)
}

// getRLocker returns the lock for read for the given row.
//
// If this is a new row,
// a new locker will be created using the NewLocker specified in NewRowLock.
//
// If NewLocker specified in NewRowLock returns a locker that didn't implement
// GetRLocker, the locker itself will be returned instead.
func (rl *RowLock) getRLocker(row Row) sync.Locker {
	locker := rl.getLocker(row)
	if rwlocker, ok := locker.(RWLocker); ok {
		return rwlocker.RLocker()
	}
	return locker
}
