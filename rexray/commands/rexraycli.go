package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/emccode/rexray"
	rrdaemon "github.com/emccode/rexray/daemon"
	rros "github.com/emccode/rexray/os"
	"github.com/emccode/rexray/storage"
	"github.com/emccode/rexray/util"
	"github.com/emccode/rexray/version"
	"github.com/emccode/rexray/volume"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v1"
)

const REXHOME = "/opt/rexray"
const EXEFILE = "/opt/rexray/rexray"
const ENVFILE = "/etc/rexray/rexray.env"
const CFGFILE = "/etc/rexray/rexray.conf"
const UNTFILE = "/etc/systemd/system/rexray.service"
const INTFILE = "/etc/init.d/rexray"

// init system types
const (
	UNKNOWN = iota
	SYSTEMD
	UPDATERCD
	CHKCONFIG
)

var (
	homeDir                 string
	debug                   bool
	client                  string
	host                    string
	fg                      bool
	cfgFile                 string
	snapshotID              string
	volumeID                string
	runAsync                bool
	description             string
	volumeType              string
	IOPS                    int64
	size                    int64
	instanceID              string
	volumeName              string
	snapshotName            string
	availabilityZone        string
	destinationSnapshotName string
	destinationRegion       string
	deviceName              string
	mountPoint              string
	mountOptions            string
	mountLabel              string
	fsType                  string
	overwriteFs             bool
	moduleTypeId            int32
	moduleInstanceId        int32
	moduleInstanceAddress   string
	moduleInstanceStart     bool
)

