package semaphore

import (
	"os"

	"github.com/akutz/goof"
)

func open(
	name string, excl bool, perm os.FileMode, val uint) (Semaphore, error) {
	return nil, goof.New("unsupported")
}
