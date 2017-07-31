package goisilon

import (
	"testing"

	"github.com/stretchr/testify/assert"

	api "github.com/codedellemc/goisilon/api/v2"
)

func TestGetVolumeACL(t *testing.T) {
	volumeName := "test_get_volume_acl"

	// make sure the volume exists
	client.CreateVolume(defaultCtx, volumeName)
	volume, err := client.GetVolume(defaultCtx, volumeName, volumeName)
	assertNoError(t, err)
	assertNotNil(t, volume)

	defer client.DeleteVolume(defaultCtx, volume.Name)

	acl, err := client.GetVolumeACL(defaultCtx, volume.Name)
	assertNoError(t, err)
	assertNotNil(t, acl)

	assertNotNil(t, acl.Owner)
	assertNotNil(t, acl.Owner.Name)
	assert.Equal(t, client.API.User(), *acl.Owner.Name)
	assertNotNil(t, acl.Owner.Type)
	assert.Equal(t, api.PersonaTypeUser, *acl.Owner.Type)
	assertNotNil(t, acl.Owner.ID)
	assert.Equal(t, "10", acl.Owner.ID.ID)
	assert.Equal(t, api.PersonaIDTypeUID, acl.Owner.ID.Type)
}

func TestSetVolumeOwnerToCurrentUser(t *testing.T) {
	volumeName := "test_set_volume_owner"

	// make sure the volume exists
	client.CreateVolume(defaultCtx, volumeName)
	volume, err := client.GetVolume(defaultCtx, volumeName, volumeName)
	assertNoError(t, err)
	assertNotNil(t, volume)

	defer client.DeleteVolume(defaultCtx, volume.Name)

	acl, err := client.GetVolumeACL(defaultCtx, volume.Name)
	assertNoError(t, err)
	assertNotNil(t, acl)

	assertNotNil(t, acl.Owner)
	assertNotNil(t, acl.Owner.Name)
	assert.Equal(t, client.API.User(), *acl.Owner.Name)
	assertNotNil(t, acl.Owner.Type)
	assert.Equal(t, api.PersonaTypeUser, *acl.Owner.Type)
	assertNotNil(t, acl.Owner.ID)
	assert.Equal(t, "10", acl.Owner.ID.ID)
	assert.Equal(t, api.PersonaIDTypeUID, acl.Owner.ID.Type)

	err = client.SetVolumeOwner(defaultCtx, volume.Name, "rexray")
	assertNoError(t, err)

	acl, err = client.GetVolumeACL(defaultCtx, volume.Name)
	assertNoError(t, err)
	assertNotNil(t, acl)

	assertNotNil(t, acl.Owner)
	assertNotNil(t, acl.Owner.Name)
	assert.Equal(t, "rexray", *acl.Owner.Name)
	assertNotNil(t, acl.Owner.Type)
	assert.Equal(t, api.PersonaTypeUser, *acl.Owner.Type)
	assertNotNil(t, acl.Owner.ID)
	assert.Equal(t, "2000", acl.Owner.ID.ID)
	assert.Equal(t, api.PersonaIDTypeUID, acl.Owner.ID.Type)

	err = client.SetVolumeOwnerToCurrentUser(defaultCtx, volume.Name)
	assertNoError(t, err)

	acl, err = client.GetVolumeACL(defaultCtx, volume.Name)
	assertNoError(t, err)
	assertNotNil(t, acl)

	assertNotNil(t, acl.Owner)
	assertNotNil(t, acl.Owner.Name)
	assert.Equal(t, client.API.User(), *acl.Owner.Name)
	assertNotNil(t, acl.Owner.Type)
	assert.Equal(t, api.PersonaTypeUser, *acl.Owner.Type)
	assertNotNil(t, acl.Owner.ID)
	assert.Equal(t, "10", acl.Owner.ID.ID)
	assert.Equal(t, api.PersonaIDTypeUID, acl.Owner.ID.Type)
}
