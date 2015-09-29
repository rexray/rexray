package cli

import (
	"fmt"
	"os"

	_ "github.com/emccode/rexray/imports"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emccode/rexray/config"
	osm "github.com/emccode/rexray/os"
	"github.com/emccode/rexray/rexray/cli/term"
	"github.com/emccode/rexray/storage"
	"github.com/emccode/rexray/util"
	"github.com/emccode/rexray/volume"
)

const (
	NoColor     = 0
	Black       = 30
	Red         = 31
	RedBg       = 41
	Green       = 32
	Yellow      = 33
	Blue        = 34
	Gray        = 37
	BlueBg      = Blue + 10
	White       = 97
	WhiteBg     = White + 10
	DarkGrayBg  = 100
	LightBlue   = 94
	LightBlueBg = LightBlue + 10
)

var (
	c *config.Config

	sdm  *storage.StorageDriverManager
	vdm  *volume.VolumeDriverManager
	osdm *osm.OSDriverManager

	client                  string
	fg                      bool
	force                   bool
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
	moduleConfig            []string
)

type HelpFlagPanic struct{}
type PrintedErrorPanic struct{}
type SubCommandPanic struct{}

//Exec function
func Exec() {
	defer func() {
		r := recover()
		if r != nil {
			switch r.(type) {
			case HelpFlagPanic, SubCommandPanic:
			// Do nothing
			case PrintedErrorPanic:
				os.Exit(1)
			default:
				panic(r)
			}
		}
	}()

	RexrayCmd.Execute()
}

func init() {
	c = config.New()
	updateLogLevel()
	initCommands()
	initFlags()
	initUsageTemplates()
}

func updateLogLevel() {
	switch c.LogLevel {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	}

	log.WithField("logLevel", c.LogLevel).Debug("updated log level")
}

func preRun(cmd *cobra.Command, args []string) {

	if cfgFile != "" && util.FileExists(cfgFile) {
		if err := c.ReadConfigFile(cfgFile); err != nil {
			panic(err)
		}
		cmd.Flags().Parse(os.Args[1:])
	}

	updateLogLevel()

	if isHelpFlag(cmd) {
		cmd.Help()
		panic(&HelpFlagPanic{})
	}

	if permErr := checkCmdPermRequirements(cmd); permErr != nil {
		if term.IsTerminal() {
			printColorizedError(permErr)
		} else {
			printNonColorizedError(permErr)
		}

		fmt.Println()
		cmd.Help()
		panic(&PrintedErrorPanic{})
	}

	if isInitDriverManagersCmd(cmd) {
		if initDmErr := initDriverManagers(); initDmErr != nil {

			if term.IsTerminal() {
				printColorizedError(initDmErr)
			} else {
				printNonColorizedError(initDmErr)
			}
			fmt.Println()

			helpCmd := cmd
			if cmd == volumeCmd {
				helpCmd = volumeGetCmd
			} else if cmd == snapshotCmd {
				helpCmd = snapshotGetCmd
			} else if cmd == deviceCmd {
				helpCmd = deviceGetCmd
			} else if cmd == adapterCmd {
				helpCmd = adapterGetTypesCmd
			}
			helpCmd.Help()

			panic(&PrintedErrorPanic{})
		}
	}
}

func isHelpFlags(cmd *cobra.Command) bool {
	help, _ := cmd.Flags().GetBool("help")
	verb, _ := cmd.Flags().GetBool("verbose")
	return help || verb
}

func checkCmdPermRequirements(cmd *cobra.Command) error {
	if cmd == installCmd {
		return checkOpPerms("installed")
	}

	if cmd == uninstallCmd {
		return checkOpPerms("uninstalled")
	}

	if cmd == serviceStartCmd {
		return checkOpPerms("started")
	}

	if cmd == serviceStopCmd {
		return checkOpPerms("stopped")
	}

	if cmd == serviceRestartCmd {
		return checkOpPerms("restarted")
	}

	return nil
}

func printColorizedError(err error) {
	stderr := os.Stderr
	l := fmt.Sprintf("\x1b[%dm\xe2\x86\x93\x1b[0m", White)

	fmt.Fprintf(stderr, "Oops, an \x1b[%[1]dmerror\x1b[0m occured!\n\n", RedBg)
	fmt.Fprintf(stderr, "  \x1b[%dm%s\n\n", Red, err.Error())
	fmt.Fprintf(stderr, "\x1b[0m")
	fmt.Fprintf(stderr,
		"To correct the \x1b[%dmerror\x1b[0m please review:\n\n", RedBg)
	fmt.Fprintf(
		stderr,
		"  - Debug output by using the flag \x1b[%dm-l debug\x1b[0m\n",
		LightBlue)
	fmt.Fprintf(stderr, "  - The REX-ray website at \x1b[%dm%s\x1b[0m\n",
		BlueBg, "https://github.com/emccode/rexray")
	fmt.Fprintf(stderr, "  - The on%[1]sine he%[1]sp be%[1]sow\n", l)
}

func printNonColorizedError(err error) {
	stderr := os.Stderr

	fmt.Fprintf(stderr, "Oops, an error occured!\n\n")
	fmt.Fprintf(stderr, "  %s\n", err.Error())
	fmt.Fprintf(stderr, "To correct the error please review:\n\n")
	fmt.Fprintf(stderr, "  - Debug output by using the flag \"-l debug\"\n")
	fmt.Fprintf(
		stderr,
		"  - The REX-ray website at https://github.com/emccode/rexray\n")
	fmt.Fprintf(stderr, "  - The online help below\n")
}

func isInitDriverManagersCmd(cmd *cobra.Command) bool {

	return cmd.Parent() != nil &&
		cmd != adapterCmd &&
		cmd != adapterGetTypesCmd &&
		cmd != versionCmd &&
		cmd != serviceCmd &&
		cmd != serviceInitSysCmd &&
		cmd != installCmd &&
		cmd != uninstallCmd &&
		cmd != serviceStatusCmd &&
		cmd != serviceStopCmd &&
		!(cmd == serviceStartCmd && (client != "" || fg || force)) &&
		cmd != moduleCmd &&
		cmd != moduleTypesCmd &&
		cmd != moduleInstancesCmd &&
		cmd != moduleInstancesListCmd
}

func initDriverManagers() error {

	var osdmErr error
	osdm, osdmErr = osm.NewOSDriverManager(c)
	if osdmErr != nil {
		return osdmErr
	}

	var sdmErr error
	sdm, sdmErr = storage.NewStorageDriverManager(c)
	if sdmErr != nil {
		return sdmErr
	}

	var vdmErr error
	vdm, vdmErr = volume.NewVolumeDriverManager(c, osdm, sdm)
	if vdmErr != nil {
		return vdmErr
	}

	return nil
}
