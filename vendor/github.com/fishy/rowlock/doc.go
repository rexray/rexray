// Package rowlock provides an implementation of row lock.
//
// A row lock is a set of locks associated with rows.
// Instead of locking and unlocking globally,
// you only operate locks on a row level.
//
// RowLock provides optional RLock and RUnlock functions to use separated read
// and write locks.
// In order to take advantage of them,
// NewLocker function used in NewRowLock must returns an implementation of
// RWLocker (for example, RWMutexNewLocker returns a new sync.RWMutex).
// If the locker returned by NewLocker didn't implement RLocker function defined
// in RWLocker,
// RLock will work the same as Lock and RUnlock will work the same as Unlock.
package rowlock
