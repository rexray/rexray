package test

import (
	"testing"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/drivers/mock"
)

func TestNewWithBadOSDriver(t *testing.T) {
	r := core.New(nil)
	r.Config.OSDrivers = []string{mock.BadMockOSDriverName}
	r.Config.VolumeDrivers = []string{mock.MockVolDriverName}
	r.Config.StorageDrivers = []string{mock.MockStorDriverName}
	if err := r.InitDrivers(); err != errors.ErrNoOSDrivers {
		t.Fatal(err)
	}
}

func TestNewWithBadVolumeDriver(t *testing.T) {
	r := core.New(nil)
	r.Config.OSDrivers = []string{mock.MockOSDriverName}
	r.Config.VolumeDrivers = []string{mock.BadMockVolDriverName}
	r.Config.StorageDrivers = []string{mock.MockStorDriverName}
	if err := r.InitDrivers(); err != errors.ErrNoVolumeDrivers {
		t.Fatal(err)
	}
}

func TestNewWithBadStorageDriver(t *testing.T) {
	r := core.New(nil)
	r.Config.OSDrivers = []string{mock.MockOSDriverName}
	r.Config.VolumeDrivers = []string{mock.MockVolDriverName}
	r.Config.StorageDrivers = []string{mock.BadMockStorDriverName}
	if err := r.InitDrivers(); err != errors.ErrNoStorageDrivers {
		t.Fatal(err)
	}
}
