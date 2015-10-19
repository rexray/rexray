package test

import (
	"testing"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

func TestNewWithBadOSDriver(t *testing.T) {
	r := core.New(nil)
	r.Config.OSDrivers = []string{badMockOSDriverName}
	r.Config.VolumeDrivers = []string{mockVolDriverName}
	r.Config.StorageDrivers = []string{mockStorDriverName}
	if err := r.InitDrivers(); err != errors.ErrNoOSDrivers {
		t.Fatal(err)
	}
}

func TestNewWithBadVolumeDriver(t *testing.T) {
	r := core.New(nil)
	r.Config.OSDrivers = []string{mockOSDriverName}
	r.Config.VolumeDrivers = []string{badMockVolDriverName}
	r.Config.StorageDrivers = []string{mockStorDriverName}
	if err := r.InitDrivers(); err != errors.ErrNoVolumeDrivers {
		t.Fatal(err)
	}
}

func TestNewWithBadStorageDriver(t *testing.T) {
	r := core.New(nil)
	r.Config.OSDrivers = []string{mockOSDriverName}
	r.Config.VolumeDrivers = []string{mockVolDriverName}
	r.Config.StorageDrivers = []string{badMockStorDriverName}
	if err := r.InitDrivers(); err != errors.ErrNoStorageDrivers {
		t.Fatal(err)
	}
}

type badMockOSDriver struct {
	mockOSDriver
}

func newBadOSDriver() core.Driver {
	var d core.OSDriver = &badMockOSDriver{
		mockOSDriver{badMockOSDriverName}}
	return d
}

func (m *badMockOSDriver) Init(r *core.RexRay) error {
	return errors.New("init error")
}

type badMockVolDriver struct {
	mockVolDriver
}

func newBadVolDriver() core.Driver {
	var d core.VolumeDriver = &badMockVolDriver{
		mockVolDriver{badMockVolDriverName}}
	return d
}

func (m *badMockVolDriver) Init(r *core.RexRay) error {
	return errors.New("init error")
}

type badMockStorDriver struct {
	mockStorDriver
}

func newBadStorDriver() core.Driver {
	var d core.StorageDriver = &badMockStorDriver{
		mockStorDriver{badMockStorDriverName}}
	return d
}

func (m *badMockStorDriver) Init(r *core.RexRay) error {
	return errors.New("init error")
}
