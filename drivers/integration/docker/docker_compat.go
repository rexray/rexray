package docker

import "github.com/akutz/gofig"

const (
	//ConfigOldRoot is a config key.
	ConfigOldRoot = "docker"

	//ConfigOldDockerFsType is a config key.
	ConfigOldDockerFsType = ConfigOldRoot + ".fsType"

	//ConfigOldDockerVolumeType is a  config key.
	ConfigOldDockerVolumeType = ConfigOldRoot + ".volumeType"

	//ConfigOldDockerIOPS is a config key.
	ConfigOldDockerIOPS = ConfigOldRoot + ".iops"

	//ConfigOldDockerSize is a config key.
	ConfigOldDockerSize = ConfigOldRoot + ".size"

	//ConfigOldDockerAvailabilityZone is a config key.
	ConfigOldDockerAvailabilityZone = ConfigOldRoot + ".availabilityZone"

	//ConfigOldDockerMountDirPath is a config key.
	ConfigOldDockerMountDirPath = ConfigOldRoot + ".mountDirPath"

	//ConfigOldDockerLinuxVolumeRootPath is a config key.
	ConfigOldDockerLinuxVolumeRootPath = "linux.volume.rootpath"
)

func backCompat(config gofig.Config) {
	checks := [][]string{
		{ConfigDockerFsType, ConfigOldDockerFsType},
		{ConfigDockerVolumeType, ConfigOldDockerVolumeType},
		{ConfigDockerIOPS, ConfigOldDockerIOPS},
		{ConfigDockerSize, ConfigOldDockerSize},
		{ConfigDockerAvailabilityZone, ConfigOldDockerAvailabilityZone},
		{ConfigDockerMountDirPath, ConfigOldDockerMountDirPath},
		{ConfigDockerLinuxVolumeRootPath, ConfigOldDockerLinuxVolumeRootPath},
	}
	for _, check := range checks {
		if !config.IsSet(check[0]) && config.IsSet(check[1]) {
			config.Set(check[0], config.Get(check[1]))
		}
	}
}