//RexrayCmd
var RexrayCmd = &cobra.Command{
	Use: "rexray",
	Short: "REX-Ray:\n" +
		"  A guest-based storage introspection tool that enables local\n" +
		"  visibility and management from cloud and storage platforms.",
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		cmd.Help()
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
		fmt.Printf("%v\n", version.Version)
	},
}

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "The volume manager",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "The snapshot manager",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "The device manager",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var moduleTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List the available module types and their IDs",
	Run: func(cmd *cobra.Command, args []string) {

		_, addr, addrErr := util.ParseAddress(host)
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

		_, addr, addrErr := util.ParseAddress(host)
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

		_, addr, addrErr := util.ParseAddress(host)
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

		client := &http.Client{}
		resp, respErr := client.PostForm(u, url.Values{
			"typeId":  {modTypeIdStr},
			"address": {moduleInstanceAddress},
			"start":   {modInstStartStr}})
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

		_, addr, addrErr := util.ParseAddress(host)
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

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the service into SystemV or SystemD",
	Run: func(cmd *cobra.Command, args []string) {
		Install()
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run: func(cmd *cobra.Command, args []string) {
		Start()
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	Run: func(cmd *cobra.Command, args []string) {
		Stop()
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the service status",
	Run: func(cmd *cobra.Command, args []string) {
		Status()
	},
}

var serviceInitSysCmd = &cobra.Command{
	Use:   "initsys",
	Short: "Print the detected init system type",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\nInit System: %s\n", GetInitSystemCmd())
	},
}

var driverCmd = &cobra.Command{
	Use:   "driver",
	Short: "The driver manager",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var driverGetTypesCmd = &cobra.Command{
	Use:     "types",
	Short:   "List the available driver types",
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {
		drivers := storage.GetDriverNames()
		for n := range drivers {
			fmt.Printf("\nStorage Driver: %v\n", drivers[n])
		}
	},
}

var driverGetInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List the configured driver instances",
	Run: func(cmd *cobra.Command, args []string) {

		allInstances, err := storage.GetInstance()
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

		allBlockDevices, err := storage.GetVolumeMapping()
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

		allVolumes, err := storage.GetVolume(volumeID, volumeName)
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

		allSnapshots, err := storage.GetSnapshot(volumeID, snapshotID, snapshotName)
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

		snapshot, err := storage.CreateSnapshot(runAsync, snapshotName, volumeID, description)
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

		err := storage.RemoveSnapshot(snapshotID)
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

		volume, err := storage.CreateVolume(runAsync, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
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

		err := storage.RemoveVolume(volumeID)
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

		volumeAttachment, err := storage.AttachVolume(runAsync, volumeID, instanceID)
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

		err := storage.DetachVolume(runAsync, volumeID, instanceID)
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

		snapshot, err := storage.CopySnapshot(runAsync, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion)
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

		mounts, err := rros.GetMounts(deviceName, mountPoint)
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
		err := rros.Mount(deviceName, mountPoint, mountOptions, mountLabel)
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

		err := rros.Unmount(mountPoint)
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

		err := rros.Format(deviceName, fsType, overwriteFs)
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

		mountPath, err := volume.Mount(volumeName, volumeID, overwriteFs, fsType)
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

		err := volume.Unmount(volumeName, volumeID)
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

		mountPath, err := volume.Path(volumeName, volumeID)
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

//Exec function
func Exec() {
	RexrayCmd.Execute()
}

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

	RexrayCmd.AddCommand(driverCmd)
	driverCmd.AddCommand(driverGetTypesCmd)
	driverCmd.AddCommand(driverGetInstancesCmd)

	RexrayCmd.AddCommand(serviceStartCmd)
	RexrayCmd.AddCommand(serviceStopCmd)
	RexrayCmd.AddCommand(serviceStatusCmd)

	RexrayCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceInitSysCmd)

	serviceCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleTypesCmd)
	moduleCmd.AddCommand(moduleInstancesCmd)
	moduleInstancesCmd.AddCommand(moduleInstancesListCmd)
	moduleInstancesCmd.AddCommand(moduleInstancesCreateCmd)
	moduleInstancesCmd.AddCommand(moduleInstancesStartCmd)

	RexrayCmd.AddCommand(versionCmd)
}

func newStringFlag(
	flagSet *pflag.FlagSet, pVar *string, name, shortName, value, desc string) {
	flagSet.StringVarP(pVar, name, shortName, value, desc)
	viper.BindPFlag(name, flagSet.Lookup(name))
}

func newInt32Flag(
	flagSet *pflag.FlagSet, pVar *int32, name, shortName string, value int32, desc string) {
	flagSet.Int32VarP(pVar, name, shortName, value, desc)
	viper.BindPFlag(name, flagSet.Lookup(name))
}

func newBoolFlag(
	flagSet *pflag.FlagSet, pVar *bool, name, shortName string, value bool, desc string) {
	flagSet.BoolVarP(pVar, name, shortName, value, desc)
	viper.BindPFlag(name, flagSet.Lookup(name))
}

func init() {
	initHomeDir()
	initCommands()
	initFlags()
	initUsageTemplates()
}

func initHomeDir() {
	homeDir = "$HOME"
	curUser, curUserErr := user.Current()
	if curUserErr == nil {
		homeDir = curUser.HomeDir
	}
}

func initFlags() {
	cobra.HelpFlagShorthand = "?"
	cobra.HelpFlagUsageFormatString = "Help for %s"

	initGlobalFlags()
	initServiceFlags()
	initVolumeFlags()
	initSnapshotFlags()
	initDeviceFlags()
	initModuleFlags()
}

func initGlobalFlags() {
	newStringFlag(RexrayCmd.PersistentFlags(), &cfgFile, "config", "c",
		fmt.Sprintf("%s/.rexray/config.yaml", homeDir),
		"The REX-Ray configuration file")
	newBoolFlag(RexrayCmd.PersistentFlags(), &debug, "debug", "d", false,
		"Enables verbose output")
	newStringFlag(RexrayCmd.PersistentFlags(), &host, "host", "h",
		"tcp://127.0.0.1:7979", "The REX-Ray service address")
}

func initServiceFlags() {
	newBoolFlag(serviceStartCmd.Flags(), &fg, "foreground", "f", false,
		"Starts the service in the foreground")
	newStringFlag(serviceStartCmd.Flags(), &client, "client", "", "",
		"Socket the daemon uses to communicate to the client")
}

func initVolumeFlags() {
	volumeGetCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	volumeGetCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeCreateCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	volumeCreateCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	volumeCreateCmd.Flags().StringVar(&volumeType, "volumetype", "", "volumetype")
	volumeCreateCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	volumeCreateCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	volumeCreateCmd.Flags().Int64Var(&IOPS, "iops", 0, "IOPS")
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
	newInt32Flag(moduleInstancesCreateCmd.Flags(), &moduleTypeId, "id",
		"i", -1, "The ID of the module type to instance")

	newStringFlag(moduleInstancesCreateCmd.Flags(), &moduleInstanceAddress,
		"address", "a", "",
		"The network address at which the module will be exposed")

	newBoolFlag(moduleInstancesCreateCmd.Flags(), &moduleInstanceStart,
		"start", "s", false,
		"A flag indicating whether or not to start the module upon creation")

	newInt32Flag(moduleInstancesStartCmd.Flags(), &moduleInstanceId, "id",
		"i", -1, "The ID of the module instance to start")
}

func initUsageTemplates() {
	serviceCmd.SetUsageTemplate(RexrayCmd.UsageTemplate())
	RexrayCmd.SetUsageTemplate(UsageTemplate)
}

//InitConfig function
func InitConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/rexray")
	viper.AddConfigPath("$HOME/.rexray")

	viper.ReadInConfig()
	// if err != nil {
	// 	fmt.Println("No configuration file loaded - using defaults")
	// }

	viper.AutomaticEnv()
	viper.SetEnvPrefix("REXRAY")
}

func GetInitSystemCmd() string {
	switch GetInitSystemType() {
	case SYSTEMD:
		return "systemd"
	case UPDATERCD:
		return "update-rc.d"
	case CHKCONFIG:
		return "chkconfig"
	default:
		return "unknown"
	}
}

func GetInitSystemType() int {
	if FileExistsInPath("systemctl") {
		return SYSTEMD
	}

	if FileExistsInPath("update-rc.d") {
		return UPDATERCD
	}

	if FileExistsInPath("chkconfig") {
		return CHKCONFIG
	}

	return UNKNOWN
}

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func FileExistsInPath(fileName string) bool {
	_, err := exec.LookPath(fileName)
	return err == nil
}

func GetPathParts(path string) (dirPath, fileName, absPath string) {
	absPath, _ = filepath.Abs(path)
	dirPath = filepath.Dir(absPath)
	fileName = filepath.Base(absPath)
	return
}

func GetThisPathParts() (dirPath, fileName, absPath string) {
	return GetPathParts(os.Args[0])
}

const (
	LETTER_BYTES      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LETTER_INDEX_BITS = 6
	LETTER_INDEX_MASK = 1<<LETTER_INDEX_BITS - 1
	LETTER_INDEX_MAX  = 63 / LETTER_INDEX_BITS
)

func RandomString(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)
	for i, cache, remain := length-1, src.Int63(), LETTER_INDEX_MAX; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), LETTER_INDEX_MAX
		}
		if idx := int(cache & LETTER_INDEX_MASK); idx < len(LETTER_BYTES) {
			b[i] = LETTER_BYTES[idx]
			i--
		}
		cache >>= LETTER_INDEX_BITS
		remain--
	}

	return string(b)
}

