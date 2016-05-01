package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/akutz/gotil"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (d *driver) getVolPath(volumeID string) string {
	p := fmt.Sprintf("%s/%s.json", d.volPath, volumeID)
	d.ctx.WithField("vol.path", p).Info(p)
	return p
}

func (d *driver) getVolumeByID(volumeID string) (*types.Volume, error) {
	volJSONPath := d.getVolPath(volumeID)

	if !gotil.FileExists(volJSONPath) {
		return nil, utils.NewNotFoundError(volumeID)
	}

	return readVolume(volJSONPath)
}

func readVolume(path string) (*types.Volume, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	v := &types.Volume{}
	if err := json.NewDecoder(f).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

func (d *driver) writeVolume(v *types.Volume) error {
	volJSONPath := d.getVolPath(v.ID)

	f, err := os.Create(volJSONPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	if err := enc.Encode(v); err != nil {
		return err
	}

	return nil
}

func (d *driver) getVolJSONs() ([]string, error) {
	return filepath.Glob(d.volJSONGlobPatt)
}

func (d *driver) newVolumeID() string {
	return fmt.Sprintf("vfs-%03d", atomic.AddInt64(&d.volCount, 1))
}
