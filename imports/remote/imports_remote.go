// +build !libstorage_storage_driver

package remote

import (
	// import to load
	_ "github.com/codedellemc/libstorage/drivers/storage/digitalocean/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/ebs/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/efs/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/fittedcloud/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/gcepd/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/isilon/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/rbd/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/s3fs/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/scaleio/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/vbox/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/vfs/storage"
)
