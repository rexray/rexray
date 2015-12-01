package mock

import (
	"github.com/akutz/goof"

	"github.com/emccode/rexray/core"
)

type mockOSDriver struct {
	name string
}

type badMockOSDriver struct {
	mockOSDriver
}

func newOSDriver() core.Driver {
	var d core.OSDriver = &mockOSDriver{MockOSDriverName}
	return d
}

func newBadOSDriver() core.Driver {
	var d core.OSDriver = &badMockOSDriver{mockOSDriver{BadMockOSDriverName}}
	return d
}

func (m *mockOSDriver) Init(r *core.RexRay) error {
	return nil
}

func (m *badMockOSDriver) Init(r *core.RexRay) error {
	return goof.New("init error")
}

func (m *mockOSDriver) Name() string {
	return m.name
}

func (m *mockOSDriver) GetMounts(string, string) (core.MountInfoArray, error) {
	return nil, nil
}

func (m *mockOSDriver) Mounted(string) (bool, error) {
	return false, nil
}

func (m *mockOSDriver) Unmount(string) error {
	return nil
}

func (m *mockOSDriver) Mount(string, string, string, string) error {
	return nil
}

func (m *mockOSDriver) Format(string, string, bool) error {
	return nil
}
