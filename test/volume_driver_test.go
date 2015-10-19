package test

import (
	"testing"

	"github.com/emccode/rexray/core"
)

func TestVolumeDriverName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if d.Name() != mockVolDriverName {
		t.Fatalf("driver name != %s, == %s", mockVolDriverName, d.Name())
	}
}

type mockVolDriver struct {
	name string
}

func newVolDriver() core.Driver {
	var d core.VolumeDriver = &mockVolDriver{mockVolDriverName}
	return d
}

func (m *mockVolDriver) Init(r *core.RexRay) error {
	return nil
}

func (m *mockVolDriver) Name() string {
	return m.name
}

func (m *mockVolDriver) Mount(
	volumeName, volumeID string,
	overwriteFs bool, newFsType string) (string, error) {
	return "", nil
}

func (m *mockVolDriver) Unmount(volumeName, volumeID string) error {
	return nil
}

func (m *mockVolDriver) Path(volumeName, volumeID string) (string, error) {
	return "", nil
}

func (m *mockVolDriver) Create(volumeName string, opts core.VolumeOpts) error {
	return nil
}

func (m *mockVolDriver) Remove(volumeName string) error {
	return nil
}

func (m *mockVolDriver) Attach(volumeName, instanceID string) (string, error) {
	return "", nil
}

func (m *mockVolDriver) Detach(volumeName, instanceID string) error {
	return nil
}

func (m *mockVolDriver) NetworkName(
	volumeName, instanceID string) (string, error) {
	return "", nil
}
