package test

import (
	"testing"

	"github.com/emccode/rexray/core"
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
	if v, err := r.OS.GetMounts("", ""); v != nil || err != nil {
		t.Fail()
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
	if v, err := r.OS.Mounted(""); v || err != nil {
		t.Fail()
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
	if err := r.OS.Unmount(""); err != nil {
		t.Fail()
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
	if err := r.OS.Mount("", "", "", ""); err != nil {
		t.Fail()
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
	if err := r.OS.Format("", "", false); err != nil {
		t.Fail()
	}
}
