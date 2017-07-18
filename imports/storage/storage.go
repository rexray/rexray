// +build !azureud
// +build !cinder
// +build !dobs
// +build !ebs
// +build !efs
// +build !fittedcloud
// +build !gcepd
// +build !isilon
// +build !rbd
// +build !s3fs
// +build !scaleio
// +build !vbox
// +build !vfs

package storage

import (
	// import the storage drivers
	_ "github.com/codedellemc/libstorage/drivers/storage/azureud/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/cinder/storage"
	_ "github.com/codedellemc/libstorage/drivers/storage/dobs/storage"
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
