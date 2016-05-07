package context

import (
	"testing"

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
