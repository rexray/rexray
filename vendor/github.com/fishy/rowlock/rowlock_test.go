package rowlock_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fishy/rowlock"
)

func TestRowLock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	lock := rowlock.NewRowLock(rowlock.MutexNewLocker)
	key1 := "key1"
	key2 := "key2"

	short := time.Millisecond * 10
	long := time.Millisecond * 100
	longer := time.Millisecond * 150

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		lock.Lock(key1)
		defer lock.Unlock(key1)
		time.Sleep(long)
	}()

	go func() {
		defer wg.Done()
		lock.Lock(key2)
		defer lock.Unlock(key2)
		time.Sleep(longer)
	}()

	go func() {
		defer wg.Done()
		started := time.Now()
		time.Sleep(short)
		lock.Lock(key1)
		defer lock.Unlock(key1)
		elapsed := time.Now().Sub(started)
		t.Logf("elapsed time: %v", elapsed)
		if elapsed < long || elapsed > longer {
			t.Errorf(
				"lock wait time should be between %v and %v, actual %v",
				long,
				longer,
				elapsed,
			)
		}
	}()

	wg.Wait()
}

func TestUseNonRWForReadLock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	lock := rowlock.NewRowLock(rowlock.MutexNewLocker)
	key1 := "key1"
	key2 := "key2"

	short := time.Millisecond * 10
	long := time.Millisecond * 100
	longer := time.Millisecond * 150

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		lock.RLock(key1)
		defer lock.RUnlock(key1)
		time.Sleep(long)
	}()

	go func() {
		defer wg.Done()
		lock.RLock(key2)
		defer lock.RUnlock(key2)
		time.Sleep(longer)
	}()

	go func() {
		defer wg.Done()
		started := time.Now()
		time.Sleep(short)
		lock.RLock(key1)
		defer lock.RUnlock(key1)
		elapsed := time.Now().Sub(started)
		t.Logf("elapsed time: %v", elapsed)
		if elapsed < long || elapsed > longer {
			t.Errorf(
				"lock wait time should be between %v and %v, actual %v",
				long,
				longer,
				elapsed,
			)
		}
	}()

	wg.Wait()
}

func BenchmarkLockUnlock(b *testing.B) {
	var numRows = []int{10, 100, 1000}
	var newLockerMap = map[string]rowlock.NewLocker{
		"Mutex":   rowlock.MutexNewLocker,
		"RWMutex": rowlock.RWMutexNewLocker,
	}

	for _, n := range numRows {
		b.Run(
			fmt.Sprintf("%d", n),
			func(b *testing.B) {
				rows := make([]int, n)
				for i := 0; i < n; i++ {
					rows[i] = i
				}
				for label, newLocker := range newLockerMap {
					b.Run(
						label,
						func(b *testing.B) {
							rl := rowlock.NewRowLock(newLocker)

							b.Run(
								"LockUnlock",
								func(b *testing.B) {
									for i := 0; i < b.N; i++ {
										row := rows[i%n]
										rl.Lock(row)
										rl.Unlock(row)
									}
								},
							)

							b.Run(
								"RLockRUnlock",
								func(b *testing.B) {
									for i := 0; i < b.N; i++ {
										row := rows[i%n]
										rl.RLock(row)
										rl.RUnlock(row)
									}
								},
							)
						},
					)
				}
			},
		)
	}
}
