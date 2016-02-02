package test

import (
	"testing"

	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/drivers/mock"
)

func TestVolumeDriverName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if d.Name() != mock.MockVolDriverName {
		t.Fatalf("driver name != %s, == %s", mock.MockVolDriverName, d.Name())
	}
}

func TestVolumeDriverManagerName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if r.Volume.Name() != mock.MockVolDriverName {
		t.Fatalf("driver name != %s, == %s", mock.MockVolDriverName, r.Volume.Name())
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

func TestVolumeDriverMount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.Mount("", "", false, "", false); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerMount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Mount("", "", false, "", false); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerMountNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Mount("", "", false, "", false); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
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

func TestVolumeDriverGet(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.Get(""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerGet(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Get(""); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerGetNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Get(""); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func TestVolumeDriverList(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.List(); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerList(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.List(); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerListNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.List(); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
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

func TestVolumeDriverAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if _, err := d.Attach("", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Attach("", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerAttachNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Volume.Attach("", "", false); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
}

func TestVolumeDriverDetach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Volume.Drivers()
	if err := d.Detach("", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerDetach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Detach("", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestVolumeDriverManagerDetachNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Volume.Detach("", "", false); err != errors.ErrNoVolumesDetected {
		t.Fatal(err)
	}
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
