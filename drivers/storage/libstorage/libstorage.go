package libstorage

import (
	log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils/semaphore"
)

const (
	// Name is the name of the driver.
	Name = "libstorage"
)

var (
	lsxMutex semaphore.Semaphore
)

func init() {
	registry.RegisterStorageDriver(Name, newDriver)

	var err error
	for {
		lsxMutex, err = semaphore.Open(types.LSX, false, 0644, 1)
		if err != nil {
			log.WithError(err).Warn(err)
		} else {
			break
		}
	}

	registerConfig()
}

// Close releases system resources.
func Close() error {
	return lsxMutex.Close()
}