func Install() {
	checkOpPerms("installed")

	_, _, thisAbsPath := GetThisPathParts()

	exeDir := filepath.Dir(EXEFILE)
	os.MkdirAll(exeDir, 0755)
	os.Chown(exeDir, 0, 0)

	exec.Command("cp", "-f", thisAbsPath, EXEFILE).Run()
	os.Chown(EXEFILE, 0, 0)
	exec.Command("chmod", "4755", EXEFILE).Run()

	switch GetInitSystemType() {
	case SYSTEMD:
		installSystemD()
	case UPDATERCD:
		installUpdateRcd()
	case CHKCONFIG:
		installChkConfig()
	}
}

func Start() {
	checkOpPerms("started")

	if debug {
		log.Printf(strings.Join(os.Args, " "))
	}

	pidFile := util.PidFile()

	if FileExists(pidFile) {
		pid, pidErr := util.ReadPidFile()
		if pidErr != nil {
			fmt.Printf("Error reading REX-Ray PID file at %s\n", pidFile)
		} else {
			fmt.Printf("REX-Ray already running at PID %d\n", pid)
		}
		panic(1)
	}

	if fg || client != "" {
		startDaemon()
	} else {
		tryToStartDaemon()
	}
}

func failOnError(err error) {
	if err != nil {
		fmt.Printf("FAILED!\n  %v\n", err)
		panic(err)
	}
}

