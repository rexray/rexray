// +build !mock

package libstorage

import (
	"testing"

	"github.com/akutz/gofig"
)

func getConfig(host string, tls bool, t *testing.T) gofig.Config {
	return gofig.New()
}
