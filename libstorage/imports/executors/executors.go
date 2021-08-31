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

package executors

import (
	// load the storage executors
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/azureud/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/cinder/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/dobs/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/efs/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/fittedcloud/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/gcepd/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/isilon/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/rbd/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/s3fs/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/scaleio/executor"
	_ "github.com/AVENTER-UG/rexray/libstorage/drivers/storage/vfs/executor"
)
