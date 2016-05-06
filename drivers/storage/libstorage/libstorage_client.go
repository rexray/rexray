package libstorage

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"io"
	"os"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/api/utils/paths"
)

type client struct {
	types.APIClient
	ctx             types.Context
	config          gofig.Config
	serviceCache    *lss
	lsxCache        *lss
	instanceIDCache types.Store
}

func (c *client) dial(ctx types.Context) error {

	svcInfos, err := c.Services(ctx)
	if err != nil {
		return err
	}

	store := utils.NewStore()
	c.ctx = c.ctx.WithContextSID(types.ContextServerName, c.ServerName())

	if !c.config.GetBool(types.ConfigExecutorNoDownload) {

		ctx.Info("initializing executors cache")
		if _, err := c.Executors(ctx); err != nil {
			return err
		}

		if err := c.updateExecutor(ctx); err != nil {
			return err
		}
	}

	for service, _ := range svcInfos {
		ctx := c.ctx.WithServiceName(service)

		ctx.Info("initializing instance ID cache")
		if _, err := c.InstanceID(ctx, store); err != nil {
			return err
		}
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

func (c *client) getServiceInfo(service string) (*types.ServiceInfo, error) {
	if si := c.serviceCache.GetServiceInfo(service); si != nil {
		return si, nil
	}
	return nil, goof.WithField("name", service, "unknown service")
}

func (c *client) updateExecutor(ctx types.Context) error {

	ctx.Debug("updating executor")

	lsxi := c.lsxCache.GetExecutorInfo(types.LSX)
	if lsxi == nil {
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

	if !paths.LSX.Exists() {
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

	f, err := os.Open(paths.LSX.String())
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
		paths.LSX.String(),
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0755)
	if err != nil {
		return err
	}

	defer f.Close()

	rdr, err := c.APIClient.ExecutorGet(ctx, types.LSX)
	if _, err := io.Copy(f, rdr); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
