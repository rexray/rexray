package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	serverName  = "box-head-us"
	serviceName = "mock"
)

type server struct {
}

func (s *server) Name() string {
	return serverName
}

type service struct {
}

func (s *service) String() string {
	return serviceName
}

func TestJoin(t *testing.T) {

	ctx1 := Background().WithValue(ServerKey, serverName)
	ctx2 := Background().WithValue(ServiceKey, &service{})

	v, ok := Server(ctx1)
	assert.True(t, ok)
	assert.Equal(t, serverName, v)

	v, ok = ServiceName(ctx1)
	assert.False(t, ok)
	assert.Empty(t, v)

	v, ok = Server(ctx2)
	assert.False(t, ok)
	assert.Empty(t, v)

	v, ok = ServiceName(ctx2)
	assert.True(t, ok)
	assert.Equal(t, serviceName, v)

	ctx2 = ctx2.Join(ctx1)

	v, ok = Server(ctx1)
	assert.True(t, ok)
	assert.Equal(t, serverName, v)

	v, ok = ServiceName(ctx1)
	assert.False(t, ok)
	assert.Empty(t, v)

	v, ok = Server(ctx2)
	assert.True(t, ok)
	assert.Equal(t, serverName, v)

	v, ok = ServiceName(ctx2)
	assert.True(t, ok)
	assert.Equal(t, serviceName, v)
}

type driver struct {
}

func (d *driver) Name() string {
	return "libstorage"
}

func TestContextIDLog(t *testing.T) {
	ctx := WithValue(Background(), ServerKey, &server{})
	ctx.Info("no storage driver")
	ctx = WithValue(ctx, DriverKey, &driver{})
	ctx.Info("storage driver set")
}
