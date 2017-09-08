package csi

import (
	"github.com/codedellemc/gocsi/csi"
	dvol "github.com/docker/go-plugins-helpers/volume"

	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
)

var csiVersion = &csi.Version{
	Major: 0,
	Minor: 0,
	Patch: 0,
}

type dockerVolDriver struct {
	ctx apitypes.Context
	cs  *csiService
}

func (d *dockerVolDriver) Create(
	*dvol.CreateRequest) error {
	return nil
}

func (d *dockerVolDriver) List() (*dvol.ListResponse, error) {
	/*d.ctx.Info("docker->csi:  listing volumes")
	res, err := d.cs.ListVolumes(
		d.ctx, &csi.ListVolumesRequest{Version: csiVersion})
	if err != nil {
		d.ctx.WithError(err).Error("docker->csi: failed to list vols")
		return nil, err
	}
	if err := gocsi.CheckResponseErrListVolumes(
		d.ctx, "/csi.Controller/ListVolumes", res); err != nil {
		d.ctx.WithError(err).Error("docker->csi: failed to list vols")
		return nil, err
	}
	dres := &dvol.ListResponse{}
	dres.Volumes = make([]*dvol.Volume, len(res.GetResult().Entries))
	for i, vi := range res.GetResult().Entries {
		v := &dvol.Volume{
			Name: fmt.Sprintf("%v", vi.VolumeInfo.Id),
		}
		dres.Volumes[i] = v
		d.ctx.WithField("volume", v).Info("docker->csi: found volume")
	}
	return dres, nil*/
	return nil, nil
}

func (d *dockerVolDriver) Get(
	*dvol.GetRequest) (*dvol.GetResponse, error) {
	return nil, nil
}

func (d *dockerVolDriver) Remove(*dvol.RemoveRequest) error {
	return nil
}

func (d *dockerVolDriver) Path(
	*dvol.PathRequest) (*dvol.PathResponse, error) {
	return nil, nil
}

func (d *dockerVolDriver) Mount(
	*dvol.MountRequest) (*dvol.MountResponse, error) {
	return nil, nil
}

func (d *dockerVolDriver) Unmount(*dvol.UnmountRequest) error {
	return nil
}

func (d *dockerVolDriver) Capabilities() *dvol.CapabilitiesResponse {
	return nil
}
