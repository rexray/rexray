// +build gofig

package registry

import (
	"github.com/akutz/gofig"
)

func init() {
	NewConfig = gofig.New
}
