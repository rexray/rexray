// +build linux

package semaphore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {
	s, err := Open(name, true, 0644, 1)
	assert.NoError(t, err)

	v, err := s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, v)

	assert.NoError(t, s.Wait())
	assert.NoError(t, err)

	v, err = s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 0, v)

	assert.NoError(t, s.Signal())
	assert.NoError(t, s.Signal())

	v, err = s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, v)

	closeAndUnlink(t, s)
}

func TestTimedWait(t *testing.T) {
	s, err := Open(name, true, 0644, 1)
	assert.NoError(t, err)
	v, err := s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, v)

	timeout := time.Duration(time.Second * 5)

	assert.NoError(t, s.TimedWait(timeout))
	setYet := false
	c1 := make(chan int)
	c2 := make(chan int)

	f1 := func() {
		<-c1
		assert.False(t, setYet)
		assert.NoError(t, s.Signal())
	}

	f2 := func() {
		c1 <- 1
		assert.NoError(t, s.TimedWait(timeout))
		setYet = true
		c2 <- 1
	}

	go f1()
	go f2()
	<-c2

	v, err = s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 0, v)

	closeAndUnlink(t, s)
}
