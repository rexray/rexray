package libstorage

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
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

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdSupported)
	if err != nil {
		if err == types.ErrNotImplemented {
			ctx.WithField("serviceDriver", driverName).Warn(
				"supported cmd not implemented")
			c.supportedCache.Set(driverName, types.LSXOpAllNoMount)
			ctx.WithField("supported", true).Debug("cached supported flag")
			return types.LSXOpAllNoMount, nil
		}
		return 0, err
	}

	if len(out) == 0 {
		return 0, nil
	}

	out = bytes.TrimSpace(out)
	i, err := strconv.Atoi(string(out))
	if err != nil {
		return 0, err
	}
	lsxSOp := types.LSXSupportedOp(i)

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

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdInstanceID)
	if err != nil {
		return nil, err
	}

	iid := &types.InstanceID{}
	if err := iid.UnmarshalText(out); err != nil {
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

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdNextDevice)
	if err != nil {
		return "", err
	}

	ctx.Debug("xli nextdevice success")
	return gotil.Trim(string(out)), nil
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

	out, err := c.runExecutor(
		ctx, driverName, types.LSXCmdLocalDevices, opts.ScanType.String())
	if err != nil {
		return nil, err
	}

	ld, err := unmarshalLocalDevices(ctx, out)
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

	out, err := c.runExecutor(
		ctx, driverName, types.LSXCmdWaitForDevice,
		opts.ScanType.String(), opts.Token, opts.Timeout.String())

	if err != types.ErrTimedOut {
		return false, nil, err
	}

	matched := err == nil

	ld, err := unmarshalLocalDevices(ctx, out)
	if err != nil {
		return false, nil, err
	}

	ctx.Debug("xli waitfordevice success")
	return matched, ld, nil
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

	args := []string{
		driverName,
		types.LSXCmdMount,
		deviceName,
		mountPoint,
	}
	if len(opts.MountLabel) > 0 {
		args = append(args, "-l", opts.MountLabel)
	}
	if len(opts.MountOptions) > 0 {
		args = append(args, "-o", opts.MountOptions)
	}

	if _, err = c.runExecutor(ctx, args...); err != nil {
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

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdMounts)
	if err != nil {
		return nil, err
	}

	var mounts []*types.MountInfo
	if err := json.Unmarshal(out, &mounts); err != nil {
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

	if _, err = c.runExecutor(
		ctx,
		driverName,
		types.LSXCmdUmount,
		mountPoint); err != nil {
		return err
	}

	ctx.Debug("xli umount success")
	return nil
}

func unmarshalLocalDevices(
	ctx types.Context, out []byte) (*types.LocalDevices, error) {

	ld := &types.LocalDevices{}
	if err := ld.UnmarshalText(out); err != nil {
		return nil, err
	}

	// remove any local devices that has no mapped volume information
	for k, v := range ld.DeviceMap {
		if len(v) == 0 {
			ctx.WithField("deviceID", k).Warn(
				"removing local device w/ invalid volume id")
			delete(ld.DeviceMap, k)
		}
	}

	return ld, nil
}

func (c *client) runExecutor(
	ctx types.Context, args ...string) ([]byte, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "runExecutor")
	}

	ctx.Debug("waiting on executor lock")
	if err := c.lsxMutexWait(); err != nil {
		return nil, err
	}

	defer func() {
		ctx.Debug("signalling executor lock")
		if err := c.lsxMutexSignal(); err != nil {
			panic(err)
		}
	}()

	lsxBin := types.LSX.String()
	cmd := exec.Command(lsxBin, args...)
	cmd.Env = os.Environ()

	ctx.WithFields(log.Fields{
		"cmd":  lsxBin,
		"args": args,
	}).Debug("invoking executor cli")

	configEnvVars := c.config.EnvVars()
	for _, cev := range configEnvVars {
		// ctx.WithField("value", cev).Debug("set executor env var")
		cmd.Env = append(cmd.Env, cev)
	}

	out, err := cmd.Output()

	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.Sys().(syscall.WaitStatus).ExitStatus()
		switch exitCode {
		case types.LSXExitCodeNotImplemented:
			return nil, types.ErrNotImplemented
		case types.LSXExitCodeTimedOut:
			return nil, types.ErrTimedOut
		}
		stderr := string(exitError.Stderr)
		ctx.WithFields(log.Fields{
			"cmd":    lsxBin,
			"args":   args,
			"stderr": stderr,
		}).Error("error from executor cli")
		return nil, goof.WithFieldsE(
			map[string]interface{}{
				"lsx":    lsxBin,
				"args":   args,
				"stderr": stderr,
			},
			"error executing xcli",
			err)
	}

	return out, err
}

func (c *client) lsxMutexWait() error {

	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "lsxMutexWait")
	}

	for {
		f, err := os.OpenFile(lsxMutex, os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		return f.Close()
	}
}

func (c *client) lsxMutexSignal() error {
	if c.isController() {
		return utils.NewUnsupportedForClientTypeError(
			c.clientType, "lsxMutexSignal")
	}
	return os.RemoveAll(lsxMutex)
}
