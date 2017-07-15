// +build !libstorage_storage_driver

package executors

import (
	// load the storage executors
	_ "github.com/codedellemc/libstorage/drivers/storage/azureud/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/cinder/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/dobs/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/ebs/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/efs/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/fittedcloud/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/gcepd/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/isilon/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/rbd/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/s3fs/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/scaleio/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/vbox/executor"
	_ "github.com/codedellemc/libstorage/drivers/storage/vfs/executor"
)
