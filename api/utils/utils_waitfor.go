package utils

import "time"

// WaitFor waits for a lambda to complete or aborts after a specified amount
// of time. If the function fails to complete in the specified amount of time
// then the second return value is a boolean false. Otherwise the return value
// and possible error of the provided lambda are returned.
func WaitFor(
	f func() (interface{}, error),
	timeout time.Duration) (interface{}, bool, error) {

	var (
		res interface{}
		err error

		fc = make(chan bool, 1)
	)

	go func() {
		res, err = f()
		fc <- true
	}()
	tc := time.After(timeout)

	select {
	case <-fc:
		return res, true, err
	case <-tc:
		return nil, false, nil
	}
}
