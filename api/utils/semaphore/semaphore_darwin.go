package semaphore

import (
	"time"

	"github.com/akutz/goof"
)

func (s *semaphore) timedWait(t time.Duration) error {
	return goof.New("unsupported")
}

func (s *semaphore) value() (int, error) {
	return int(s.count), nil
}
