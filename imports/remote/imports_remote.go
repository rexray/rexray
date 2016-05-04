package remote

import (
	// import to load
	_ "github.com/emccode/libstorage/drivers/storage/mock"
	_ "github.com/emccode/libstorage/drivers/storage/scaleio/storage"
	_ "github.com/emccode/libstorage/drivers/storage/vfs/storage"
	_ "github.com/emccode/libstorage/drivers/storage/virtualbox"
)
