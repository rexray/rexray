package cli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/util"
	version "github.com/emccode/rexray/version_info"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v1"
)

//AddCommands function
func initCommands() {
	RexrayCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(volumeCreateCmd)
	volumeCmd.AddCommand(volumeGetCmd)
	volumeCmd.AddCommand(volumeMapCmd)
	volumeCmd.AddCommand(volumePathCmd)
	volumeCmd.AddCommand(volumeMountCmd)
	volumeCmd.AddCommand(volumeAttachCmd)
	volumeCmd.AddCommand(volumeUnmountCmd)
	volumeCmd.AddCommand(volumeDetachCmd)
	volumeCmd.AddCommand(volumeRemoveCmd)

	RexrayCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotGetCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotCopyCmd)
	snapshotCmd.AddCommand(snapshotRemoveCmd)

	RexrayCmd.AddCommand(deviceCmd)
	deviceCmd.AddCommand(deviceGetCmd)
	deviceCmd.AddCommand(deviceFormatCmd)
	deviceCmd.AddCommand(deviceMountCmd)
	deviceCmd.AddCommand(devuceUnmountCmd)

	RexrayCmd.AddCommand(adapterCmd)
	adapterCmd.AddCommand(adapterGetTypesCmd)
	adapterCmd.AddCommand(adapterGetInstancesCmd)

	RexrayCmd.AddCommand(serviceStartCmd)
	RexrayCmd.AddCommand(serviceRestartCmd)
	RexrayCmd.AddCommand(serviceStopCmd)
	RexrayCmd.AddCommand(serviceStatusCmd)

	RexrayCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceInitSysCmd)

	serviceCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleTypesCmd)
	moduleCmd.AddCommand(moduleInstancesCmd)
	moduleInstancesCmd.AddCommand(moduleInstancesListCmd)
	moduleInstancesCmd.AddCommand(moduleInstancesCreateCmd)
	moduleInstancesCmd.AddCommand(moduleInstancesStartCmd)

	RexrayCmd.AddCommand(installCmd)
	RexrayCmd.AddCommand(uninstallCmd)

	RexrayCmd.AddCommand(versionCmd)
}

//RexrayCmd
var RexrayCmd = &cobra.Command{
	Use: "rexray",
	Short: "REX-Ray:\n" +
		"  A guest-based storage introspection tool that enables local\n" +
		"  visibility and management from cloud and storage platforms.",
	PersistentPreRun: preRun,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "The service controller",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "The module manager",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {

		epochInt, epochIntErr := strconv.ParseInt(version.Epoch, 10, 64)
		if epochIntErr != nil {
			panic(epochIntErr)
		}
		buildDate := time.Unix(epochInt, 0)

		fmt.Printf("SemVer: %s\n", version.SemVer)
		fmt.Printf("Binary: %s\n", version.Arch)
		fmt.Printf("Branch: %s\n", version.Branch)
		fmt.Printf("Commit: %s\n", version.ShaLong)
		fmt.Printf("Formed: %s\n", buildDate.Format(time.RFC1123))
	},
}

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "The volume manager",
	Run: func(cmd *cobra.Command, args []string) {
		if isHelpFlags(cmd) {
			cmd.Usage()
		} else {
			volumeGetCmd.Run(volumeGetCmd, args)
		}
	},
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "The snapshot manager",
	Run: func(cmd *cobra.Command, args []string) {
		if isHelpFlags(cmd) {
			cmd.Usage()
		} else {
			snapshotGetCmd.Run(snapshotGetCmd, args)
		}
	},
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "The device manager",
	Run: func(cmd *cobra.Command, args []string) {
		if isHelpFlags(cmd) {
			cmd.Usage()
		} else {
			deviceGetCmd.Run(deviceGetCmd, args)
		}
	},
}

var moduleTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List the available module types and their IDs",
	Run: func(cmd *cobra.Command, args []string) {

		_, addr, addrErr := util.ParseAddress(c.Host)
		if addrErr != nil {
			panic(addrErr)
		}

		u := fmt.Sprintf("http://%s/r/module/types", addr)

		client := &http.Client{}
		resp, respErr := client.Get(u)
		if respErr != nil {
			panic(respErr)
		}

		defer resp.Body.Close()
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			panic(bodyErr)
		}

		fmt.Println(string(body))
	},
}

var moduleInstancesCmd = &cobra.Command{
	Use:   "instance",
	Short: "The module instance manager",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var moduleInstancesListCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"list"},
	Short:   "List the running module instances",
	Run: func(cmd *cobra.Command, args []string) {

		_, addr, addrErr := util.ParseAddress(c.Host)
		if addrErr != nil {
			panic(addrErr)
		}

		u := fmt.Sprintf("http://%s/r/module/instances", addr)

		client := &http.Client{}
		resp, respErr := client.Get(u)
		if respErr != nil {
			panic(respErr)
		}

		defer resp.Body.Close()
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			panic(bodyErr)
		}

		fmt.Println(string(body))
	},
}

var moduleInstancesCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new module instance",
	Run: func(cmd *cobra.Command, args []string) {

		_, addr, addrErr := util.ParseAddress(c.Host)
		if addrErr != nil {
			panic(addrErr)
		}

		if moduleTypeId == -1 || moduleInstanceAddress == "" {
			cmd.Usage()
			return
		}

		modTypeIdStr := fmt.Sprintf("%d", moduleTypeId)
		modInstStartStr := fmt.Sprintf("%v", moduleInstanceStart)

		u := fmt.Sprintf("http://%s/r/module/instances", addr)
		cfgJson, cfgJsonErr := c.ToJson()

		if cfgJsonErr != nil {
			panic(cfgJsonErr)
		}

		log.WithFields(log.Fields{
			"url":     u,
			"typeId":  modTypeIdStr,
			"address": moduleInstanceAddress,
			"start":   modInstStartStr,
			"config":  cfgJson}).Debug("post create module instance")

		client := &http.Client{}
		resp, respErr := client.PostForm(u,
			url.Values{
				"typeId":  {modTypeIdStr},
				"address": {moduleInstanceAddress},
				"start":   {modInstStartStr},
				"config":  {cfgJson},
			})
		if respErr != nil {
			panic(respErr)
		}

		defer resp.Body.Close()
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			panic(bodyErr)
		}

		fmt.Println(string(body))
	},
}

var moduleInstancesStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a module instance",
	Run: func(cmd *cobra.Command, args []string) {

		_, addr, addrErr := util.ParseAddress(c.Host)
		if addrErr != nil {
			panic(addrErr)
		}

		if moduleInstanceId == -1 {
			cmd.Usage()
			return
		}

		u := fmt.Sprintf(
			"http://%s/r/module/instances/%d/start", addr, moduleInstanceId)

		client := &http.Client{}
		resp, respErr := client.Get(u)
		if respErr != nil {
			panic(respErr)
		}

		defer resp.Body.Close()
		body, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			panic(bodyErr)
		}

		fmt.Println(string(body))
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install REX-Ray",
	Run: func(cmd *cobra.Command, args []string) {
		install()
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall REX-Ray",
	Run: func(cmd *cobra.Command, args []string) {
		pkgManager, _ := cmd.Flags().GetBool("package")
		uninstall(pkgManager)
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:     "restart",
	Aliases: []string{"reload", "force-reload"},
	Short:   "Restart the service",
	Run: func(cmd *cobra.Command, args []string) {
		restart()
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	Run: func(cmd *cobra.Command, args []string) {
		stop()
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the service status",
	Run: func(cmd *cobra.Command, args []string) {
		status()
	},
}

var serviceInitSysCmd = &cobra.Command{
	Use:   "initsys",
	Short: "Print the detected init system type",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\nInit System: %s\n", GetInitSystemCmd())
	},
}

var adapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "The adapter manager",
	Run: func(cmd *cobra.Command, args []string) {
		if isHelpFlags(cmd) {
			cmd.Usage()
		} else {
			adapterGetTypesCmd.Run(adapterGetTypesCmd, args)
		}
	},
}

var adapterGetTypesCmd = &cobra.Command{
	Use:     "types",
	Short:   "List the available adapter types",
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {
		drivers := sdm.GetDriverNames()
		for n := range drivers {
			fmt.Printf("Storage Driver: %v\n", drivers[n])
		}
	},
}

var adapterGetInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List the configured adapter instances",
	Run: func(cmd *cobra.Command, args []string) {

		allInstances, err := sdm.GetInstance()
		if err != nil {
			panic(err)
		}

		if len(allInstances) > 0 {
			yamlOutput, err := yaml.Marshal(&allInstances)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(string(yamlOutput))
		}
	},
}

var volumeMapCmd = &cobra.Command{
	Use:   "map",
	Short: "Print the volume mapping(s)",
	Run: func(cmd *cobra.Command, args []string) {

		allBlockDevices, err := sdm.GetVolumeMapping()
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		if len(allBlockDevices) > 0 {
			yamlOutput, err := yaml.Marshal(&allBlockDevices)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(string(yamlOutput))
		}
	},
}

var volumeGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get one or more volumes",
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {

		allVolumes, err := sdm.GetVolume(volumeID, volumeName)
		if err != nil {
			log.Fatal(err)
		}

		if len(allVolumes) > 0 {
			yamlOutput, err := yaml.Marshal(&allVolumes)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(string(yamlOutput))
		}
	},
}

var snapshotGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get one or more snapshots",
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {

		allSnapshots, err := sdm.GetSnapshot(volumeID, snapshotID, snapshotName)
		if err != nil {
			log.Fatal(err)
		}

		if len(allSnapshots) > 0 {
			yamlOutput, err := yaml.Marshal(&allSnapshots)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(string(yamlOutput))
		}
	},
}

var snapshotCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new snapshot",
	Aliases: []string{"new"},
	Run: func(cmd *cobra.Command, args []string) {

		if volumeID == "" {
			log.Fatalf("missing --volumeid")
		}

		snapshot, err := sdm.CreateSnapshot(runAsync, snapshotName, volumeID, description)
		if err != nil {
			log.Fatal(err)
		}

		yamlOutput, err := yaml.Marshal(&snapshot)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(yamlOutput))

	},
}

var snapshotRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a snapshot",
	Aliases: []string{"rm"},
	Run: func(cmd *cobra.Command, args []string) {

		if snapshotID == "" {
			log.Fatalf("missing --snapshotid")
		}

		err := sdm.RemoveSnapshot(snapshotID)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var volumeCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new volume",
	Aliases: []string{"new"},
	Run: func(cmd *cobra.Command, args []string) {

		if size == 0 && snapshotID == "" && volumeID == "" {
			log.Fatalf("missing --size")
		}

		volume, err := sdm.CreateVolume(
			runAsync, volumeName, volumeID, snapshotID,
			volumeType, IOPS, size, availabilityZone)
		if err != nil {
			log.Fatal(err)
		}

		yamlOutput, err := yaml.Marshal(&volume)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(yamlOutput))

	},
}

var volumeRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a volume",
	Aliases: []string{"create"},
	Run: func(cmd *cobra.Command, args []string) {

		if volumeID == "" {
			log.Fatalf("missing --volumeid")
		}

		err := sdm.RemoveVolume(volumeID)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var volumeAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach a volume",
	Run: func(cmd *cobra.Command, args []string) {

		if volumeID == "" {
			log.Fatalf("missing --volumeid")
		}

		volumeAttachment, err := sdm.AttachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			log.Fatal(err)
		}

		yamlOutput, err := yaml.Marshal(&volumeAttachment)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(yamlOutput))

	},
}

var volumeDetachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach a volume",
	Run: func(cmd *cobra.Command, args []string) {

		if volumeID == "" {
			log.Fatalf("missing --volumeid")
		}

		err := sdm.DetachVolume(runAsync, volumeID, instanceID)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var snapshotCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copie a snapshot",
	Run: func(cmd *cobra.Command, args []string) {

		if snapshotID == "" && volumeID == "" && volumeName == "" {
			log.Fatalf("missing --volumeid or --snapshotid or --volumename")
		}

		snapshot, err := sdm.CopySnapshot(
			runAsync, volumeID, snapshotID,
			snapshotName, destinationSnapshotName, destinationRegion)
		if err != nil {
			log.Fatal(err)
		}

		yamlOutput, err := yaml.Marshal(&snapshot)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(yamlOutput))

	},
}

var deviceGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get a device's mount(s)",
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {

		mounts, err := osdm.GetMounts(deviceName, mountPoint)
		if err != nil {
			log.Fatal(err)
		}

		yamlOutput, err := yaml.Marshal(&mounts)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(yamlOutput))
	},
}

var deviceMountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount a device",
	Run: func(cmd *cobra.Command, args []string) {

		if deviceName == "" || mountPoint == "" {
			log.Fatal("Missing --devicename and --mountpoint")
		}

		// mountOptions = fmt.Sprintf("val,%s", mountOptions)
		err := osdm.Mount(deviceName, mountPoint, mountOptions, mountLabel)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var devuceUnmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount a device",
	Run: func(cmd *cobra.Command, args []string) {

		if mountPoint == "" {
			log.Fatal("Missing --mountpoint")
		}

		err := osdm.Unmount(mountPoint)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var deviceFormatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format a device",
	Run: func(cmd *cobra.Command, args []string) {

		if deviceName == "" {
			log.Fatal("Missing --devicename")
		}

		if fsType == "" {
			fsType = "ext4"
		}

		err := osdm.Format(deviceName, fsType, overwriteFs)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var volumeMountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount a volume",
	Run: func(cmd *cobra.Command, args []string) {

		if volumeName == "" && volumeID == "" {
			log.Fatal("Missing --volumename or --volumeid")
		}

		mountPath, err := vdm.Mount(volumeName, volumeID, overwriteFs, fsType)
		if err != nil {
			log.Fatal(err)
		}

		yamlOutput, err := yaml.Marshal(&mountPath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(yamlOutput))

	},
}

var volumeUnmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount a volume",
	Run: func(cmd *cobra.Command, args []string) {

		if volumeName == "" && volumeID == "" {
			log.Fatal("Missing --volumename or --volumeid")
		}

		err := vdm.Unmount(volumeName, volumeID)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var volumePathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the volume path",
	Run: func(cmd *cobra.Command, args []string) {

		if volumeName == "" && volumeID == "" {
			log.Fatal("Missing --volumename or --volumeid")
		}

		mountPath, err := vdm.Path(volumeName, volumeID)
		if err != nil {
			log.Fatal(err)
		}

		if mountPath != "" {
			yamlOutput, err := yaml.Marshal(&mountPath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(string(yamlOutput))
		}
	},
}
