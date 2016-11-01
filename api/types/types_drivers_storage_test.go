package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolumeAttachmentsTypes(t *testing.T) {
	var a VolumeAttachmentsTypes
	assert.False(t, a.Requested())
	assert.False(t, a.Mine())
	assert.False(t, a.Devices())
	assert.False(t, a.Attached())

	a = VolumeAttachmentsTypes(5)
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.True(t, a.Devices())
	assert.False(t, a.Attached())

	a = VolumeAttachmentsTypes(9)
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.False(t, a.Devices())
	assert.True(t, a.Attached())

	a = VolumeAttachmentsTypes(13)
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.True(t, a.Devices())
	assert.True(t, a.Attached())
}

func TestParseVolumeAttachmentsTypes(t *testing.T) {
	var a VolumeAttachmentsTypes
	assert.False(t, a.Requested())
	assert.False(t, a.Mine())
	assert.False(t, a.Devices())
	assert.False(t, a.Attached())

	a = ParseVolumeAttachmentTypes("true")
	assert.True(t, a.Requested())
	assert.True(t, a.Mine())
	assert.True(t, a.Devices())
	assert.True(t, a.Attached())
	a = ParseVolumeAttachmentTypes(true)
	assert.True(t, a.Requested())
	assert.True(t, a.Mine())
	assert.True(t, a.Devices())
	assert.True(t, a.Attached())

	a = ParseVolumeAttachmentTypes("5")
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.True(t, a.Devices())
	assert.False(t, a.Attached())
	a = ParseVolumeAttachmentTypes(5)
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.True(t, a.Devices())
	assert.False(t, a.Attached())

	a = ParseVolumeAttachmentTypes("9")
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.False(t, a.Devices())
	assert.True(t, a.Attached())
	a = ParseVolumeAttachmentTypes(9)
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.False(t, a.Devices())
	assert.True(t, a.Attached())

	a = ParseVolumeAttachmentTypes("13")
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.True(t, a.Devices())
	assert.True(t, a.Attached())
	a = ParseVolumeAttachmentTypes(13)
	assert.True(t, a.Requested())
	assert.False(t, a.Mine())
	assert.True(t, a.Devices())
	assert.True(t, a.Attached())
}
