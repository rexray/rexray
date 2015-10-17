package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emccode/rexray"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/rexray/cli/term"
	"github.com/emccode/rexray/util"
)

const (
	noColor     = 0
	black       = 30
	red         = 31
	redBg       = 41
	green       = 32
	yellow      = 33
	blue        = 34
	gray        = 37
	blueBg      = blue + 10
	white       = 97
	whiteBg     = white + 10
	darkGrayBg  = 100
	lightBlue   = 94
	lightBlueBg = lightBlue + 10
)

var (
	r *core.RexRay

	client                  string
	fg                      bool
	force                   bool
	cfgFile                 string
	snapshotID              string
	volumeID                string
	runAsync                bool
	description             string
	volumeType              string
	iops                    int64
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
	moduleTypeID            int32
	moduleInstanceID        int32
	moduleInstanceAddress   string
	moduleInstanceStart     bool
	moduleConfig            []string
)

type helpFlagPanic struct{}
type printedErrorPanic struct{}
type subCommandPanic struct{}

//Exec function
func Exec() {
	defer func() {
		r := recover()
		if r != nil {
			switch r.(type) {
			case helpFlagPanic, subCommandPanic:
			// Do nothing
			case printedErrorPanic:
				os.Exit(1)
			default:
				panic(r)
			}
		}
	}()

	RexrayCmd.Execute()
}

func init() {
	var err error
	if r, err = rexray.New(); err != nil {
		panic(err)
	}
	updateLogLevel()
	initCommands()
	initFlags()
	initUsageTemplates()
}

func updateLogLevel() {
	switch r.Config.LogLevel {
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

	log.WithField("logLevel", r.Config.LogLevel).Debug("updated log level")
}

func preRun(cmd *cobra.Command, args []string) {

	if cfgFile != "" && util.FileExists(cfgFile) {
		if err := r.Config.ReadConfigFile(cfgFile); err != nil {
			panic(err)
		}
		cmd.Flags().Parse(os.Args[1:])
	}

	updateLogLevel()

	if isHelpFlag(cmd) {
		cmd.Help()
		panic(&helpFlagPanic{})
	}

	if permErr := checkCmdPermRequirements(cmd); permErr != nil {
		if term.IsTerminal() {
			printColorizedError(permErr)
		} else {
			printNonColorizedError(permErr)
		}

		fmt.Println()
		cmd.Help()
		panic(&printedErrorPanic{})
	}

	if isInitDriverManagersCmd(cmd) {
		if err := r.InitDrivers(); err != nil {

			if term.IsTerminal() {
				printColorizedError(err)
			} else {
				printNonColorizedError(err)
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

			panic(&printedErrorPanic{})
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
	l := fmt.Sprintf("\x1b[%dm\xe2\x86\x93\x1b[0m", white)

	fmt.Fprintf(stderr, "Oops, an \x1b[%[1]dmerror\x1b[0m occured!\n\n", redBg)
	fmt.Fprintf(stderr, "  \x1b[%dm%s\n\n", red, err.Error())
	fmt.Fprintf(stderr, "\x1b[0m")
	fmt.Fprintf(stderr,
		"To correct the \x1b[%dmerror\x1b[0m please review:\n\n", redBg)
	fmt.Fprintf(
		stderr,
		"  - Debug output by using the flag \x1b[%dm-l debug\x1b[0m\n",
		lightBlue)
	fmt.Fprintf(stderr, "  - The REX-ray website at \x1b[%dm%s\x1b[0m\n",
		blueBg, "https://github.com/emccode/rexray")
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
