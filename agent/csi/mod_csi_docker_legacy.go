package csi

import (
	"fmt"

	gofig "github.com/akutz/gofig/types"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiutils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
	dvol "github.com/docker/go-plugins-helpers/volume"
)

type dockerLegacy struct {
	ctx    apitypes.Context
	config gofig.Config
	lsc    apitypes.Client
}

func newDockerLegacy(
	ctx apitypes.Context,
	config gofig.Config,
	lsc apitypes.Client) (*dockerLegacy, error) {

	endpoint := &dockerLegacy{
		ctx:    ctx,
		config: config,
		lsc:    lsc,
	}

	ctx.Info("docker-legacy: created Docker legacy endpoint")
	return endpoint, nil
}

func (d *dockerLegacy) Create(req *dvol.CreateRequest) error {

	store := apiutils.NewStoreWithVars(req.Options)
	volType := store.GetStringPtr("type")
	if volType == nil {
		volType = store.GetStringPtr("volumetype")
	}

	_, err := d.lsc.Integration().Create(
		d.ctx,
		req.Name,
		&apitypes.VolumeCreateOpts{
			AvailabilityZone: store.GetStringPtr("availabilityZone"),
			IOPS:             store.GetInt64Ptr("iops"),
			Size:             store.GetInt64Ptr("size"),
			Type:             volType,
			Encrypted:        store.GetBoolPtr("encrypted"),
			Opts:             store,
		})

	if err != nil {
		err = fmt.Errorf(
			"docker-legacy: Create: %s: failed: %v",
			req.Name, err)
		d.ctx.Error(err)
		return err
	}

	return nil
}

func (d *dockerLegacy) List() (*dvol.ListResponse, error) {

	vols, err := d.lsc.Integration().List(d.ctx, apiutils.NewStore())

	if err != nil {
		err = fmt.Errorf("docker-legacy: List: failed: %v", err)
		d.ctx.Error(err)
		return nil, err
	}

	res := &dvol.ListResponse{Volumes: make([]*dvol.Volume, len(vols))}
	for i, v := range vols {
		res.Volumes[i] = d.toDockerVolume(v)
	}

	return res, nil
}

func (d *dockerLegacy) Get(req *dvol.GetRequest) (*dvol.GetResponse, error) {

	vol, err := d.lsc.Integration().Inspect(
		d.ctx,
		req.Name,
		apiutils.NewStore())

	if err != nil {
		err = fmt.Errorf(
			"docker-legacy: Get: %s: failed: %v",
			req.Name, err)
		d.ctx.Error(err)
		return nil, err
	}

	return &dvol.GetResponse{Volume: d.toDockerVolume(vol)}, nil
}

func (d *dockerLegacy) Remove(req *dvol.RemoveRequest) error {

	err := d.lsc.Integration().Remove(
		d.ctx,
		req.Name,
		&apitypes.VolumeRemoveOpts{Opts: apiutils.NewStore()})

	if err != nil {
		err = fmt.Errorf(
			"docker-legacy: Remove: %s: failed: %v",
			req.Name, err)
		d.ctx.Error(err)
		return err
	}

	return nil
}

func (d *dockerLegacy) Path(req *dvol.PathRequest) (*dvol.PathResponse, error) {

	mountPath, err := d.lsc.Integration().Path(
		d.ctx,
		"",
		req.Name,
		apiutils.NewStore())

	if err != nil {
		err = fmt.Errorf(
			"docker-legacy: Path: %s: failed: %v",
			req.Name, err)
		d.ctx.Error(err)
		return nil, err
	}

	return &dvol.PathResponse{Mountpoint: mountPath}, nil
}

func (d *dockerLegacy) Mount(
	req *dvol.MountRequest) (*dvol.MountResponse, error) {

	mountPath, _, err := d.lsc.Integration().Mount(
		d.ctx,
		"",
		req.Name,
		&apitypes.VolumeMountOpts{})

	if err != nil {
		err = fmt.Errorf(
			"docker-legacy: Mount: %s: failed: %v",
			req.Name, err)
		d.ctx.Error(err)
		return nil, err
	}

	return &dvol.MountResponse{Mountpoint: mountPath}, nil
}

func (d *dockerLegacy) Unmount(req *dvol.UnmountRequest) error {

	_, err := d.lsc.Integration().Unmount(
		d.ctx,
		"",
		req.Name,
		apiutils.NewStore())

	if err != nil {
		err = fmt.Errorf(
			"docker-legacy: Unmount: %s: failed: %v",
			req.Name, err)
		d.ctx.Error(err)
		return err
	}

	return nil
}

func (d *dockerLegacy) Capabilities() *dvol.CapabilitiesResponse {
	res := &dvol.CapabilitiesResponse{}
	res.Capabilities.Scope = "global"
	return res
}

func (d *dockerLegacy) toDockerVolume(v apitypes.VolumeMapping) *dvol.Volume {
	return &dvol.Volume{
		Name:       v.VolumeName(),
		Mountpoint: v.MountPoint(),
		Status:     v.Status(),
	}
}
