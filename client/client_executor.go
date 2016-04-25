package client

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/types"
)

func (c *lsc) updateExecutorInfo() error {
	c.ctx.Log().Debug("getting executor information")
	lsxInfo, err := c.Client.Executors(c.ctx)
	if err != nil {
		return err
	}
	c.lsxInfo = lsxInfo
	return nil
}

func (c *lsc) updateExecutor() error {
	lsxi, ok := c.lsxInfo[types.LSX]
	if !ok {
		return goof.WithField("lsx", types.LSX, "unknown executor")
	}

	c.ctx.Log().Debug("waiting on executor lock")
	if err := lsxMutex.Wait(); err != nil {
		return err
	}
	defer func() {
		c.ctx.Log().Debug("signalling executor lock")
		if err := lsxMutex.Signal(); err != nil {
			panic(err)
		}
	}()

	if !gotil.FileExists(c.lsxBinPath) {
		return c.downloadExecutor()
	}

	checksum, err := c.getExecutorChecksum()
	if err != nil {
		return err
	}

	if lsxi.MD5Checksum != checksum {
		return c.downloadExecutor()
	}

	return nil
}

func (c *lsc) getExecutorChecksum() (string, error) {
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

func (c *lsc) downloadExecutor() error {

	f, err := os.OpenFile(
		c.lsxBinPath,
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0755)
	if err != nil {
		return err
	}

	defer f.Close()

	rdr, err := c.Client.ExecutorGet(c.ctx, types.LSX)
	if _, err := io.Copy(f, rdr); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
