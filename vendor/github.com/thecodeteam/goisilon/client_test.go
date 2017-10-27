package goisilon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert.NotNil(t, client)
	assert.NotZero(t, client.API.APIVersion())
	t.Logf("api version=%d", client.API.APIVersion())
}
