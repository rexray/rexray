package client

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
)

func (c *lsc) updateExecutorInfo() error {
	ctx := c.getTXCTX()
	ctx.Log().Debug("getting executor information")
	lsxInfo, err := c.Client.Executors(ctx)
	if err != nil {
		return err
	}
	c.lsxInfo = lsxInfo
	return nil
}

func (c *lsc) updateExecutor() error {
	ctx := c.getTXCTX()
	ctx.Log().Debug("updating executor")

	lsxi, ok := c.lsxInfo[types.LSX]
	if !ok {
		return goof.WithField("lsx", types.LSX, "unknown executor")
	}

	ctx.Log().Debug("waiting on executor lock")
	if err := lsxMutex.Wait(); err != nil {
		return err
	}
	defer func() {
		ctx.Log().Debug("signalling executor lock")
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

func (c *lsc) getExecutorChecksum(ctx context.Context) (string, error) {
	ctx.Log().Debug("getting executor checksum")

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

func (c *lsc) downloadExecutor(ctx context.Context) error {

	ctx.Log().Debug("downloading executor")

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

func (c *lsc) runExecutor(
	ctx context.Context, driverName, cmdName string) ([]byte, error) {
	ctx.Log().Debug("waiting on executor lock")
	if err := lsxMutex.Wait(); err != nil {
		return nil, err
	}
	defer func() {
		ctx.Log().Debug("signalling executor lock")
		if err := lsxMutex.Signal(); err != nil {
			panic(err)
		}
	}()
	cmd := exec.Command(c.lsxBinPath, driverName, cmdName)
	cmd.Env = os.Environ()
	for _, cev := range c.config.EnvVars() {
		ctx.Log().WithField("value", cev).Debug("set executor env var")
		cmd.Env = append(cmd.Env, cev)
	}
	return cmd.CombinedOutput()
}
