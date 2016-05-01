package context

import (
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/types"
)

func TestJoin(t *testing.T) {

	ctx1 := Background().WithValue(types.ContextServerName, "box-head-us")
	ctx2 := Background().WithServiceName("mock")

	assert.Equal(t, "box-head-us", ctx1.ServerName())
	assert.Empty(t, "", ctx1.ServiceName())
	assert.Equal(t, "mock", ctx2.ServiceName())
	assert.Empty(t, "", ctx2.ServerName())

	ctx2 = ctx2.Join(ctx1)

	assert.Equal(t, "box-head-us", ctx1.ServerName())
	assert.Empty(t, ctx1.ServiceName())
	assert.Equal(t, "mock", ctx2.ServiceName())
	assert.Equal(t, "box-head-us", ctx2.ServerName())
}

func TestContextIDLog(t *testing.T) {
	ctx := Background()
	ctx.Info("no storage driver")
	ctx = ctx.WithContextSID(types.ContextStorageDriver, "mock")
	ctx.Info("storage driver set")
}

func TestConfig(t *testing.T) {

	assertEmpty := func(configs ...gofig.Config) {
		for _, c := range configs {
			assert.Empty(t, c.GetString("hello"))
		}
	}

	assertEqual := func(configs ...gofig.Config) {
		for _, c := range configs {
			assert.Equal(t, "world", c.GetString("hello"))
		}
	}

	config := gofig.New()
	ctx := NewContext(Background(), config, nil)
	assertEmpty(config, ctx)

	config.Set("hello", "world")
	assertEqual(config, ctx)

	ctx = ctx.WithServiceName("mock")
	assertEqual(ctx)

	ctx2 := Background().Join(ctx)
	assertEqual(ctx2)
}
