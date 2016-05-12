package handlers

import (
	"net/http"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/types"
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

	txHeader := req.Header.Get(types.TransactionHeader)
	ctx.WithField(types.TransactionHeader, txHeader).Debug("http header")

	if txHeader == "" {
		ctx = context.RequireTX(ctx)
	} else {
		tx := &types.Transaction{}
		if err := tx.UnmarshalText([]byte(txHeader)); err != nil {
			return err
		}
		ctx = ctx.WithValue(context.TransactionKey, tx)
	}

	return h.handler(ctx, w, req, store)
}
