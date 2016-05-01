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

func (d *driver) getSnapPath(snapshotID string) string {
	return fmt.Sprintf("%s/%s.json", d.snapPath, snapshotID)
}

func (d *driver) getSnapshotByID(snapshotID string) (*types.Snapshot, error) {
	snapJSONPath := d.getSnapPath(snapshotID)

	if !gotil.FileExists(snapJSONPath) {
		return nil, utils.NewNotFoundError(snapshotID)
	}

	return readSnapshot(snapJSONPath)
}

func readSnapshot(path string) (*types.Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	v := &types.Snapshot{}
	if err := json.NewDecoder(f).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

func (d *driver) writeSnapshot(s *types.Snapshot) error {
	snapJSONPath := d.getSnapPath(s.ID)

	f, err := os.Create(snapJSONPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	if err := enc.Encode(s); err != nil {
		return err
	}

	return nil
}

func (d *driver) getSnapJSONs() ([]string, error) {
	return filepath.Glob(d.snapJSONGlobPatt)
}

func (d *driver) newSnapshotID(volumeID string) string {
	return fmt.Sprintf(
		"%s-%03d", volumeID, atomic.AddInt64(&d.snapCount, 1))
}
