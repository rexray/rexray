// +build mock
// +build driver executor

package mock

import (
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

var (
	nextDeviceVals = []string{"/dev/mock0", "/dev/mock1", "/dev/mock2"}
)

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx context.Context,
	opts types.Store) (*types.InstanceID, error) {
	return d.instanceID, nil
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx context.Context,
	opts types.Store) (string, error) {
	return "", nil
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx context.Context,
	opts types.Store) (map[string]string, error) {
	return nil, nil
}
