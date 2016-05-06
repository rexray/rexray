package libstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils/paths"
	"github.com/emccode/libstorage/cli/lsx"
)

func (c *client) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	ctx = c.withContext(ctx)
	serviceName := ctx.ServiceName()

	if iid := c.instanceIDCache.GetInstanceID(serviceName); iid != nil {
		return iid, nil
	}

	si, err := c.getServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	driverName := si.Driver.Name

	out, err := c.runExecutor(ctx, driverName, lsx.InstanceID)
	if err != nil {
		return nil, err
	}

	iid := &types.InstanceID{}
	if err := json.Unmarshal(out, iid); err != nil {
		return nil, err
	}

	if err := c.updateInstanceIDHeaders(driverName, iid); err != nil {
		return nil, err
	}

	instance, err := c.InstanceInspect(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	iid = instance.InstanceID

	// add the formatted instance ID back to the headers, replacing the
	// unformatted one
	if err := c.updateInstanceIDHeaders(driverName, iid); err != nil {
		return nil, err
	}

	c.instanceIDCache.Set(serviceName, iid)

	ctx.Debug("xli instanceID success")
	return iid, nil
}

func (c *client) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	si, err := c.getServiceInfo(ctx.ServiceName())
	if err != nil {
		return "", err
	}
	driverName := si.Driver.Name

	out, err := c.runExecutor(
		c.withContext(ctx), driverName, lsx.NextDevice)
	if err != nil {
		return "", err
	}

	ctx.Debug("xli nextdevice success")
	return gotil.Trim(string(out)), nil
}

func (c *client) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (map[string]string, error) {

	ctx = c.withContext(ctx)
	serviceName := ctx.ServiceName()

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

	ldm := map[string]string{}
	if err := json.Unmarshal(out, &ldm); err != nil {
		return nil, err
	}

	if err := c.updateLocalDevicesHeaders(driverName, ldm); err != nil {
		return nil, err
	}

	ctx.Debug("xli localdevices success")
	return ldm, nil
}

func (c *client) WaitForDevice(
	ctx types.Context,
	opts *types.WaitForDeviceOpts) (bool, map[string]string, error) {

	ctx = c.withContext(ctx)

	si, err := c.getServiceInfo(ctx.ServiceName())
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

	ldm := map[string]string{}
	if err := json.Unmarshal(out, &ldm); err != nil {
		return matched, nil, err
	}

	buf := &bytes.Buffer{}
	for k, v := range ldm {
		fmt.Fprintf(buf, "%s=%s, ", k, v)
	}

	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 2)
	}

	c.AddHeaderForDriver(driverName, types.LocalDevicesHeader, buf.String())

	ctx.Debug("xli waitfordevice success")
	return matched, ldm, nil
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
