package docker

import (
	"fmt"
	"path"
	"strings"

	"github.com/akutz/goof"
	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
)

func (d *driver) getVolumeMountPath(volumeName string) (string, error) {
	if volumeName == "" {
		return "", goof.New("missing volume name")
	}

	return path.Join(d.mountDirPath(), volumeName), nil
}

func (d *driver) volumeInspectByID(
	ctx types.Context,
	volumeID string,
	attachments types.VolumeAttachmentsTypes,
	opts types.Store) (*types.Volume, error) {
	client := context.MustClient(ctx)
	vol, err := client.Storage().VolumeInspect(ctx, volumeID,
		&types.VolumeInspectOpts{
			Attachments: attachments})
	if err != nil {
		return nil, err
	}
	return vol, nil
}

func (d *driver) volumeInspectByIDOrName(
	ctx types.Context,
	volumeID, volumeName string,
	attachments types.VolumeAttachmentsTypes,
	opts types.Store) (*types.Volume, error) {

	if volumeID != "" && volumeName != "" {
		return nil, goof.New("specify either volumeID or volumeName")
	}

	client := context.MustClient(ctx)

	var obj *types.Volume
	if volumeID != "" {
		var err error
		obj, err = d.volumeInspectByID(
			ctx, volumeID, types.VolAttReqTrue, opts)
		if err != nil {
			return nil, err
		}
	} else {
		objs, err := client.Storage().Volumes(ctx, &types.VolumesOpts{
			Attachments: 0})
		if err != nil {
			return nil, err
		}
		for _, o := range objs {
			if strings.EqualFold(volumeName, o.Name) {
				if attachments.Requested() {
					obj, err = d.volumeInspectByID(
						ctx, o.ID, types.VolAttReqTrue, opts)
					if err != nil {
						return nil, err
					}
				} else {
					obj = o
				}
				break
			}
		}
	}

	if obj == nil {
		return nil, utils.NewNotFoundError(
			fmt.Sprintf("volumeID=%s,volumeName=%s", volumeID, volumeName))
	}
	return obj, nil
}

func isErrNotFound(err error) bool {
	switch err.(type) {
	case *types.ErrNotFound:
		return true
	default:
		return false
	}
}

func (d *driver) volumeMountPath(target string) string {
	return path.Join(target, d.volumeRootPath())
}
