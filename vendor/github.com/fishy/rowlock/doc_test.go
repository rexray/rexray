package rowlock_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/fishy/rowlock"
)

func Example() {
	lock := rowlock.NewRowLock(rowlock.MutexNewLocker)
	key1 := "key1"
	key2 := "key2"
	round := time.Millisecond * 50

	keys := []string{key1, key1, key2, key2}
	sleeps := []time.Duration{
		time.Millisecond * 250,
		time.Millisecond * 200,
		time.Millisecond * 350,
		time.Millisecond * 300,
	}

	var wg sync.WaitGroup
	wg.Add(len(keys))

	for i := range keys {
		go func(key string, sleep time.Duration) {
			started := time.Now()
			defer wg.Done()
			time.Sleep(sleep)
			lock.Lock(key)
			defer lock.Unlock(key)
			elapsed := time.Now().Sub(started).Round(round)
			// The same key with longer sleep will get an elapsed time about
			// 2 * the same key with shorter sleep instead of its own sleep time,
			// because that's when the other goroutine releases the lock.
			fmt.Printf("%s got lock after about %v\n", key, elapsed)
			time.Sleep(sleep)
		}(keys[i], sleeps[i])
	}

	wg.Wait()
	// Output:
	// key1 got lock after about 200ms
	// key2 got lock after about 300ms
	// key1 got lock after about 400ms
	// key2 got lock after about 600ms
}

func ExampleRWMutexNewLocker() {
	lock := rowlock.NewRowLock(rowlock.RWMutexNewLocker)
	key1 := "key1"
	key2 := "key2"
	round := time.Millisecond * 50

	var wg sync.WaitGroup

	readKeys := []string{key1, key1, key2, key2}
	readSleeps := []time.Duration{
		time.Millisecond * 250,
		time.Millisecond * 200,
		time.Millisecond * 350,
		time.Millisecond * 300,
	}
	wg.Add(len(readKeys))

	writeKeys := []string{key1, key1, key2, key2}
	writeSleeps := []time.Duration{
		time.Millisecond * 350,
		time.Millisecond * 150,
		time.Millisecond * 450,
		time.Millisecond * 200,
	}
	wg.Add(len(writeKeys))

	// Read locks
	for i := range readKeys {
		go func(key string, sleep time.Duration) {
			started := time.Now()
			defer wg.Done()
			time.Sleep(sleep)
			lock.RLock(key)
			defer lock.RUnlock(key)
			elapsed := time.Now().Sub(started).Round(round)
			// Should be:
			//   max(shorter write sleep time * 2, self sleep time)
			fmt.Printf("%s got read lock after about %v\n", key, elapsed)
			time.Sleep(sleep)
		}(readKeys[i], readSleeps[i])
	}

	// Write locks
	for i := range writeKeys {
		go func(key string, sleep time.Duration) {
			started := time.Now()
			defer wg.Done()
			time.Sleep(sleep)
			lock.Lock(key)
			defer lock.Unlock(key)
			elapsed := time.Now().Sub(started).Round(round)
			// For the longer sleep one, it should be
			//   max(shorter write * 2, longer read) + longer read
			// instead of it's self sleep time
			fmt.Printf("%s got lock after about %v\n", key, elapsed)
			time.Sleep(sleep)
		}(writeKeys[i], writeSleeps[i])
	}

	wg.Wait()
	// Output:
	// key1 got lock after about 150ms
	// key2 got lock after about 200ms
	// key1 got read lock after about 300ms
	// key1 got read lock after about 300ms
	// key2 got read lock after about 400ms
	// key2 got read lock after about 400ms
	// key1 got lock after about 550ms
	// key2 got lock after about 750ms
}
