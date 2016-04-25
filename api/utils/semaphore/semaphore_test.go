package semaphore

import (
	"os"
	"testing"

	"github.com/akutz/goof"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/types"
)

var (
	name = types.LSX
)

func init() {
	goof.IncludeFieldsInFormat = true
}

func TestMain(m *testing.M) {
	Unlink(name)
	ec := m.Run()
	Unlink(name)
	os.Exit(ec)
}

func TestOpenClose(t *testing.T) {
	s, err := Open(name, true, 0644, 0)
	assert.NoError(t, err)

	_, err = Open(name, true, 0644, 0)
	assert.Error(t, err)
	gerr := err.(goof.Goof)
	assert.EqualValues(t, gerr.Fields()["error"], 17)

	closeAndUnlink(t, s)
}

func TestTryWait(t *testing.T) {
	s, err := Open(name, true, 0644, 0)
	assert.NoError(t, err)

	ea, err := s.TryWait()
	assert.True(t, ea)
	assert.NoError(t, err)

	assert.NoError(t, s.Signal())
	ea, err = s.TryWait()
	assert.False(t, ea)
	assert.NoError(t, err)

	closeAndUnlink(t, s)
}

func TestWait(t *testing.T) {
	s, err := Open(name, true, 0644, 1)
	assert.NoError(t, err)
	v, err := s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, v)

	assert.NoError(t, s.Wait())
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
		assert.NoError(t, s.Wait())
		setYet = true
		c2 <- 1
	}

	go f1()
	go f2()
	<-c2

	v, err = s.Value()
	assert.NoError(t, err)
	assert.EqualValues(t, 0, v)
}

func closeAndUnlink(t *testing.T, s Semaphore) {
	assert.NoError(t, s.Close())
	assert.NoError(t, Unlink(s.Name()))
}
