package docker

import (
	"fmt"
	"path"
	"strings"

	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (d *driver) getVolumeMountPath(volumeName string) (string, error) {
	if volumeName == "" {
		return "", goof.New("missing volume name")
	}

	return path.Join(d.mountDirPath, volumeName), nil
}

func (d *driver) inspectByIDOrName(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	objs, err := ctx.Client().Storage().Volumes(ctx, &types.VolumesOpts{})
	if err != nil {
		return nil, err
	}

	var obj *types.Volume
	for _, o := range objs {
		if strings.ToLower(volumeID) == strings.ToLower(o.ID) ||
			strings.ToLower(volumeName) == strings.ToLower(o.Name) {
			obj = o
			break
		}
	}

	if obj == nil {
		return nil, utils.NewNotFoundError(
			fmt.Sprintf("volumeID=%s,volumeName=%s", volumeID, volumeName))
	}

	return obj, nil
}

func (d *driver) volumeMountPath(target string) string {
	return path.Join(target, d.volumeRootPath())
}
