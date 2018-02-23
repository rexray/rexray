package goisilon

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	log "github.com/akutz/gournal"
	glogrus "github.com/akutz/gournal/logrus"
)

var (
	err        error
	client     *Client
	defaultCtx context.Context
)

func init() {
	defaultCtx = context.Background()
	defaultCtx = context.WithValue(
		defaultCtx,
		log.AppenderKey(),
		glogrus.NewWithOptions(
			logrus.StandardLogger().Out,
			logrus.DebugLevel,
			logrus.StandardLogger().Formatter))
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		defaultCtx = context.WithValue(
			defaultCtx,
			log.LevelKey(),
			log.DebugLevel)
	}

	client, err = NewClient(defaultCtx)
	if err != nil {
		log.WithError(err).Panic(defaultCtx, "error creating test client")
	}
	os.Exit(m.Run())
}

func assertLen(t *testing.T, obj interface{}, expLen int) {
	if !assert.Len(t, obj, expLen) {
		t.FailNow()
	}
}

func assertError(t *testing.T, err error) {
	if !assert.Error(t, err) {
		t.FailNow()
	}
}

func assertNoError(t *testing.T, err error) {
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func assertNil(t *testing.T, i interface{}) {
	if !assert.Nil(t, i) {
		t.FailNow()
	}
}

func assertNotNil(t *testing.T, i interface{}) {
	if !assert.NotNil(t, i) {
		t.FailNow()
	}
}
