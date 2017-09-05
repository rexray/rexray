package gocsi

import (
	"golang.org/x/net/context"
)

type contextKey uint64

var (
	requestIDKey      interface{} = contextKey(0)
	fullMethodNameKey interface{} = contextKey(1)
)

// GetRequestID gets the gRPC request ID from the provided context.
func GetRequestID(ctx context.Context) (uint64, bool) {
	v, ok := ctx.Value(requestIDKey).(uint64)
	return v, ok
}