func startDaemon() {

	fmt.Printf("%s\n", RexRayLogoAscii)

	var success []byte
	var failure []byte
	var conn net.Conn

	if !fg {

		success = []byte{0}
		failure = []byte{1}

		var dialErr error

		log.Printf("Dialing %s\n", client)
		conn, dialErr = net.Dial("unix", client)
		if dialErr != nil {
			panic(dialErr)
		}
	}

	writePidErr := util.WritePidFile(-1)
	if writePidErr != nil {
		if conn != nil {
			conn.Write(failure)
		}
		panic(writePidErr)
	}

	defer func() {
		r := recover()
		os.Remove(util.PidFile())
		if r != nil {
			panic(r)
		}
	}()

	log.Printf("Created pid file, pid=%d\n", os.Getpid())

	init := make(chan error)
	sigc := make(chan os.Signal, 1)
	stop := make(chan os.Signal)

	signal.Notify(sigc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		rrdaemon.Start(host, init, stop)
	}()

	initErrors := make([]error, 0)

	for initErr := range init {
		initErrors = append(initErrors, initErr)
		log.Println(initErr)
	}

	if conn != nil {
		if len(initErrors) == 0 {
			conn.Write(success)
		} else {
			conn.Write(failure)
		}

		conn.Close()
	}

	if len(initErrors) > 0 {
		return
	}

	sigv := <-sigc
	log.Printf("Received shutdown signal %v\n", sigv)
	stop <- sigv
}

