// +build none

package gae

import (
	"context"
	"fmt"
	"os"
	"testing"

	gaetest "google.golang.org/appengine/aetest"

	"github.com/thecodeteam/gournal"
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

func ctx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = context.WithValue(ctx, gournal.AppenderKey(), New())
	return ctx
}
