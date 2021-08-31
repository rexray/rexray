package libstorage

import (
	"strings"
	"time"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

func (c *client) Supported(
	ctx types.Context,
	opts types.Store) (types.LSXSupportedOp, error) {

	if c.isController() {
		return 0, utils.NewUnsupportedForClientTypeError(
			c.clientType, "Supported")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return 0, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return 0, err
	}
	driverName := strings.ToLower(si.Driver.Name)

	// check to see if the driver's executor is supported on this host
	if ok := c.supportedCache.IsSet(driverName); ok {
		return c.supportedCache.GetLSXSupported(driverName), nil
	}

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return 0, err
	}

	lsxSOp := types.LSXOpAllNoMount

	if dws, ok := d.(types.StorageExecutorWithSupported); ok {
		if ok, err := dws.Supported(ctx, opts); err != nil {
			return 0, err
		} else if ok {
			if _, ok := dws.(types.StorageExecutorWithMount); ok {
				lsxSOp = lsxSOp | types.LSXSOpMount
			}
			if _, ok := dws.(types.StorageExecutorWithUnmount); ok {
				lsxSOp = lsxSOp | types.LSXSOpUmount
			}
			if _, ok := dws.(types.StorageExecutorWithMounts); ok {
				lsxSOp = lsxSOp | types.LSXSOpMounts
			}
		}
	}

	c.supportedCache.Set(driverName, lsxSOp)
	ctx.WithField("supported", lsxSOp).Debug("cached supported flag")
	return lsxSOp, nil
}

func (c *client) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "InstanceID")
	}

	if lsxSO, _ := c.Supported(ctx, opts); !lsxSO.InstanceID() {
		return nil, errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return nil, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := strings.ToLower(si.Driver.Name)

	// check to see if the driver's instance ID is cached
	if iid := c.instanceIDCache.GetInstanceID(serviceName); iid != nil {
		ctx.WithField("service", serviceName).Debug("found cached instance ID")
		return iid, nil
	}

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return nil, err
	}

	iid, err := d.InstanceID(ctx, opts)
	if err != nil {
		return nil, err
	}

	iid.Service = ""
	ctx = ctx.WithValue(context.InstanceIDKey, iid)

	if iid.HasMetadata() {
		ctx.Debug("sending instanceID in API.InstanceInspect call")
		instance, err := c.InstanceInspect(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		ctx.Debug("received instanceID from API.InstanceInspect call")
		iid.ID = instance.InstanceID.ID
		iid.Driver = driverName

		// Set the instance ID's Service field to be the service name. This is
		// important as when the driver is marshalled to a string for inclusion
		// in HTTP headers, the Service field will be included in the output.
		//
		// Instance ID headers without the Service field that are the same
		// otherwise can be collapsed into a single header, reducing the
		// amount of data that needs to be transmitted.
		//
		// Since this instance ID required inspection to transform it into its
		// final format, it's likely that the instance ID is dependent upon
		// the service and not the same for another service that uses the same
		// driver type.
		iid.Service = serviceName

		iid.Fields = instance.InstanceID.Fields
		iid.DeleteMetadata()
	}

	c.instanceIDCache.Set(serviceName, iid)
	ctx.Debug("cached instanceID")

	ctx.Debug("xli instanceID success")
	return iid, nil
}

func (c *client) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	if c.isController() {
		return "", utils.NewUnsupportedForClientTypeError(
			c.clientType, "NextDevice")
	}

	if lsxSO, _ := c.Supported(ctx, opts); !lsxSO.NextDevice() {
		return "", errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return "", goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	driverName := si.Driver.Name

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return "", err
	}

	nextDevice, err := d.NextDevice(ctx, opts)
	if err != nil {
		if err.Error() == types.ErrNotImplemented.Error() {
			return "", nil
		}
		return "", err
	}

	ctx.WithField("nextDevice", nextDevice).Debug("xli nextdevice success")
	return gotil.Trim(string(nextDevice)), nil
}

func (c *client) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "LocalDevices")
	}

	if lsxSO, _ := c.Supported(ctx, opts.Opts); !lsxSO.LocalDevices() {
		return nil, errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return nil, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := si.Driver.Name

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return nil, err
	}

	ld, err := c.getLocalDevices(ctx, d, opts)
	if err != nil {
		return nil, err
	}

	ctx.Debug("xli localdevices success")
	return ld, nil
}

