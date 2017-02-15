// +build !libstorage_storage_driver libstorage_storage_driver_gcepd

package gcepd

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "gcepd"

	// InstanceIDFieldProjectID is the key to retrieve the ProjectID value
	// from the InstanceID Field map.
	InstanceIDFieldProjectID = "projectID"

	// InstanceIDFieldZone is the key to retrieve the zone value from the
	// InstanceID Field map.
	InstanceIDFieldZone = "zone"

	// DiskTypeSSD indicates an SSD based disk should be created
	DiskTypeSSD = "pd-ssd"

	// DiskTypeStandard indicates a standard (non-SSD) disk
	DiskTypeStandard = "pd-standard"

	// DefaultDiskType indicates what type of disk to create by default
	DefaultDiskType = DiskTypeSSD
)

func init() {
	r := gofigCore.NewRegistration("GCEPD")
	r.Key(gofig.String, "", "",
		"If defined, location of JSON keyfile for service account",
		"gcepd.keyfile")
	r.Key(gofig.String, "", "",
		"If defined, limit GCE access to given zone", "gcepd.zone")
	r.Key(gofig.String, "", DefaultDiskType, "Default GCE disk type",
		"gcepd.defaultDiskType")
	r.Key(gofig.String, "", "", "Tag to apply and filter disks",
		"gcepd.tag")

	gofigCore.Register(r)
}
