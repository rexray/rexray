package cli

import (
	"fmt"

	"github.com/emccode/rexray/util"
	"github.com/spf13/cobra"
)

func initFlags() {
	cobra.HelpFlagShorthand = "?"
	cobra.HelpFlagUsageFormatString = "Help for %s"

	initGlobalFlags()
	initServiceFlags()
	initInstallerFlags()
	initVolumeFlags()
	initSnapshotFlags()
	initDeviceFlags()
	initModuleFlags()
}

func initGlobalFlags() {
	RexrayCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c",
		fmt.Sprintf("%s/.rexray/config.yml", util.HomeDir()),
		"The REX-Ray configuration file")
	RexrayCmd.PersistentFlags().BoolP(
		"verbose", "v", false, "Print verbose help information")

	RexrayCmd.PersistentFlags().AddFlagSet(r.Config.GlobalFlags)
	RexrayCmd.PersistentFlags().AddFlagSet(r.Config.AdditionalFlags)
}

func initServiceFlags() {
	serviceStartCmd.Flags().BoolVarP(&fg, "foreground", "f", false,
		"Starts the service in the foreground")
	serviceStartCmd.Flags().StringVarP(&client, "client", "", "",
		"Socket the daemon uses to communicate to the client")
	serviceStartCmd.Flags().BoolVarP(&force, "force", "", false,
		"Forces the service to start, ignoring errors")
}

func initInstallerFlags() {
	uninstallCmd.Flags().Bool("package", false,
		"A flag indicating a package manager is performing the uninstallation")
}

func initVolumeFlags() {
	volumeGetCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	volumeGetCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeCreateCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	volumeCreateCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	volumeCreateCmd.Flags().StringVar(&volumeType, "volumetype", "", "volumetype")
	volumeCreateCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeCreateCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	volumeCreateCmd.Flags().Int64Var(&iops, "iops", 0, "IOPS")
	volumeCreateCmd.Flags().Int64Var(&size, "size", 0, "size")
	volumeCreateCmd.Flags().StringVar(&availabilityZone, "availabilityzone", "", "availabilityzone")
	volumeRemoveCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeAttachCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	volumeAttachCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeAttachCmd.Flags().StringVar(&instanceID, "instanceid", "", "instanceid")
	volumeDetachCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	volumeDetachCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeDetachCmd.Flags().StringVar(&instanceID, "instanceid", "", "instanceid")
	volumeMountCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeMountCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	volumeMountCmd.Flags().BoolVar(&overwriteFs, "overwritefs", false, "overwritefs")
	volumeMountCmd.Flags().StringVar(&fsType, "fstype", "", "fstype")
	volumeUnmountCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeUnmountCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	volumePathCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumePathCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
}

func initDeviceFlags() {
	deviceGetCmd.Flags().StringVar(&deviceName, "devicename", "", "devicename")
	deviceGetCmd.Flags().StringVar(&mountPoint, "mountpoint", "", "mountpoint")
	deviceMountCmd.Flags().StringVar(&deviceName, "devicename", "", "devicename")
	deviceMountCmd.Flags().StringVar(&mountPoint, "mountpoint", "", "mountpoint")
	deviceMountCmd.Flags().StringVar(&mountOptions, "mountoptions", "", "mountoptions")
	deviceMountCmd.Flags().StringVar(&mountLabel, "mountlabel", "", "mountlabel")
	devuceUnmountCmd.Flags().StringVar(&mountPoint, "mountpoint", "", "mountpoint")
	deviceFormatCmd.Flags().StringVar(&deviceName, "devicename", "", "devicename")
	deviceFormatCmd.Flags().StringVar(&fsType, "fstype", "", "fstype")
	deviceFormatCmd.Flags().BoolVar(&overwriteFs, "overwritefs", false, "overwritefs")
}

func initSnapshotFlags() {
	snapshotGetCmd.Flags().StringVar(&snapshotName, "snapshotname", "", "snapshotname")
	snapshotGetCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	snapshotGetCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	snapshotCreateCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	snapshotCreateCmd.Flags().StringVar(&snapshotName, "snapshotname", "", "snapshotname")
	snapshotCreateCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	snapshotCreateCmd.Flags().StringVar(&description, "description", "", "description")
	snapshotRemoveCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	snapshotCopyCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	snapshotCopyCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	snapshotCopyCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	snapshotCopyCmd.Flags().StringVar(&snapshotName, "snapshotname", "", "snapshotname")
	snapshotCopyCmd.Flags().StringVar(&destinationSnapshotName, "destinationsnapshotname", "", "destinationsnapshotname")
	snapshotCopyCmd.Flags().StringVar(&destinationRegion, "destinationregion", "", "destinationregion")
}

func initModuleFlags() {
	moduleInstancesCreateCmd.Flags().Int32VarP(&moduleTypeID, "id",
		"i", -1, "The ID of the module type to instance")

	moduleInstancesCreateCmd.Flags().StringVarP(&moduleInstanceAddress,
		"address", "a", "",
		"The network address at which the module will be exposed")

	moduleInstancesCreateCmd.Flags().BoolVarP(&moduleInstanceStart,
		"start", "s", false,
		"A flag indicating whether or not to start the module upon creation")

	moduleInstancesCreateCmd.Flags().StringSliceVarP(&moduleConfig,
		"options", "o", nil,
		"A comma-seperated string of key=value pairs used by some module "+
			"types for custom configuraitons.")

	moduleInstancesStartCmd.Flags().Int32VarP(&moduleInstanceID, "id",
		"i", -1, "The ID of the module instance to start")
}
