package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emccode/clue"
	_ "github.com/emccode/rexray"
	rrdaemon "github.com/emccode/rexray/daemon"
	rros "github.com/emccode/rexray/os"
	"github.com/emccode/rexray/storage"
	"github.com/emccode/rexray/volume"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v1"
)

var (
	daemon                  bool
	host                    string
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
)

//FlagValue struct
type FlagValue struct {
	value      *string
	mandatory  bool
	persistent bool
	overrideby string
}

//RexrayCmd
var RexrayCmd = &cobra.Command{
	Use: "rexray",
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()

		if !daemon {
			cmd.Usage()
			return
		}

		if err := rrdaemon.Start(host); err != nil {
			log.Fatalf("Error starting daemon: %s", err)
		}

	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\nRexray Version: %v\n", "0.1.150429")
	},
}

var getinstanceCmd = &cobra.Command{
	Use: "get-instance",
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

var getvolumemapCmd = &cobra.Command{
	Use: "get-volumemap",
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

var getvolumeCmd = &cobra.Command{
	Use: "get-volume",
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

var getsnapshotCmd = &cobra.Command{
	Use: "get-snapshot",
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

var newsnapshotCmd = &cobra.Command{
	Use: "new-snapshot",
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

var removesnapshotCmd = &cobra.Command{
	Use: "remove-snapshot",
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

var newvolumeCmd = &cobra.Command{
	Use: "new-volume",
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

var removevolumeCmd = &cobra.Command{
	Use: "remove-volume",
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

var attachvolumeCmd = &cobra.Command{
	Use: "attach-volume",
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

var detachvolumeCmd = &cobra.Command{
	Use: "detach-volume",
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

var copysnapshotCmd = &cobra.Command{
	Use: "copy-snapshot",
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

var getmountCmd = &cobra.Command{
	Use: "get-mount",
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

var mountdeviceCmd = &cobra.Command{
	Use: "mount-device",
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

var unmountdeviceCmd = &cobra.Command{
	Use: "unmount-device",
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

var formatdeviceCmd = &cobra.Command{
	Use: "format-device",
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

var mountvolumeCmd = &cobra.Command{
	Use: "mount-volume",
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

var unmountvolumeCmd = &cobra.Command{
	Use: "unmount-volume",
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

var getvolumepathCmd = &cobra.Command{
	Use: "get-volumepath",
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
	AddCommands()
	RexrayCmd.Execute()
}

//AddCommands function
func AddCommands() {
	RexrayCmd.AddCommand(versionCmd)
	RexrayCmd.AddCommand(getinstanceCmd)
	RexrayCmd.AddCommand(getvolumemapCmd)
	RexrayCmd.AddCommand(getvolumeCmd)
	RexrayCmd.AddCommand(getsnapshotCmd)
	RexrayCmd.AddCommand(newsnapshotCmd)
	RexrayCmd.AddCommand(removesnapshotCmd)
	RexrayCmd.AddCommand(newvolumeCmd)
	RexrayCmd.AddCommand(removevolumeCmd)
	RexrayCmd.AddCommand(attachvolumeCmd)
	RexrayCmd.AddCommand(detachvolumeCmd)
	RexrayCmd.AddCommand(copysnapshotCmd)
	RexrayCmd.AddCommand(getmountCmd)
	RexrayCmd.AddCommand(mountdeviceCmd)
	RexrayCmd.AddCommand(unmountdeviceCmd)
	RexrayCmd.AddCommand(formatdeviceCmd)
	RexrayCmd.AddCommand(mountvolumeCmd)
	RexrayCmd.AddCommand(unmountvolumeCmd)
	RexrayCmd.AddCommand(getvolumepathCmd)
}

var rexrayCmdV *cobra.Command

func init() {
	RexrayCmd.PersistentFlags().StringVar(&cfgFile, "Config", "", "config file (default is $HOME/rexray/config.yaml)")
	RexrayCmd.Flags().BoolVar(&daemon, "daemon", false, "Daemonize and establish a socket file")
	RexrayCmd.Flags().StringVar(&host, "host", "", "Socket to connect to")
	getvolumeCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	getvolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	getsnapshotCmd.Flags().StringVar(&snapshotName, "snapshotname", "", "snapshotname")
	getsnapshotCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	getsnapshotCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	newsnapshotCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	newsnapshotCmd.Flags().StringVar(&snapshotName, "snapshotname", "", "snapshotname")
	newsnapshotCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	newsnapshotCmd.Flags().StringVar(&description, "description", "", "description")
	removesnapshotCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	newvolumeCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	newvolumeCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	newvolumeCmd.Flags().StringVar(&volumeType, "volumetype", "", "volumetype")
	newvolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	newvolumeCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	newvolumeCmd.Flags().Int64Var(&IOPS, "iops", 0, "IOPS")
	newvolumeCmd.Flags().Int64Var(&size, "size", 0, "size")
	newvolumeCmd.Flags().StringVar(&availabilityZone, "availabilityzone", "", "availabilityzone")
	removevolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	attachvolumeCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	attachvolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	attachvolumeCmd.Flags().StringVar(&instanceID, "instanceid", "", "instanceid")
	detachvolumeCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	detachvolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	detachvolumeCmd.Flags().StringVar(&instanceID, "instanceid", "", "instanceid")
	copysnapshotCmd.Flags().BoolVar(&runAsync, "runasync", false, "runasync")
	copysnapshotCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	copysnapshotCmd.Flags().StringVar(&snapshotID, "snapshotid", "", "snapshotid")
	copysnapshotCmd.Flags().StringVar(&snapshotName, "snapshotname", "", "snapshotname")
	copysnapshotCmd.Flags().StringVar(&destinationSnapshotName, "destinationsnapshotname", "", "destinationsnapshotname")
	copysnapshotCmd.Flags().StringVar(&destinationRegion, "destinationregion", "", "destinationregion")
	getmountCmd.Flags().StringVar(&deviceName, "devicename", "", "devicename")
	getmountCmd.Flags().StringVar(&mountPoint, "mountpoint", "", "mountpoint")
	mountdeviceCmd.Flags().StringVar(&deviceName, "devicename", "", "devicename")
	mountdeviceCmd.Flags().StringVar(&mountPoint, "mountpoint", "", "mountpoint")
	mountdeviceCmd.Flags().StringVar(&mountOptions, "mountoptions", "", "mountoptions")
	mountdeviceCmd.Flags().StringVar(&mountLabel, "mountlabel", "", "mountlabel")
	unmountdeviceCmd.Flags().StringVar(&mountPoint, "mountpoint", "", "mountpoint")
	formatdeviceCmd.Flags().StringVar(&deviceName, "devicename", "", "devicename")
	formatdeviceCmd.Flags().StringVar(&fsType, "fstype", "", "fstype")
	formatdeviceCmd.Flags().BoolVar(&overwriteFs, "overwritefs", false, "overwritefs")
	mountvolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	mountvolumeCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	mountvolumeCmd.Flags().BoolVar(&overwriteFs, "overwritefs", false, "overwritefs")
	mountvolumeCmd.Flags().StringVar(&fsType, "fstype", "", "fstype")
	unmountvolumeCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	unmountvolumeCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")
	getvolumepathCmd.Flags().StringVar(&volumeID, "volumeid", "", "volumeid")
	getvolumepathCmd.Flags().StringVar(&volumeName, "volumename", "", "volumename")

	rexrayCmdV = RexrayCmd

	// initConfig(systemCmd, "rexray", true, map[string]FlagValue{
	// 	"endpoint": {&endpoint, false, false, ""},
	// 	"insecure": {&insecure, false, false, ""},
	// })

}

func initConfig(cmd *cobra.Command, suffix string, checkValues bool, flags map[string]FlagValue) {
	InitConfig()

	defaultFlags := map[string]FlagValue{}
	if len(flags) == 0 {
		defaultFlags = map[string]FlagValue{
		// "username": {&username, true, false, ""},
		// "password": {&password, true, false, ""},
		// "endpoint": {&endpoint, true, false, ""},
		// "insecure": {&insecure, false, false, ""},
		}
	}

	for key, field := range flags {
		defaultFlags[key] = field
	}

	var fieldsMissing []string
	var fieldsMissingRemove []string

	type statusFlag struct {
		key                        string
		flagValue                  string
		flagValueExists            bool
		flagChanged                bool
		keyOverrideBy              string
		flagValueOverrideBy        string
		flagValueOverrideByExists  bool
		flagChangedOverrideBy      bool
		viperValue                 string
		viperValueExists           bool
		viperValueOverrideBy       string
		viperValueOverrideByExists bool
		gobValue                   string
		gobValueExists             bool
		finalViperValue            string
		setFrom                    string
	}

	cmdFlags := &pflag.FlagSet{}
	var statusFlags []statusFlag

	for key, field := range defaultFlags {
		viper.BindEnv(key)

		switch field.persistent {
		case true:
			cmdFlags = cmd.PersistentFlags()
		case false:
			cmdFlags = cmd.Flags()
		default:
		}

		var flagLookupValue string
		var flagLookupChanged bool

		if cmdFlags.Lookup(key) != nil {
			flagLookupValue = cmdFlags.Lookup(key).Value.String()
			flagLookupChanged = cmdFlags.Lookup(key).Changed
		}

		statusFlag := &statusFlag{
			key:                  key,
			flagValue:            flagLookupValue,
			flagChanged:          flagLookupChanged,
			viperValue:           viper.GetString(key),
			viperValueOverrideBy: viper.GetString(field.overrideby),
		}

		if statusFlag.flagValue != "" {
			statusFlag.flagValueExists = true
		}
		if statusFlag.flagValueOverrideBy != "" {
			statusFlag.flagValueOverrideByExists = true
		}
		if statusFlag.viperValue != "" {
			statusFlag.viperValueExists = true
		}
		if statusFlag.viperValueOverrideBy != "" {
			statusFlag.viperValueOverrideByExists = true
		}

		if field.overrideby != "" {
			statusFlag.keyOverrideBy = field.overrideby
			if cmdFlags.Lookup(field.overrideby) != nil {
				statusFlag.flagChangedOverrideBy = cmdFlags.Lookup(field.overrideby).Changed
				statusFlag.flagValueOverrideBy = cmdFlags.Lookup(field.overrideby).Value.String()
			}
		}

		statusFlags = append(statusFlags, *statusFlag)

		if err := setGobValues(cmd, suffix, key); err != nil {
			log.Fatal(err)
		}
	}

	for i := range statusFlags {
		statusFlags[i].setFrom = "none"
		statusFlags[i].gobValue = viper.GetString(statusFlags[i].key)
		if statusFlags[i].gobValue != "" {
			statusFlags[i].gobValueExists = true
			statusFlags[i].setFrom = "gob"
		}

		if statusFlags[i].gobValue == statusFlags[i].viperValue {
			if statusFlags[i].gobValueExists {
				statusFlags[i].setFrom = "ConfigOrEnv"
			} else {
				statusFlags[i].setFrom = "none"
			}
		}

		if statusFlags[i].flagValueOverrideByExists {
			viper.Set(statusFlags[i].key, "")
			statusFlags[i].setFrom = "flagValueOverrideByExists"
			continue
		}
		if statusFlags[i].flagValueExists {
			viper.Set(statusFlags[i].key, statusFlags[i].flagValue)
			statusFlags[i].setFrom = "flagValueExists"
			continue
		}
		if statusFlags[i].viperValueOverrideByExists {
			viper.Set(statusFlags[i].key, "")
			statusFlags[i].setFrom = "viperValueOverrideByExists"
			continue
		}

	}

	for _, statusFlag := range statusFlags {
		statusFlag.finalViperValue = viper.GetString(statusFlag.key)
		if os.Getenv("REXRAY_SHOW_FLAG") == "true" {
			fmt.Printf("%+v\n", statusFlag)
		}
	}

	if checkValues {
		for key, field := range defaultFlags {
			if field.mandatory == true {
				if viper.GetString(key) != "" && (field.overrideby != "" && viper.GetString(field.overrideby) == "") {
					fieldsMissingRemove = append(fieldsMissingRemove, field.overrideby)
				} else {
					//if viper.GetString(key) == "" && (field.overrideby != "" && viper.GetString(field.overrideby) == "") {
					if viper.GetString(key) == "" {
						fieldsMissing = append(fieldsMissing, key)
					}
				}
			}
		}

		for _, fieldMissingRemove := range fieldsMissingRemove {
		Loop1:
			for i, fieldMissing := range fieldsMissing {
				if fieldMissing == fieldMissingRemove {
					fieldsMissing = append(fieldsMissing[:i], fieldsMissing[i+1:]...)
					break Loop1
				}
			}
		}

		if len(fieldsMissing) != 0 {
			log.Fatalf("missing parameter: %v", strings.Join(fieldsMissing, ", "))
		}
	}

	for key := range defaultFlags {
		if viper.GetString(key) != "" {
			os.Setenv(fmt.Sprintf("REXRAY_%v", strings.ToUpper(key)), viper.GetString(key))
		}
		//fmt.Println(viper.GetString(key))
	}

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

func setGobValues(cmd *cobra.Command, suffix string, field string) (err error) {
	getValue := clue.GetValue{}
	if err := clue.DecodeGobFile(suffix, &getValue); err != nil {
		return fmt.Errorf("Problem with decodeGobFile: %v", err)
	}

	if os.Getenv("REXRAY_SHOW_GOB") == "true" {
		for key, value := range getValue.VarMap {
			fmt.Printf("%v: %v\n", key, *value)
		}
		fmt.Printf("%+v\n", getValue.VarMap)
		fmt.Println()
	}

	for key := range getValue.VarMap {
		lowerKey := strings.ToLower(key)
		if field != "" && field != lowerKey {
			continue
		}

		viper.Set(lowerKey, *getValue.VarMap[key])
	}
	return
}
