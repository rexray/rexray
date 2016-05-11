package libstorage

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils/paths"
	"github.com/emccode/libstorage/cli/lsx"
)

func (c *client) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	ctx = context.RequireTX(ctx.Join(c.ctx))

	serviceName, ok := context.ServiceName(ctx)
	if !ok {
		return nil, goof.New("missing service name")
	}

	if iid := c.instanceIDCache.GetInstanceID(serviceName); iid != nil {
		return iid, nil
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := strings.ToLower(si.Driver.Name)

	out, err := c.runExecutor(ctx, driverName, lsx.InstanceID)
	if err != nil {
		return nil, err
	}

	iid := &types.InstanceID{}
	if err := iid.UnmarshalText(out); err != nil {
		return nil, err
	}

	ctx = ctx.WithValue(context.InstanceIDKey, iid)
	instance, err := c.InstanceInspect(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	iid = instance.InstanceID
	c.instanceIDCache.Set(serviceName, iid)

	ctx.Debug("xli instanceID success")
	return iid, nil
}

func (c *client) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

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

	out, err := c.runExecutor(ctx, driverName, lsx.NextDevice)
	if err != nil {
		return "", err
	}

	ctx.Debug("xli nextdevice success")
	return gotil.Trim(string(out)), nil
}

func (c *client) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

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
		ctx, driverName, lsx.LocalDevices, opts.ScanType.String())
	if err != nil {
		return nil, err
	}

	ld := &types.LocalDevices{}
	if err := ld.UnmarshalText(out); err != nil {
		return nil, err
	}

	ctx.Debug("xli localdevices success")
	return ld, nil
}

func (c *client) WaitForDevice(
	ctx types.Context,
	opts *types.WaitForDeviceOpts) (bool, *types.LocalDevices, error) {

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

	exitCode := 0
	out, err := c.runExecutor(
		ctx, driverName, lsx.WaitForDevice,
		opts.ScanType.String(), opts.Token, opts.Timeout.String())
	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode = exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if err != nil && exitCode > 0 {
		return false, nil, err
	}

	matched := exitCode == 0

	ld := &types.LocalDevices{}
	if err := ld.UnmarshalText(out); err != nil {
		return false, nil, err
	}

	ctx.Debug("xli waitfordevice success")
	return matched, ld, nil
}

func (c *client) runExecutor(
	ctx types.Context, args ...string) ([]byte, error) {

	ctx.Debug("waiting on executor lock")
	if err := lsxMutex.Wait(); err != nil {
		return nil, err
	}

	defer func() {
		ctx.Debug("signalling executor lock")
		if err := lsxMutex.Signal(); err != nil {
			panic(err)
		}
	}()

	cmd := exec.Command(paths.LSX.String(), args...)
	cmd.Env = os.Environ()

	configEnvVars := c.config.EnvVars()
	for _, cev := range configEnvVars {
		// ctx.WithField("value", cev).Debug("set executor env var")
		cmd.Env = append(cmd.Env, cev)
	}

	out, err := cmd.CombinedOutput()
	return out, err
}
