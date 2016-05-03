package libstorage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func withTX(ctx types.Context) types.Context {
	txIDUUID, _ := utils.NewUUID()
	txID := txIDUUID.String()

	ctx = ctx.WithTransactionID(txID)
	ctx = ctx.WithContextSID(types.ContextTransactionID, txID)

	txCR := time.Now().UTC()
	ctx = ctx.WithTransactionCreated(txCR)
	ctx = ctx.WithContextSID(
		types.ContextTransactionCreated,
		fmt.Sprintf("%d", txCR.Unix()))

	return ctx
}

func withContext(ctx types.Context, parent types.Context) types.Context {
	if ctx == nil || ctx == parent {
		return withTX(parent)
	}
	return ctx.Join(parent)
}

func (c *client) withContext(ctx types.Context) types.Context {
	return withContext(ctx, c.ctx)
}

func (c *client) updateInstanceIDHeaders(
	driverName string,
	iid *types.InstanceID) error {

	headerKey := types.InstanceIDHeader
	headerValue := fmt.Sprintf("%s,%v", iid.ID, iid.Formatted)

	if len(iid.Metadata) > 0 {
		buf, err := json.Marshal(iid)
		if err != nil {
			return err
		}
		headerKey = types.InstanceID64Header
		headerValue = base64.StdEncoding.EncodeToString(buf)
	}

	c.AddHeaderForDriver(driverName, headerKey, headerValue)
	return nil
}

func (c *client) updateLocalDevicesHeaders(
	driverName string,
	ldm map[string]string) error {

	buf := &bytes.Buffer{}
	for k, v := range ldm {
		if _, err := fmt.Fprintf(buf, "%s=%s, ", k, v); err != nil {
			return nil
		}
	}

	if buf.Len() > 2 {
		buf.Truncate(buf.Len() - 2)
	}

	c.AddHeaderForDriver(driverName, types.LocalDevicesHeader, buf.String())

	return nil
}
