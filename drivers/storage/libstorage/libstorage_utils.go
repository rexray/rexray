package libstorage

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (c *client) runExecutor(
	ctx types.Context, driverName, cmdName string) ([]byte, error) {

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
	cmd := exec.Command(c.lsxBinPath, driverName, cmdName)
	cmd.Env = os.Environ()
	for _, cev := range c.config.EnvVars() {
		ctx.WithField("value", cev).Debug("set executor env var")
		cmd.Env = append(cmd.Env, cev)
	}
	return cmd.CombinedOutput()
}

func withTX(ctx types.Context) types.Context {
	txIDUUID, _ := utils.NewUUID()
	txID := txIDUUID.String()

	ctx = ctx.WithTransactionID(txID)
	ctx = ctx.WithContextSID(types.CtxKeyTransactionID, txID)

	txCR := time.Now().UTC()
	ctx = ctx.WithTransactionCreated(txCR)
	ctx = ctx.WithContextSID(
		types.CtxKeyTransactionCreated,
		fmt.Sprintf("%d", txCR.Unix()))

	return ctx
}

func withContext(ctx types.Context, parent types.Context) types.Context {
	if ctx == nil {
		return withTX(parent)
	}
	return ctx.Join(parent)
}

func (c *client) withContext(ctx types.Context) types.Context {
	return withContext(ctx, c.ctx)
}
