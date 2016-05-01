package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

// transactionHandler is a global HTTP filter for grokking the transaction info
// from the headers
type transactionHandler struct {
	handler types.APIFunc
}

// NewTransactionHandler returns a new global HTTP filter for grokking the
// transaction info from the headers
func NewTransactionHandler() types.Middleware {
	return &transactionHandler{}
}

func (h *transactionHandler) Name() string {
	return "transaction-handler"
}

func (h *transactionHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&transactionHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *transactionHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	txIDHeaders := utils.GetHeader(req.Header, types.TransactionIDHeader)
	ctx.WithField(
		types.TransactionIDHeader, txIDHeaders).Debug("http header")

	var txID string
	if len(txIDHeaders) > 0 {
		txID = txIDHeaders[0]
	} else {
		txIDUUID, _ := utils.NewUUID()
		txID = txIDUUID.String()
	}
	ctx = ctx.WithTransactionID(txID)
	ctx = ctx.WithContextSID(types.ContextTransactionID, txID)

	txCRHeaders := utils.GetHeader(req.Header, types.TransactionCreatedHeader)
	ctx.WithField(
		types.TransactionCreatedHeader, txCRHeaders).Debug("http header")

	txCR := time.Now().UTC()
	if len(txCRHeaders) > 0 {
		epoch, err := strconv.ParseInt(txCRHeaders[0], 10, 64)
		if err != nil {
			return err
		}
		txCR = time.Unix(epoch, 0)
	}
	ctx = ctx.WithTransactionCreated(txCR)
	ctx = ctx.WithContextSID(
		types.ContextTransactionCreated,
		fmt.Sprintf("%d", txCR.Unix()))

	return h.handler(ctx, w, req, store)
}
