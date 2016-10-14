package libstorage

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (c *client) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	if c.isController() {
		return false, utils.NewUnsupportedForClientTypeError(
			c.clientType, "Supported")
	}

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return false, goof.New("missing service name")
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return false, err
	}
	driverName := strings.ToLower(si.Driver.Name)

	// check to see if the driver's executor is supported on this host
	if ok := c.supportedCache.IsSet(driverName); ok {
		return c.supportedCache.GetBool(driverName), nil
	}

	out, err := c.runExecutor(ctx, driverName, types.LSXCmdSupported)
	if err != nil {
		if err == types.ErrNotImplemented {
			ctx.WithField("serviceDriver", driverName).Warn(
				"supported cmd not implemented")
			c.supportedCache.Set(driverName, true)
			ctx.WithField("supported", true).Debug("cached supported flag")
			return true, nil
		}
		return false, err
	}

	if len(out) == 0 {
		return false, nil
	}

	out = bytes.TrimSpace(out)
	b, err := strconv.ParseBool(string(out))
	if err != nil {
		return false, err
	}

	c.supportedCache.Set(driverName, b)
	ctx.WithField("supported", b).Debug("cached supported flag")
	return b, nil
}

func (c *client) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	if c.isController() {
		return nil, utils.NewUnsupportedForClientTypeError(
			c.clientType, "InstanceID")
	}

	if supported, _ := c.Supported(ctx, opts); !supported {
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
	if iid := c.instanceIDCache.GetInstanceID(driverName); iid != nil {
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

	ctx = ctx.WithValue(context.InstanceIDKey, iid)

	if iid.HasMetadata() {
		ctx.Debug("sending instanceID in API.InstanceInspect call")
		instance, err := c.InstanceInspect(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		ctx.Debug("received instanceID from API.InstanceInspect call")
		iid.ID = instance.InstanceID.ID
		iid.Fields = instance.InstanceID.Fields
		iid.DeleteMetadata()
	}

	c.instanceIDCache.Set(driverName, iid)
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

	if supported, _ := c.Supported(ctx, opts); !supported {
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

	if supported, _ := c.Supported(ctx, opts.Opts); !supported {
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

	if supported, _ := c.Supported(ctx, opts.Opts); !supported {
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
		return nil, goof.WithFieldsE(
			map[string]interface{}{
				"lsx":  lsxBin,
				"args": args,
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
