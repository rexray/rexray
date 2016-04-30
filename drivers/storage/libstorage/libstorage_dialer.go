package libstorage

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/types"
)

func (c *client) dial(ctx types.Context) error {
	if err := c.updateServiceInfo(ctx); err != nil {
		return err
	}

	c.ctx = c.ctx.WithContextSID(types.CtxKeyServerName, c.ServerName())

	if !c.config.GetBool(lsxOffline) {
		if err := c.updateExecutorInfo(ctx); err != nil {
			return err
		}

		if err := c.updateExecutor(ctx); err != nil {
			return err
		}
	}

	if err := c.updateInstanceIDs(ctx); err != nil {
		return err
	}

	if err := c.updateLocalDevices(ctx); err != nil {
		return err
	}

	return nil
}

func getHost(proto, lAddr string, tlsConfig *tls.Config) string {
	if tlsConfig != nil && tlsConfig.ServerName != "" {
		return tlsConfig.ServerName
	} else if proto == "unix" {
		return "libstorage-server"
	} else {
		return lAddr
	}
}

func (c *client) updateServiceInfo(ctx types.Context) error {

	ctx.Debug("getting service information")
	svcInfo, err := c.Client.Services(ctx)
	if err != nil {
		return err
	}
	c.svcInfo = svcInfo
	return nil
}

type iidHeader struct {
	driverName string
	headerName string
	headerValu string
}

func (c *client) updateInstanceIDs(ctx types.Context) error {
	if !c.enableIIDHeader {
		return nil
	}

	ctx.Debug("getting instance IDs")
	cache := map[string]*iidHeader{}

	for service, si := range c.svcInfo {

		if _, ok := cache[si.Driver.Name]; ok {
			continue
		}

		iid, err := c.InstanceID(ctx, service)
		if err != nil {
			return err
		}

		var h *iidHeader

		if len(iid.Metadata) == 0 {
			h = &iidHeader{
				headerName: types.InstanceIDHeader,
				headerValu: iid.ID,
			}
		} else {
			jBuf, err := json.Marshal(iid)
			if err != nil {
				return err
			}
			h = &iidHeader{
				headerName: types.InstanceID64Header,
				headerValu: base64.StdEncoding.EncodeToString(jBuf),
			}
		}

		h.driverName = si.Driver.Name
		cache[h.driverName] = h
	}

	for _, h := range cache {
		c.Client.Headers.Add(
			h.headerName,
			fmt.Sprintf("%s=%s", h.driverName, h.headerValu))
	}

	return nil
}

type ldHeader struct {
	driverName string
	headerName string
	headerValu map[string]string
}

func (c *client) updateLocalDevices(ctx types.Context) error {
	if !c.enableLclDevHeader {
		return nil
	}

	ctx.Debug("getting local devices")

	cache := map[string]*ldHeader{}

	for service, si := range c.svcInfo {

		if _, ok := cache[si.Driver.Name]; ok {
			continue
		}

		ldm, err := c.LocalDevices(ctx, service)
		if err != nil {
			return err
		}

		h := &ldHeader{
			driverName: si.Driver.Name,
			headerName: types.LocalDevicesHeader,
			headerValu: ldm,
		}

		cache[h.driverName] = h
	}

	for _, h := range cache {
		buf := &bytes.Buffer{}

		fmt.Fprintf(buf, "%s=", h.driverName)
		for device, mountPoint := range h.headerValu {
			fmt.Fprintf(buf, "%s=%s, ", device, mountPoint)
		}

		if buf.Len() > (len(h.driverName) + 1) {
			buf.Truncate(buf.Len() - 2)
		}

		c.Client.Headers.Add(h.headerName, buf.String())
	}

	return nil
}

func (c *client) getServiceInfo(service string) (*types.ServiceInfo, error) {
	si, ok := c.svcInfo[strings.ToLower(service)]
	if !ok {
		return nil, goof.WithField("name", service, "unknown service")
	}
	return si, nil
}

func (c *client) updateExecutorInfo(ctx types.Context) error {

	ctx.Debug("getting executor information")
	lsxInfo, err := c.Executors(ctx)
	if err != nil {
		return err
	}
	c.lsxInfo = lsxInfo
	return nil
}

func (c *client) updateExecutor(ctx types.Context) error {

	ctx.Debug("updating executor")

	lsxi, ok := c.lsxInfo[types.LSX]
	if !ok {
		return goof.WithField("lsx", types.LSX, "unknown executor")
	}

	ctx.Debug("waiting on executor lock")
	if err := lsxMutex.Wait(); err != nil {
		return err
	}
	defer func() {
		ctx.Debug("signalling executor lock")
		if err := lsxMutex.Signal(); err != nil {
			panic(err)
		}
	}()

	if !gotil.FileExists(c.lsxBinPath) {
		return c.downloadExecutor(ctx)
	}

	checksum, err := c.getExecutorChecksum(ctx)
	if err != nil {
		return err
	}

	if lsxi.MD5Checksum != checksum {
		return c.downloadExecutor(ctx)
	}

	return nil
}

func (c *client) getExecutorChecksum(ctx types.Context) (string, error) {
	ctx.Debug("getting executor checksum")

	f, err := os.Open(c.lsxBinPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if _, err := h.Write(buf[:n]); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c *client) downloadExecutor(ctx types.Context) error {

	ctx.Debug("downloading executor")

	f, err := os.OpenFile(
		c.lsxBinPath,
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0755)
	if err != nil {
		return err
	}

	defer f.Close()

	rdr, err := c.Client.ExecutorGet(ctx, types.LSX)
	if _, err := io.Copy(f, rdr); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
