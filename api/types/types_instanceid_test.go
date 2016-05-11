package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	metadataBase64       = "eyJIZWxsbyI6ImhpIiwiVGhlcmUiOiJoZXJlIiwiVmFsdWUiOjMsIkRhdGEiOnsiZjEiOjAsImYyIjoiMSJ9fQ=="
	expectedI1String     = "vfs=1234,false," + metadataBase64
	expectedI1NoIDString = "vfs=,false," + metadataBase64
)

func newMetadata() interface{} {
	return &struct {
		Hello string
		There string
		Value int
		Data  map[string]interface{}
	}{
		"hi", "here", 3, map[string]interface{}{"f1": 0, "f2": "1"},
	}
}

func TestInstanceIDMarshalText(t *testing.T) {

	i1 := &InstanceID{ID: "1234", Driver: "vfs"}
	assert.Equal(t, "vfs=1234", i1.String())
	t.Logf("instanceID=%s", i1)

	i2 := &InstanceID{}
	assert.NoError(t, i2.UnmarshalText([]byte(i1.String())))
	assert.EqualValues(t, i1, i2)

	assert.NoError(t, i1.AddMetadata(newMetadata()))
	assert.Equal(t, expectedI1String, i1.String())
	t.Logf("instanceID=%s", i1)

	i3 := &InstanceID{}
	assert.NoError(t, i3.UnmarshalText([]byte(expectedI1String)))
	assert.EqualValues(t, i1, i3)
}

func TestInstanceIDMarshalJSON(t *testing.T) {
	i1 := &InstanceID{ID: "1234", Driver: "vfs"}
	i1.AddMetadata(newMetadata())

	buf, err := i1.MarshalJSON()
	assert.NoError(t, err)
	t.Logf("instanceID=%s", string(buf))

	i2 := &InstanceID{}
	assert.NoError(t, i2.UnmarshalJSON(buf))

	assert.EqualValues(t, i1, i2)
}

func TestInstanceIDMarshalTextNoID(t *testing.T) {

	i1 := &InstanceID{Driver: "vfs"}
	assert.Equal(t, "vfs=", i1.String())
	t.Logf("instanceID=%s", i1)

	i2 := &InstanceID{}
	assert.NoError(t, i2.UnmarshalText([]byte(i1.String())))
	assert.EqualValues(t, i1, i2)

	assert.NoError(t, i1.AddMetadata(newMetadata()))
	assert.Equal(t, expectedI1NoIDString, i1.String())
	t.Logf("instanceID=%s", i1)

	i3 := &InstanceID{}
	assert.NoError(t, i3.UnmarshalText([]byte(expectedI1NoIDString)))
	assert.EqualValues(t, i1, i3)
}
