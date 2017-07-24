package gcepd

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "gcepd"

	defaultStatusMaxAttempts  = 10
	defaultStatusInitDelay    = "100ms"
	defaultConvertUnderscores = false

	/* This is hard deadline when waiting for the volume status to change to
	a desired state. At minimum is has to be more than the expontential
	backoff of sum 100*2^x, x=0 to 9 == 102s3ms, but should also account for
	RTT of API requests, and how many API requests would be made to
	exhaust retries */
	defaultStatusTimeout = "2m"

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

	// ConfigKeyfile is the key for the service account JSON credential file
	ConfigKeyfile = Name + ".keyfile"

	// ConfigZone is the key for the availability zone
	ConfigZone = Name + ".zone"

	// ConfigDefaultDiskType is the key for the default disk type to use
	ConfigDefaultDiskType = Name + ".defaultDiskType"

	// ConfigTag is the key for the tag to apply to and filter disks
	ConfigTag = Name + ".tag"

	// ConfigStatusMaxAttempts is the key for the maximum number of times
	// a volume status will be queried when waiting for an action to finish
	ConfigStatusMaxAttempts = Name + ".statusMaxAttempts"

	// ConfigStatusInitDelay is the key for the initial time duration
	// for exponential backoff
	ConfigStatusInitDelay = Name + ".statusInitialDelay"

	// ConfigStatusTimeout is the key for the time duration for a timeout
	// on how long to wait for a desired volume status to appears
	ConfigStatusTimeout = Name + ".statusTimeout"

	// ConfigConvertUnderscores is the key for a boolean flag on whether
	// incoming requests that have names with underscores should be
	// converted to dashes to satisfy GCE naming requirements
	ConfigConvertUnderscores = Name + ".convertUnderscores"
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
	r.Key(gofig.Int, "", defaultStatusMaxAttempts, "Max Status Attempts",
		ConfigStatusMaxAttempts)
	r.Key(gofig.String, "", defaultStatusInitDelay, "Status Initial Delay",
		ConfigStatusInitDelay)
	r.Key(gofig.String, "", defaultStatusTimeout, "Status Timeout",
		ConfigStatusTimeout)
	r.Key(gofig.Bool, "", defaultConvertUnderscores,
		"Convert Underscores", ConfigConvertUnderscores)

	gofigCore.Register(r)
}
