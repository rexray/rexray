package test

import (
	"testing"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

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

func TestVolumeDriverManagerName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if r.Volume.Name() != mockVolDriverName {
		t.Fatalf("driver name != %s, == %s", mockVolDriverName, r.Volume.Name())
	}
}

func TestVolumeDriverManagerNameNoDrivers(t *testing.T) {
	r, _ := getRexRayNoDrivers()
	if r.Volume.Name() != "" {
		t.Fatal("name not empty")
	}
}

func TestUnmountAll(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.UnmountAll(); err != nil {
		t.Fatal(err)
	}
}

func TestUnmountAllNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.UnmountAll(); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func TestRemoveAll(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.RemoveAll(); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveAllNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.RemoveAll(); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func TestDetachAll(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.DetachAll(""); err != nil {
		t.Fatal(err)
	}
}

func TestDetachAllNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.DetachAll(""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Mount(
	volumeName, volumeID string,
	overwriteFs bool, newFsType string) (string, error) {
	return "", nil
}

func TestVolumeDriverMount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.Mount("", "", false, ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerMount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Mount("", "", false, ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerMountNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Mount("", "", false, ""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Unmount(volumeName, volumeID string) error {
	return nil
}

func TestVolumeDriverUnmount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if err := d.Unmount("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerUnmount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Unmount("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerUnmountNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Unmount("", ""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Path(volumeName, volumeID string) (string, error) {
	return "", nil
}

func TestVolumeDriverPath(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.Path("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerPath(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Path("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerPathNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Path("", ""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Create(volumeName string, opts core.VolumeOpts) error {
	return nil
}

func TestVolumeDriverCreate(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if err := d.Create("", nil); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerCreate(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Create("", nil); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerCreateNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Create("", nil); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Remove(volumeName string) error {
	return nil
}

func TestVolumeDriverRemove(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if err := d.Remove(""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerRemove(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Remove(""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerRemoveNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Remove(""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Attach(volumeName, instanceID string) (string, error) {
	return "", nil
}

func TestVolumeDriverAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.Attach("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Attach("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerAttachNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Attach("", ""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) Detach(volumeName, instanceID string) error {
	return nil
}

func TestVolumeDriverDetach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if err := d.Detach("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerDetach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Detach("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerDetachNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Detach("", ""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func (m *mockVolDriver) NetworkName(
	volumeName, instanceID string) (string, error) {
	return "", nil
}

func TestVolumeDriverNetworkName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.NetworkName("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerNetworkName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.NetworkName("", ""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerNetworkNameNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.NetworkName("", ""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}
