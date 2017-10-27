package goisilon

import (
	"context"

	api "github.com/thecodeteam/goisilon/api/v2"
)

type ACL *api.ACL

// GetVolumeACL returns the ACL for a volume.
func (c *Client) GetVolumeACL(
	ctx context.Context,
	volumeName string) (ACL, error) {

	return api.ACLInspect(ctx, c.API, volumeName)
}

// SetVolumeOwnerToCurrentUser sets the owner for a volume to the user that
// was used to connect to the API.
func (c *Client) SetVolumeOwnerToCurrentUser(
	ctx context.Context,
	volumeName string) error {

	return c.SetVolumeOwner(ctx, volumeName, c.API.User())
}

// SetVolumeOwner sets the owner for a volume.
func (c *Client) SetVolumeOwner(
	ctx context.Context,
	volumeName, userName string) error {

	mode := api.FileMode(0777)

	return api.ACLUpdate(
		ctx,
		c.API,
		volumeName,
		&api.ACL{
			Action:        &api.PActionTypeReplace,
			Authoritative: &api.PAuthoritativeTypeMode,
			Owner: &api.Persona{
				ID: &api.PersonaID{
					ID:   userName,
					Type: api.PersonaIDTypeUser,
				},
			},
			Mode: &mode,
		})
}
