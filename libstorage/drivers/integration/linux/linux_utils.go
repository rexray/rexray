package linux

import (
	"fmt"
	"path"
	"strings"

	"github.com/akutz/goof"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
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

func (d *driver) volumeInspectByName(
	ctx types.Context,
	volumeName string,
	attachments types.VolumeAttachmentsTypes,
	opts types.Store) (*types.Volume, error) {
	client := context.MustClient(ctx)
	if sd, ok := client.Storage().(types.StorageDriverVolInspectByName); ok {
		vol, err := sd.VolumeInspectByName(
			ctx,
			volumeName,
			&types.VolumeInspectOpts{
				Attachments: attachments})
		if err != nil {
			return nil, err
		}
		return vol, nil
	}

	vols, err := client.Storage().Volumes(
		ctx, &types.VolumesOpts{Attachments: 0})
	if err != nil {
		return nil, err
	}
	for _, v := range vols {
		if strings.EqualFold(volumeName, v.Name) {
			vol, err := d.volumeInspectByID(
				ctx, v.ID, attachments, opts)
			if err != nil {
				return nil, err
			}
			return vol, nil
		}
	}

	return nil, nil
}

func (d *driver) volumeInspectByIDOrName(
	ctx types.Context,
	volumeID, volumeName string,
	attachments types.VolumeAttachmentsTypes,
	opts types.Store) (*types.Volume, error) {

	if volumeID != "" && volumeName != "" {
		return nil, goof.New("specify either volumeID or volumeName")
	}

	var (
		obj *types.Volume
		err error
	)

	if volumeID != "" {
		obj, err = d.volumeInspectByID(
			ctx, volumeID, attachments, opts)
	} else {
		obj, err = d.volumeInspectByName(
			ctx, volumeName, attachments, opts)
	}
	if err != nil {
		return nil, err
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
