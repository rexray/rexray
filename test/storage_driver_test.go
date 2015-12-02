package test

import (
	"testing"

	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/drivers/mock"
)

func TestStorageDriverName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if d.Name() != mock.MockStorDriverName {
		t.Fatalf("driver name != %s, == %s", mock.MockStorDriverName, d.Name())
	}
}

func TestStorageDriverManagerName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if r.Storage.Name() != mock.MockStorDriverName {
		t.Fatalf("driver name != %s, == %s", mock.MockStorDriverName, r.Storage.Name())
	}
}

func TestStorageDriverManagerNameNoDrivers(t *testing.T) {
	r, _ := getRexRayNoDrivers()
	if r.Storage.Name() != "" {
		t.Fatal("name not empty")
	}
}

func TestStorageDriverGetVolumeMapping(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetVolumeMapping(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeMapping(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeMapping(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeMappingNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeMapping(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverGetInstance(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetInstance(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetInstance(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstance(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetInstanceNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstance(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestGetInstances(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstances(); err != nil {
		t.Fatal(err)
	}
}

func TestGetInstancesNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetInstances(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverGetVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetVolume(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolume(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolume(
		"", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverGetVolumeAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetVolumeAttach(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeAttach(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeAttach(
		"", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetVolumeAttachNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetVolumeAttach(
		"", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverCreateSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.CreateSnapshot(
		false, "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateSnapshot(
		false, "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateSnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateSnapshot(
		false, "", "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverGetSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetSnapshot(
		"", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetSnapshot(
		"", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetSnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetSnapshot(
		"", "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverRemoveSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if err := d.RemoveSnapshot(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveSnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveSnapshot(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveSnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveSnapshot(""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverCreateVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.CreateVolume(
		false, "", "", "", "", 0, 0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateVolume(
		false, "", "", "", "", 0, 0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCreateVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CreateVolume(
		false, "", "", "", "", 0, 0, ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverRemoveVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if err := d.RemoveVolume(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveVolume(""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerRemoveVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.RemoveVolume(""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverGetDeviceNextAvailable(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.GetDeviceNextAvailable(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetDeviceNextAvailable(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetDeviceNextAvailable(); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerGetDeviceNextAvailableNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.GetDeviceNextAvailable(); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverAttachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.AttachVolume(
		false, "", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerAttachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.AttachVolume(
		false, "", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerAttachVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.AttachVolume(
		false, "", "", false); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverDetachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if err := d.DetachVolume(
		false, "", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerDetachVolume(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.DetachVolume(
		false, "", "", false); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerDetachVolumeNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Storage.DetachVolume(
		false, "", "", false); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}

func TestStorageDriverCopySnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.Storage.Drivers()
	if _, err := d.CopySnapshot(false, "", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCopySnapshot(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CopySnapshot(false, "", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
}

func TestStorageDriverManagerCopySnapshotNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Storage.CopySnapshot(
		false, "", "", "", "", ""); err != errors.ErrNoStorageDetected {
		t.Fatal(err)
	}
}
