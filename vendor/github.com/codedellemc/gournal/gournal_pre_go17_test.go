// +build !go1.7

package gournal

import (
	"bytes"

	"golang.org/x/net/context"
)

func newTestContext() (*bytes.Buffer, context.Context) {
	w := &bytes.Buffer{}
	a := NewAppenderWithOptions(w)
	return w, context.WithValue(context.Background(), AppenderKey(), a)
}
