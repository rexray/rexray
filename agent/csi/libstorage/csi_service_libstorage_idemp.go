package libstorage

import (
	"errors"
	"fmt"
	"os"

	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/gocsi/mount"
	xctx "golang.org/x/net/context"

	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiutils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

var (
	errMissingIDKeyPath   = errors.New("missing id key path")
	errMissingTokenKey    = errors.New("missing token key")
	errUnableToGetLocDevs = errors.New("unable to get local devices")
	errLocDevsMissingVol  = errors.New("unable to find attached vol in local devices")
	errMissingTargetPath  = errors.New("target path not created")
)

const resNotFound = "resource not found"

func isNotFoundErr(err error) bool {
	return err.Error() == resNotFound
}

// GetVolumeName should return the name of the volume specified
// by the provided volume ID. If the volume does not exist then
// an empty string should be returned.
func (d *driver) GetVolumeName(
	ctx xctx.Context,
	id *csi.VolumeID) (string, error) {

	idVal, ok := id.Values["id"]
	if !ok {
		return "", errMissingIDKeyPath
	}

	opts := &apitypes.VolumeInspectOpts{
		Opts: apiutils.NewStore(),
	}

	vol, err := d.client.Storage().VolumeInspect(d.ctx, idVal, opts)
	if err != nil {

		// If the volume is not found then return an empty string
		// for the name to indicate such.
		if isNotFoundErr(err) {
			return "", nil
		}

		return "", err
	}

	return vol.Name, nil
}

// GetVolumeInfo should return information about the volume
// specified by the provided volume name. If the volume does not
// exist then a nil value should be returned.
func (d *driver) GetVolumeInfo(
	ctx xctx.Context,
	name string) (*csi.VolumeInfo, error) {

	td, ok := d.client.Storage().(apitypes.StorageDriverVolInspectByName)
	if !ok {
		return nil, fmt.Errorf(
			"stor driver not by name: %T", d.client.Storage())
	}

	opts := &apitypes.VolumeInspectOpts{
		Opts: apiutils.NewStore(),
	}

	vol, err := td.VolumeInspectByName(d.ctx, name, opts)
	if err != nil {

		// If the volume is not found then return nil for the
		// volume info to indicate such.
		if isNotFoundErr(err) {
			return nil, nil
		}

		return nil, err
	}

	return toVolumeInfo(vol), nil
}

// IsControllerPublished should return publication info about
// the volume specified by the provided volume name or ID.
func (d *driver) IsControllerPublished(
	ctx xctx.Context,
	id *csi.VolumeID) (*csi.PublishVolumeInfo, error) {

	idVal, ok := id.Values["id"]
	if !ok {
		return nil, errMissingIDKeyPath
	}

	// Request only volumes that are attached.
	viOpts := &apitypes.VolumeInspectOpts{
		Attachments: apitypes.VolAttReqWithDevMapForInstance,
		Opts:        apiutils.NewStore(),
	}

	vol, err := d.client.Storage().VolumeInspect(d.ctx, idVal, viOpts)
	if err != nil {
		return nil, err
	}

	// If the volume is not attached to this node then do not
	// indicate an error; just return nil to indicate
	// the volume is not attached to this node.
	if vol.AttachmentState != apitypes.VolumeAttached {
		return nil, nil
	}

	pvi := &csi.PublishVolumeInfo{
		Values: map[string]string{
			"encrypted": fmt.Sprintf("%v", vol.Encrypted),
		},
	}

	// We know that the volume is attached, we need to go get it's token
	// via local devices
	ldOpts := &apitypes.LocalDevicesOpts{
		Opts:     apiutils.NewStore(),
		ScanType: apitypes.DeviceScanQuick,
	}
	devs, err := d.client.Executor().LocalDevices(d.ctx, ldOpts)
	if err != nil {
		return nil, errUnableToGetLocDevs
	}

	for k, v := range devs.DeviceMap {
		if v == vol.Attachments[0].DeviceName {
			pvi.Values["token"] = k
			break
		}
	}

	if _, ok := pvi.Values["token"]; !ok {
		// Token was not found via local devices
		return nil, errLocDevsMissingVol
	}

	return pvi, nil
}

// IsNodePublished should return a flag indicating whether or
// not the volume exists and is published on the current host.
func (d *driver) IsNodePublished(
	ctx xctx.Context,
	id *csi.VolumeID,
	pubInfo *csi.PublishVolumeInfo,
	targetPath string) (bool, error) {

	var devPath string

	st, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = errMissingTargetPath
		}
		return false, err
	}

	volTypeIsMount := st.IsDir()

	if pubInfo != nil {
		token, ok := pubInfo.Values["token"]
		if !ok {
			return false, errMissingTokenKey
		}

		// Get device from local devices
		opts := &apitypes.LocalDevicesOpts{
			Opts:     apiutils.NewStore(),
			ScanType: apitypes.DeviceScanQuick,
		}
		devs, lderr := d.client.Executor().LocalDevices(d.ctx, opts)
		if lderr != nil {
			return false, errUnableToGetLocDevs
		}

		devPath, ok = devs.DeviceMap[token]
		if !ok {
			// device not in device map yet. That may not be an error, as
			// it may not have shown up yet. Defer to lower-level publish
			return false, nil
		}
	} else {
		// pubInfo is nil, all we have is the ID, so we will have to
		// make a libStorage inspect call to get the underlying device

		idVal, ok := id.Values["id"]
		if !ok {
			return false, errMissingIDKeyPath
		}

		// Request dev map if volume attached to this instance.
		opts := &apitypes.VolumeInspectOpts{
			Attachments: apitypes.VolAttReqWithDevMapForInstance,
			Opts:        apiutils.NewStore(),
		}

		vol, vierr := d.client.Storage().VolumeInspect(d.ctx, idVal, opts)
		if vierr != nil {
			return false, err
		}

		// If the volume is not attached to this node then do not
		// indicate an error; just return false to indicate
		// the volume is not attached to this node.
		if vol.AttachmentState != apitypes.VolumeAttached {
			return false, nil
		}

		// If the volume has no attachments then it's not possible to
		// determine the node publication status.
		if len(vol.Attachments) == 0 {
			return false, nil
		}

		devPath = vol.Attachments[0].DeviceName
	}

	// Get the local mount table.
	minfo, err := mount.GetMounts()
	if err != nil {
		return false, err
	}

	// Scan the mount table and get the path to which the device of
	// the attached volume is mounted.
	if volTypeIsMount {
		for _, mi := range minfo {
			if mi.Device == devPath && mi.Path == targetPath {
				return true, nil
			}
		}
	} else {
		for _, mi := range minfo {
			if mi.Device == devtmpfs && mi.Path == targetPath {
				return true, nil
			}
		}
	}

	// If no mount was discovered then indicate the volume is not
	// published on this node.
	return false, nil
}