func tryToStartDaemon() {
	_, _, thisAbsPath := GetThisPathParts()

	fmt.Print("Starting REX-Ray...")

	signal := make(chan byte)
	client := fmt.Sprintf("%s/%s.sock", os.TempDir(), RandomString(32))
	if debug {
		log.Printf("\nclient=%s\n", client)
	}

	l, lErr := net.Listen("unix", client)
	failOnError(lErr)

	go func() {
		conn, acceptErr := l.Accept()
		if acceptErr != nil {
			fmt.Printf("FAILED!\n  %v\n", acceptErr)
			panic(acceptErr)
		}
		defer conn.Close()
		defer os.Remove(client)

		if debug {
			log.Println("accepted connection")
		}

		buff := make([]byte, 1)
		conn.Read(buff)

		if debug {
			log.Println("received data")
		}

		signal <- buff[0]
	}()

	cmdArgs := []string{
		"start",
		fmt.Sprintf("--client=%s", client),
		fmt.Sprintf("--debug=%v", debug)}

	if host != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--host=%s", host))
	}

	cmd := exec.Command(thisAbsPath, cmdArgs...)

	logFile, logFileErr :=
		os.OpenFile(util.LogFile(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	failOnError(logFileErr)

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	cmdErr := cmd.Start()
	failOnError(cmdErr)

	sigVal := <-signal
	if sigVal != 0 {
		fmt.Println("FAILED!")
		panic(1)
	}

	pid, _ := util.ReadPidFile()
	fmt.Printf("SUCESS!\n\n")
	fmt.Printf("  The REX-Ray daemon is now running at PID %d. To\n", pid)
	fmt.Printf("  shutdown the daemon execute the following command:\n\n")
	fmt.Printf("    sudo %s stop\n\n", thisAbsPath)
}

func Stop() {
	checkOpPerms("stopped")

	if !FileExists(util.PidFile()) {
		fmt.Println("REX-Ray is already stopped")
		panic(1)
	}

	fmt.Print("Shutting down REX-Ray...")

	pid, pidErr := util.ReadPidFile()
	failOnError(pidErr)

	proc, procErr := os.FindProcess(pid)
	failOnError(procErr)

	killErr := proc.Signal(syscall.SIGHUP)
	failOnError(killErr)

	fmt.Println("SUCCESS!")
}

func Status() {
	if !FileExists(util.PidFile()) {
		fmt.Println("REX-Ray is stopped")
		return
	}
	pid, _ := util.ReadPidFile()
	fmt.Printf("REX-Ray is running at pid %d\n", pid)
}

func Restart() {
	checkOpPerms("restarted")

	if FileExists(util.PidFile()) {
		Stop()
	}

	Start()
}

func checkOpPerms(op string) {
	if os.Geteuid() != 0 {
		log.Fatalf("REX-Ray can only be %s by root\n", op)
	}
}

func installSystemD() {
	createUnitFile()
	createEnvFile()

	cmd := exec.Command("systemctl", "enable", "rexray.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("installation error %v", err)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/emccode/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo systemctl start rexray'.\n\n")
}

func installUpdateRcd() {
	createInitFile()
	cmd := exec.Command("update-rc.d", "rexray", "defaults")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("installation error %v", err)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/emccode/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo /etc/init.d/rexray start'.\n\n")
}

func installChkConfig() {
	createInitFile()
	cmd := exec.Command("chkconfig", "rexray", "on")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Fatalf("installation error %v", err)
	}

	fmt.Print("REX-Ray is now installed. Before starting it please check ")
	fmt.Print("http://github.com/emccode/rexray for instructions on how to ")
	fmt.Print("configure it.\n\n Once configured the REX-Ray service can be ")
	fmt.Print("started with the command 'sudo /etc/init.d/rexray start'.\n\n")
}

