package test

import (
	"testing"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
)

type mockOSDriver struct {
	name string
}

func newOSDriver() core.Driver {
	var d core.OSDriver = &mockOSDriver{mockOSDriverName}
	return d
}

func (m *mockOSDriver) Init(r *core.RexRay) error {
	return nil
}

func (m *mockOSDriver) Name() string {
	return m.name
}

func TestOSDriverName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.OS.Drivers()
	if d.Name() != mockOSDriverName {
		t.Fatalf("driver name != %s, == %s", mockOSDriverName, d.Name())
	}
}

func TestOSDriverManagerName(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if r.OS.Name() != mockOSDriverName {
		t.Fatalf("driver name != %s, == %s", mockOSDriverName, r.OS.Name())
	}
}

func TestOSDriverManagerNameNoDrivers(t *testing.T) {
	r, _ := getRexRayNoDrivers()
	if r.OS.Name() != "" {
		t.Fatal("name not empty")
	}
}

func (m *mockOSDriver) GetMounts(string, string) (core.MountInfoArray, error) {
	return nil, nil
}

func TestOSDriverGetMounts(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.OS.Drivers()
	if v, err := d.GetMounts("", ""); v != nil || err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerGetMounts(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if v, err := r.OS.GetMounts("", ""); v != nil || err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerGetMountsNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.OS.GetMounts("", ""); err != errors.ErrNoOSDetected {
		t.Fatal(err)
	}
}

func (m *mockOSDriver) Mounted(string) (bool, error) {
	return false, nil
}

func TestOSDriverMounted(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.OS.Drivers()
	if v, err := d.Mounted(""); v || err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerMounted(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if v, err := r.OS.Mounted(""); v || err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerMountedNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.OS.Mounted(""); err != errors.ErrNoOSDetected {
		t.Fatal(err)
	}
}

func (m *mockOSDriver) Unmount(string) error {
	return nil
}

func TestOSDriverUnmount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.OS.Drivers()
	if err := d.Unmount(""); err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerUnmount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.OS.Unmount(""); err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerUnmountNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.OS.Unmount(""); err != errors.ErrNoOSDetected {
		t.Fatal(err)
	}
}

func (m *mockOSDriver) Mount(string, string, string, string) error {
	return nil
}

func TestOSDriverMount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.OS.Drivers()
	if err := d.Mount("", "", "", ""); err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerMount(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.OS.Mount("", "", "", ""); err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerMountNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.OS.Mount("", "", "", ""); err != errors.ErrNoOSDetected {
		t.Fatal(err)
	}
}

func (m *mockOSDriver) Format(string, string, bool) error {
	return nil
}

func TestOSDriverFormat(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	d := <-r.OS.Drivers()
	if err := d.Format("", "", false); err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerFormat(t *testing.T) {
	r, err := getRexRay()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.OS.Format("", "", false); err != nil {
		t.Fail()
	}
}

func TestOSDriverManagerFormatNoDrivers(t *testing.T) {
	r, err := getRexRayNoDrivers()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.OS.Format("", "", false); err != errors.ErrNoOSDetected {
		t.Fatal(err)
	}
}
