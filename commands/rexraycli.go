package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emccode/clue"
	"github.com/emccode/rexray"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v1"
)

var (
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
		cmd.Usage()
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\nRexray Version: %v\n", "0.1.150418")
	},
}

var getinstanceCmd = &cobra.Command{
	Use: "get-instance",
	Run: func(cmd *cobra.Command, args []string) {

		allInstances, err := rexray.GetInstance()
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

var getblockdeviceCmd = &cobra.Command{
	Use: "get-blockdevice",
	Run: func(cmd *cobra.Command, args []string) {

		allBlockDevices, err := rexray.GetBlockDeviceMapping()
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

		allVolumes, err := rexray.GetVolume(volumeID, volumeName)
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

		allSnapshots, err := rexray.GetSnapshot(volumeID, snapshotID, snapshotName)
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

		snapshot, err := rexray.CreateSnapshot(runAsync, volumeName, volumeID, description)
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

		err := rexray.RemoveSnapshot(snapshotID)
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

		volume, err := rexray.CreateVolume(runAsync, volumeName, volumeID, snapshotID, volumeType, IOPS, size, availabilityZone)
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

		err := rexray.RemoveVolume(volumeID)
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

		volumeAttachment, err := rexray.AttachVolume(runAsync, volumeID, instanceID)
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

		err := rexray.DetachVolume(runAsync, volumeID, instanceID)
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

		snapshot, err := rexray.CopySnapshot(runAsync, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion)
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

//Exec function
func Exec() {
	AddCommands()
	RexrayCmd.Execute()
}

//AddCommands function
func AddCommands() {
	RexrayCmd.AddCommand(versionCmd)
	RexrayCmd.AddCommand(getinstanceCmd)
	RexrayCmd.AddCommand(getblockdeviceCmd)
	RexrayCmd.AddCommand(getvolumeCmd)
	RexrayCmd.AddCommand(getsnapshotCmd)
	RexrayCmd.AddCommand(newsnapshotCmd)
	RexrayCmd.AddCommand(removesnapshotCmd)
	RexrayCmd.AddCommand(newvolumeCmd)
	RexrayCmd.AddCommand(removevolumeCmd)
	RexrayCmd.AddCommand(attachvolumeCmd)
	RexrayCmd.AddCommand(detachvolumeCmd)
	RexrayCmd.AddCommand(copysnapshotCmd)
}

var rexrayCmdV *cobra.Command

func init() {
	RexrayCmd.PersistentFlags().StringVar(&cfgFile, "Config", "", "config file (default is $HOME/rexray/config.yaml)")
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