func createEnvFile() {

	envdir := filepath.Dir(ENVFILE)
	os.MkdirAll(envdir, 0755)

	f, err := os.OpenFile(ENVFILE, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("REXRAY_HOME=/opt/rexray")
}

func createUnitFile() {
	f, err := os.OpenFile(UNTFILE, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("[Unit]\n")
	f.WriteString("Description=rexray\n")
	f.WriteString("Before=docker.service\n")
	f.WriteString("\n")
	f.WriteString("[Service]\n")
	f.WriteString(fmt.Sprintf("EnvironmentFile=%s\n", ENVFILE))
	f.WriteString(fmt.Sprintf("ExecStart=%s --daemon\n", EXEFILE))
	f.WriteString("ExecReload=/bin/kill -HUP $MAINPID\n")
	f.WriteString("KillMode=process\n")
	f.WriteString("Restart=always\n")
	f.WriteString("StartLimitInterval=0\n")
	f.WriteString("\n")
	f.WriteString("[Install]\n")
	f.WriteString("WantedBy=docker.service\n")
	f.WriteString("\n")
}

func createInitFile() {
	os.Symlink(EXEFILE, INTFILE)
}

const RexRayLogoAscii = `
                          ⌐▄Q▓▄Ç▓▄,▄_                              
                         Σ▄▓▓▓▓▓▓▓▓▓▓▄π                            
                       ╒▓▓▌▓▓▓▓▓▓▓▓▓▓▀▓▄▄.                         
                    ,_▄▀▓▓ ▓▓ ▓▓▓▓▓▓▓▓▓▓▓█                         
                   │▄▓▓ _▓▓▓▓▓▓▓▓▓┌▓▓▓▓▓█                          
                  _J┤▓▓▓▓▓▓▓▓▓▓▓▓▓├█▓█▓▀Γ                          
            ,▄▓▓▓▓▓▓^██▓▓▓▓▓▓▓▓▓▓▓▓▄▀▄▄▓▓Ω▄                        
            F▌▓▌█ⁿⁿⁿ  ⁿ└▀ⁿ██▓▀▀▀▀▀▀▀▀▀▀▌▓▓▓▌                       
             'ⁿ_  ,▄▄▄▄▄▄▄▄▄█_▄▄▄▄▄▄▄▄▄ⁿ▀~██                       
               Γ  ├▓▓▓▓▓█▀ⁿ█▌▓Ω]█▓▓▓▓▓▓ ├▓                         
               │  ├▓▓▓▓▓▌≡,__▄▓▓▓█▓▓▓▓▓ ╞█~   Y,┐                  
               ╞  ├▓▓▓▓▓▄▄__^^▓▓▓▌▓▓▓▓▓  ▓   /▓▓▓                  
                  ├▓▓▓▓▓▓▓▄▄═▄▓▓▓▓▓▓▓▓▓  π ⌐▄▓▓█║n                 
                _ ├▓▓▓▓▓▓▓▓▓~▓▓▓▓▓▓▓▓▓▓  ▄4▄▓▓▓██                  
                µ ├▓▓▓▓█▀█▓▓_▓▓███▓▓▓▓▓  ▓▓▓▓▓Ω4                   
                µ ├▓▀▀L   └ⁿ  ▀   ▀ ▓▓█w ▓▓▓▀ìⁿ                    
                ⌐ ├_                τ▀▓  Σ⌐└                       
                ~ ├▓▓  ▄  _     ╒  ┌▄▓▓  Γ                         
                  ├▓▓▓▌█═┴▓▄╒▀▄_▄▌═¢▓▓▓  ╚                         
               ⌠  ├▓▓▓▓▓ⁿ▄▓▓▓▓▓▓▓┐▄▓▓▓▓  └                         
               Ω_.└██▓▀ⁿÇⁿ▀▀▀▀▀▀█≡▀▀▀▀▀   µ                        
               ⁿ  .▄▄▓▓▓▓▄▄┌ ╖__▓_▄▄▄▄▄*Oⁿ                         
                 û▌├▓█▓▓▓██ⁿ ¡▓▓▓▓▓▓▓▓█▓╪                          
                 ╙Ω▀█ ▓██ⁿ    └█▀██▀▓█├█Å                          
                     ⁿⁿ             ⁿ ⁿ^                           
:::::::..  .,::::::    .,::      .::::::::..    :::.  .-:.     ::-.
;;;;'';;;; ;;;;''''    ';;;,  .,;; ;;;;'';;;;   ;;';;  ';;.   ;;;;'
 [[[,/[[['  [[cccc       '[[,,[['   [[[,/[[['  ,[[ '[[,  '[[,[[['  
 $$$$$$c    $$""""        Y$$$Pcccc $$$$$$c   c$$$cc$$$c   c$$"    
 888b "88bo,888oo,__    oP"''"Yo,   888b "88bo,888   888,,8P"'     
 MMMM   "W" """"YUMMM,m"       "Mm, MMMM   "W" YMM   ""'mM"        
`

const UsageTemplate = `{{ $cmd := . }}
Usage: {{if .Runnable}}
  {{.UseLine}}{{if .HasFlags}} [flags]{{end}}{{end}}{{if .HasSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasNonHelpSubCommands}}

Available Commands: {{range .Commands}}{{if (not .IsHelpCommand)}}` +
	`{{if ne .Use "start"}}{{if ne .Use "stop"}}` +
	`{{if ne .Use "restart"}}{{if ne .Use "status"}}` +
	"\n  {{rpad .Name .NamePadding }} {{.Short}}" +
	`{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}{{ if .HasLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages}}{{end}}{{ if .HasInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics: {{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}}{{end}}{{end}}{{ if .HasSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`
