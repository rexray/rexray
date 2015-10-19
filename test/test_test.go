package test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/emccode/rexray"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/util"
)

const (
	mockOSDriverName   = "mockOSDriver"
	mockVolDriverName  = "mockVolumeDriver"
	mockStorDriverName = "mockStorageDriver"
)

func TestMain(m *testing.M) {
	core.RegisterDriver(mockOSDriverName, newOSDriver)
	core.RegisterDriver(mockVolDriverName, newVolDriver)
	core.RegisterDriver(mockStorDriverName, newStorDriver)
	os.Exit(m.Run())
}

func getRexRay() (*core.RexRay, error) {
	c := config.New()
	c.OSDrivers = []string{mockOSDriverName}
	c.VolumeDrivers = []string{mockVolDriverName}
	c.StorageDrivers = []string{mockStorDriverName}
	r, err := core.New(c)
	if err != nil {
		return nil, err
	}

	if err := r.InitDrivers(); err != nil {
		return nil, err
	}

	return r, nil
}

func getRexRayNoDrivers() (*core.RexRay, error) {
	c := config.New()
	c.OSDrivers = []string{""}
	c.VolumeDrivers = []string{""}
	c.StorageDrivers = []string{""}
	r, err := core.New(c)
	if err != nil {
		return nil, err
	}

	r.InitDrivers()

	return r, nil
}

func TestNewWithConfig(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNew(t *testing.T) {
	os.Setenv("REXRAY_OSDRIVERS", mockOSDriverName)
	os.Setenv("REXRAY_VOLUMEDRIVERS", mockVolDriverName)
	os.Setenv("REXRAY_STORAGEDRIVERS", mockStorDriverName)

	r, err := rexray.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNewNoOSDrivers(t *testing.T) {
	c := config.New()
	c.OSDrivers = []string{}
	c.VolumeDrivers = []string{mockVolDriverName}
	c.StorageDrivers = []string{mockStorDriverName}
	r, err := core.New(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.InitDrivers(); err != errors.ErrNoOSDrivers {
		t.Fatal(err)
	}
}

func TestNewNoVolumeDrivers(t *testing.T) {
	c := config.New()
	c.OSDrivers = []string{mockOSDriverName}
	c.VolumeDrivers = []string{}
	c.StorageDrivers = []string{mockStorDriverName}
	r, err := core.New(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.InitDrivers(); err != errors.ErrNoVolumeDrivers {
		t.Fatal(err)
	}
}

func TestNewNoStorageDrivers(t *testing.T) {
	c := config.New()
	c.OSDrivers = []string{mockOSDriverName}
	c.VolumeDrivers = []string{mockVolDriverName}
	c.StorageDrivers = []string{}
	r, err := core.New(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.InitDrivers(); err != errors.ErrNoStorageDrivers {
		t.Fatal(err)
	}
}

func TestNewWithEnv(t *testing.T) {
	r, err := rexray.NewWithEnv(map[string]string{
		"REXRAY_OSDRIVERS":      mockOSDriverName,
		"REXRAY_VOLUMEDRIVERS":  mockVolDriverName,
		"REXRAY_STORAGEDRIVERS": mockStorDriverName,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNewWithConfigFile(t *testing.T) {
	var err error
	var tmp *os.File
	if tmp, err = ioutil.TempFile("", "TestNewWithConfigFile"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(yamlConfig1); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	r, err := rexray.NewWithConfigFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if err = r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNewWithConfigReader(t *testing.T) {
	r, err := rexray.NewWithConfigReader(bytes.NewReader(yamlConfig1))

	if err != nil {
		t.Fatal(err)
	}

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestDriverNames(t *testing.T) {
	allDriverNames := []string{
		strings.ToLower(mockOSDriverName),
		strings.ToLower(mockVolDriverName),
		strings.ToLower(mockStorDriverName),
		"linux",
		"docker",
		"ec2",
		"openstack",
		"rackspace",
		"scaleio",
		"xtremio",
	}

	var regDriverNames []string
	for dn := range core.DriverNames() {
		regDriverNames = append(regDriverNames, strings.ToLower(dn))
	}

	for _, n := range allDriverNames {
		if !util.StringInSlice(n, regDriverNames) {
			t.Fail()
		}
	}

	for _, n := range regDriverNames {
		if !util.StringInSlice(n, allDriverNames) {
			t.Fail()
		}
	}
}

func assertDriverNames(t *testing.T, r *core.RexRay) {
	od := <-r.OS.Drivers()
	if od.Name() != mockOSDriverName {
		t.Fatalf("expected %s but was %s", mockOSDriverName, od.Name())
	}

	vd := <-r.Volume.Drivers()
	if vd.Name() != mockVolDriverName {
		t.Fatalf("expected %s but was %s", mockVolDriverName, vd.Name())
	}

	sd := <-r.Storage.Drivers()
	if sd.Name() != mockStorDriverName {
		t.Fatalf("expected %s but was %s", mockStorDriverName, sd.Name())
	}
}

var yamlConfig1 = []byte(`logLevel: error
osDrivers:
- mockOSDriver
volumeDrivers:
- mockVolumeDriver
storageDrivers:
- mockStorageDriver
`)
