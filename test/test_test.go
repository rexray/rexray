package test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	"github.com/emccode/rexray"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/drivers/mock"
)

func TestMain(m *testing.M) {
	mock.RegisterMockDrivers()
	mock.RegisterBadMockDrivers()
	os.Exit(m.Run())
}

func getRexRay() (*core.RexRay, error) {
	c := gofig.New()
	c.Set("rexray.osDrivers", []string{mock.MockOSDriverName})
	c.Set("rexray.volumeDrivers", []string{mock.MockVolDriverName})
	c.Set("rexray.storageDrivers", []string{mock.MockStorDriverName})
	r := core.New(c)

	if err := r.InitDrivers(); err != nil {
		return nil, err
	}

	return r, nil
}

func getRexRayNoDrivers() (*core.RexRay, error) {
	c := gofig.New()
	c.Set("rexray.osDrivers", []string{""})
	c.Set("rexray.volumeDrivers", []string{""})
	c.Set("rexray.storageDrivers", []string{""})
	r := core.New(c)
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

func TestNewWithNilConfig(t *testing.T) {
	r := core.New(nil)
	r.Config.Set("rexray.osDrivers", []string{mock.MockOSDriverName})
	r.Config.Set("rexray.volumeDrivers", []string{mock.MockVolDriverName})
	r.Config.Set("rexray.storageDrivers", []string{mock.MockStorDriverName})

	if err := r.InitDrivers(); err != nil {
		t.Fatal(err)
	}

	assertDriverNames(t, r)
}

func TestNew(t *testing.T) {
	os.Setenv("REXRAY_OSDRIVERS", mock.MockOSDriverName)
	os.Setenv("REXRAY_VOLUMEDRIVERS", mock.MockVolDriverName)
	os.Setenv("REXRAY_STORAGEDRIVERS", mock.MockStorDriverName)

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
	c := gofig.New()
	c.Set("rexray.osDrivers", []string{})
	c.Set("rexray.volumeDrivers", []string{mock.MockVolDriverName})
	c.Set("rexray.storageDrivers", []string{mock.MockStorDriverName})
	r := core.New(c)
	if err := r.InitDrivers(); err != errors.ErrNoOSDrivers {
		t.Fatal(err)
	}
}

func TestNewNoVolumeDrivers(t *testing.T) {
	c := gofig.New()
	c.Set("rexray.osDrivers", []string{mock.MockOSDriverName})
	c.Set("rexray.volumeDrivers", []string{})
	c.Set("rexray.storageDrivers", []string{mock.MockStorDriverName})
	r := core.New(c)
	if err := r.InitDrivers(); err != errors.ErrNoVolumeDrivers {
		t.Fatal(err)
	}
}

func TestNewNoStorageDrivers(t *testing.T) {
	c := gofig.New()
	c.Set("rexray.osDrivers", []string{mock.MockOSDriverName})
	c.Set("rexray.volumeDrivers", []string{mock.MockVolDriverName})
	c.Set("rexray.storageDrivers", []string{})
	r := core.New(c)
	if err := r.InitDrivers(); err != errors.ErrNoStorageDrivers {
		t.Fatal(err)
	}
}

func TestNewWithEnv(t *testing.T) {
	r, err := rexray.NewWithEnv(map[string]string{
		"REXRAY_OSDRIVERS":      mock.MockOSDriverName,
		"REXRAY_VOLUMEDRIVERS":  mock.MockVolDriverName,
		"REXRAY_STORAGEDRIVERS": mock.MockStorDriverName,
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

func TestNewWithBadConfigFilePath(t *testing.T) {
	if _, err := rexray.NewWithConfigFile(gotil.RandomString(10)); err == nil {
		t.Fatal("expected error from bad config file path")
	}
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

func TestNewWithBadConfigReader(t *testing.T) {
	if _, err := rexray.NewWithConfigReader(nil); err == nil {
		t.Fatal("expected error from bad config reader")
	}
}

func TestDriverNames(t *testing.T) {
	allDriverNames := []string{
		strings.ToLower(mock.MockOSDriverName),
		strings.ToLower(mock.MockVolDriverName),
		strings.ToLower(mock.MockStorDriverName),
		strings.ToLower(mock.BadMockOSDriverName),
		strings.ToLower(mock.BadMockVolDriverName),
		strings.ToLower(mock.BadMockStorDriverName),
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
		if !gotil.StringInSlice(n, regDriverNames) {
			t.Fail()
		}
	}

	for _, n := range regDriverNames {
		if !gotil.StringInSlice(n, allDriverNames) {
			t.Fail()
		}
	}
}

func TestRexRayDriverNames(t *testing.T) {

	var err error
	var r *core.RexRay
	if r, err = getRexRay(); err != nil {
		panic(err)
	}

	allDriverNames := []string{
		strings.ToLower(mock.MockOSDriverName),
		strings.ToLower(mock.MockVolDriverName),
		strings.ToLower(mock.MockStorDriverName),
		strings.ToLower(mock.BadMockOSDriverName),
		strings.ToLower(mock.BadMockVolDriverName),
		strings.ToLower(mock.BadMockStorDriverName),
		"linux",
		"docker",
		"ec2",
		"openstack",
		"rackspace",
		"scaleio",
		"xtremio",
	}

	var regDriverNames []string
	for dn := range r.DriverNames() {
		regDriverNames = append(regDriverNames, strings.ToLower(dn))
	}

	for _, n := range allDriverNames {
		if !gotil.StringInSlice(n, regDriverNames) {
			t.Fail()
		}
	}

	for _, n := range regDriverNames {
		if !gotil.StringInSlice(n, allDriverNames) {
			t.Fail()
		}
	}
}

func assertDriverNames(t *testing.T, r *core.RexRay) {
	od := <-r.OS.Drivers()
	if od.Name() != mock.MockOSDriverName {
		t.Fatalf("expected %s but was %s", mock.MockOSDriverName, od.Name())
	}

	vd := <-r.Volume.Drivers()
	if vd.Name() != mock.MockVolDriverName {
		t.Fatalf("expected %s but was %s", mock.MockVolDriverName, vd.Name())
	}

	sd := <-r.Storage.Drivers()
	if sd.Name() != mock.MockStorDriverName {
		t.Fatalf("expected %s but was %s", mock.MockStorDriverName, sd.Name())
	}
}

var yamlConfig1 = []byte(`
rexray:
  logLevel: error
  osDrivers:
  - mockOSDriver
  volumeDrivers:
  - mockVolumeDriver
  storageDrivers:
  - mockStorageDriver
`)
