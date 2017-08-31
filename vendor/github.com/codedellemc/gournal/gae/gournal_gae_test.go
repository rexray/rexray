// +build none

package gae

import (
	"fmt"
	"os"
	"testing"

	gaetest "google.golang.org/appengine/aetest"

	"github.com/codedellemc/gournal"
)

var gaeCtx gournal.Context

func TestMain(m *testing.M) {

	var (
		err  error
		done func()
	)

	if gaeCtx, done, err = gaetest.NewContext(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ec := m.Run()

	done()
	os.Exit(ec)
}

func TestGAEAppenderNoFields(t *testing.T) {
	gournal.Info(ctx(), "Hello %s", "Bob")
}

func TestGAEAppenderWithField(t *testing.T) {
	gournal.WithField("size", 2).Info(ctx(), "Hello %s", "Alice")
}

func TestGAEAppenderWithFields(t *testing.T) {
	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx(), "Hello %s", "Mary")
}

func TestGAEAppenderPanic(t *testing.T) {
	gournal.Panic(ctx(), "Hello %s", "Bob")
}

func ctx() gournal.Context {
	ctx := gournal.WithValue(gaeCtx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = gournal.WithValue(ctx, gournal.AppenderKey(), New())
	return ctx
}