func (c *client) WaitForDevice(
	ctx types.Context,
	opts *types.WaitForDeviceOpts) (bool, *types.LocalDevices, error) {

	if c.isController() {
		return false, nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "WaitForDevice")
	}

	if lsxSO, _ := c.Supported(ctx, opts.Opts); !lsxSO.WaitForDevice() {
		return false, nil, errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return false, nil, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return false, nil, err
	}
	driverName := si.Driver.Name

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return false, nil, err
	}

	found, ld, err := func() (bool, *types.LocalDevices, error) {
		timeoutC := time.After(opts.Timeout)
		tick := time.Tick(500 * time.Millisecond)
		for {
			select {
			case <-timeoutC:
				return false, nil, types.ErrTimedOut
			case <-tick:
				ld, err := c.getLocalDevices(ctx, d, &opts.LocalDevicesOpts)
				if err != nil {
					return false, nil, err
				}
				for k := range ld.DeviceMap {
					if strings.ToLower(k) == opts.Token {
						return true, ld, nil
					}
				}
			}
		}
	}()
	if err != nil {
		return false, nil, err
	}

	ctx.Debug("xli waitfordevice success")
	return found, ld, nil
}

// Mount mounts a device to a specified path.
func (c *client) Mount(
	ctx types.Context,
	deviceName, mountPoint string,
	opts *types.DeviceMountOpts) error {

	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "Mount")
	}

	if lsxSO, _ := c.Supported(ctx, opts.Opts); !lsxSO.Mount() {
		return errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return err
	}
	driverName := si.Driver.Name

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return err
	}

	dd, ok := d.(types.StorageExecutorWithMount)
	if !ok {
		return types.ErrNotImplemented
	}

	if err := dd.Mount(ctx, deviceName, mountPoint, opts); err != nil {
		return err
	}

	ctx.Debug("xli mount success")
	return nil
}

func (c *client) Mounts(
	ctx types.Context,
	opts types.Store) ([]*types.MountInfo, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "Mounts")
	}

	if lsxSO, _ := c.Supported(ctx, opts); !lsxSO.Mounts() {
		return nil, errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return nil, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := si.Driver.Name

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return nil, nil
	}

	dd, ok := d.(types.StorageExecutorWithMounts)
	if !ok {
		return nil, types.ErrNotImplemented
	}

	mounts, err := dd.Mounts(ctx, opts)
	if err != nil {
		return nil, err
	}

	ctx.Debug("xli mounts success")
	return mounts, nil
}

// Unmount unmounts the underlying device from the specified path.
func (c *client) Unmount(
	ctx types.Context,
	mountPoint string,
	opts types.Store) error {

	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "Unmount")
	}

	if lsxSO, _ := c.Supported(ctx, opts); !lsxSO.Umount() {
		return errExecutorNotSupported
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return err
	}
	driverName := si.Driver.Name

	// create the executor
	d, err := c.getExecutor(ctx, driverName)
	if err != nil {
		return err
	}

	dd, ok := d.(types.StorageExecutorWithUnmount)
	if !ok {
		return types.ErrNotImplemented
	}

	if err := dd.Unmount(ctx, mountPoint, opts); err != nil {
		return err
	}

	ctx.Debug("xli unmount success")
	return nil
}

func (c *client) getExecutor(
	ctx types.Context,
	driverName string) (types.StorageExecutor, error) {

	// create the executor
	d, err := registry.NewStorageExecutor(driverName)
	if err != nil {
		ctx.WithField("driver", driverName).WithError(err).Error(
			"error creating executor")
		return nil, err
	}
	if err := d.Init(ctx, c.config); err != nil {
		ctx.WithField("driver", driverName).WithError(err).Error(
			"error initializing executor")
		return nil, err
	}
	return d, nil
}

func (c *client) getLocalDevices(
	ctx types.Context,
	d types.StorageExecutor,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	ld, err := d.LocalDevices(ctx, opts)
	if err != nil {
		return nil, err
	}

	return ld, nil
}
