package docker

import "github.com/emccode/libstorage/api/types"

const (
	// ConfigRoot is a config key.
	ConfigRoot = types.ConfigIntegration + ".docker"

	//ConfigDockerFsType is a config key.
	ConfigDockerFsType = ConfigRoot + ".fsType"

	//ConfigDockerVolumeType is a  config key.
	ConfigDockerVolumeType = ConfigRoot + ".volumeType"

	//ConfigDockerIOPS is a config key.
	ConfigDockerIOPS = ConfigRoot + ".iops"

	//ConfigDockerSize is a config key.
	ConfigDockerSize = ConfigRoot + ".size"

	//ConfigDockerAvailabilityZone is a config key.
	ConfigDockerAvailabilityZone = ConfigRoot + ".availabilityZone"

	//ConfigDockerMountDirPath is a config key.
	ConfigDockerMountDirPath = ConfigRoot + ".mountDirPath"

	//ConfigDockerLinuxVolumeRootPath is a config key.
	ConfigDockerLinuxVolumeRootPath = ConfigRoot + ".linux.volume.rootpath"
)
