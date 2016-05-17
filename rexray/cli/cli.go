package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	glog "github.com/akutz/golf/logrus"
	"github.com/akutz/gotil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v1"

	"github.com/emccode/libstorage"
	"github.com/emccode/libstorage/api/context"
	apitypes "github.com/emccode/libstorage/api/types"
	apiutils "github.com/emccode/libstorage/api/utils"

	"github.com/emccode/rexray/rexray/cli/term"
	"github.com/emccode/rexray/util"
)

func init() {
	log.SetFormatter(&glog.TextFormatter{TextFormatter: log.TextFormatter{}})
}

type helpFlagPanic struct{}
type printedErrorPanic struct{}
type subCommandPanic struct{}

// CLI is the REX-Ray command line interface.
type CLI struct {
	l      *log.Logger
	r      apitypes.Client
	rs     io.Closer
	c      *cobra.Command
	config gofig.Config
	ctx    apitypes.Context

	serviceCmd               *cobra.Command
	moduleCmd                *cobra.Command
	versionCmd               *cobra.Command
	envCmd                   *cobra.Command
	volumeCmd                *cobra.Command
	snapshotCmd              *cobra.Command
	deviceCmd                *cobra.Command
	moduleTypesCmd           *cobra.Command
	moduleInstancesCmd       *cobra.Command
	moduleInstancesListCmd   *cobra.Command
	moduleInstancesCreateCmd *cobra.Command
	moduleInstancesStartCmd  *cobra.Command
	installCmd               *cobra.Command
	uninstallCmd             *cobra.Command
	serviceStartCmd          *cobra.Command
	serviceRestartCmd        *cobra.Command
	serviceStopCmd           *cobra.Command
	serviceStatusCmd         *cobra.Command
	serviceInitSysCmd        *cobra.Command
	adapterCmd               *cobra.Command
	adapterGetTypesCmd       *cobra.Command
	adapterGetInstancesCmd   *cobra.Command
	volumeMapCmd             *cobra.Command
	volumeGetCmd             *cobra.Command
	snapshotGetCmd           *cobra.Command
	snapshotCreateCmd        *cobra.Command
	snapshotRemoveCmd        *cobra.Command
	volumeCreateCmd          *cobra.Command
	volumeRemoveCmd          *cobra.Command
	volumeAttachCmd          *cobra.Command
	volumeDetachCmd          *cobra.Command
	snapshotCopyCmd          *cobra.Command
	deviceGetCmd             *cobra.Command
	deviceMountCmd           *cobra.Command
	devuceUnmountCmd         *cobra.Command
	deviceFormatCmd          *cobra.Command
	volumeMountCmd           *cobra.Command
	volumeUnmountCmd         *cobra.Command
	volumePathCmd            *cobra.Command

	service                 string
	outputFormat            string
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
	moduleTypeName          string
	moduleInstanceName      string
	moduleInstanceAddress   string
	moduleInstanceStart     bool
	moduleConfig            []string
}

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

func validateConfig(path string) {
	if !gotil.FileExists(path) {
		return
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "rexray: error reading config: %s\n%v\n", path, err)
		os.Exit(1)
	}

	s := string(buf)

	if _, err := gofig.ValidateYAMLString(s); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"rexray: invalid config: %s\n\n  %v\n\n", path, err)
		fmt.Fprint(
			os.Stderr,
			"paste the contents between ---BEGIN--- and ---END---\n")
		fmt.Fprint(
			os.Stderr,
			"into http://www.yamllint.com/ to discover the issue\n\n")
		fmt.Fprintln(os.Stderr, "---BEGIN---")
		fmt.Fprintln(os.Stderr, s)
		fmt.Fprintln(os.Stderr, "---END---")
		os.Exit(1)
	}
}

// New returns a new CLI using the current process's arguments.
func New() *CLI {
	return NewWithArgs(os.Args[1:]...)
}

// NewWithArgs returns a new CLI using the specified arguments.
func NewWithArgs(a ...string) *CLI {

	validateConfig(util.EtcFilePath("config.yml"))
	validateConfig(fmt.Sprintf("%s/.rexray/config.yml", gotil.HomeDir()))

	s := "REX-Ray:\n" +
		"  A guest-based storage introspection tool that enables local\n" +
		"  visibility and management from cloud and storage platforms."

	c := &CLI{
		l:      log.New(),
		config: gofig.New(),
	}

	c.c = &cobra.Command{
		Use:              "rexray",
		Short:            s,
		PersistentPreRun: c.preRun,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	c.c.SetArgs(a)

	c.initOtherCmdsAndFlags()

	c.initAdapterCmdsAndFlags()
	c.initDeviceCmdsAndFlags()
	c.initVolumeCmdsAndFlags()
	// c.initSnapshotCmdsAndFlags()

	c.initServiceCmdsAndFlags()
	c.initModuleCmdsAndFlags()

	c.initUsageTemplates()

	return c
}

// Execute executes the CLI using the current process's arguments.
func Execute() {
	New().Execute()
}

// ExecuteWithArgs executes the CLI using the specified arguments.
func ExecuteWithArgs(a ...string) {
	NewWithArgs(a...).Execute()
}

// Execute executes the CLI.
func (c *CLI) Execute() {
	defer func() {
		r := recover()
		if c.rs != nil {
			c.rs.Close()
		}
		switch r := r.(type) {
		case nil:
			return
		case int:
			log.Debugf("exiting with error code %d", r)
			os.Exit(r)
		case error:
			log.Panic(r)
		default:
			log.Debugf("exiting with default error code 1, r=%v", r)
			os.Exit(1)
		}
	}()
	c.execute()
	if c.rs != nil {
		c.rs.Close()
	}
}

func (c *CLI) execute() {
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
	c.c.Execute()
}

func (c *CLI) marshalOutput(v interface{}) (string, error) {
	var err error
	var buf []byte
	if strings.ToUpper(c.outputFormat) == "JSON" {
		buf, err = marshalJSONOutput(v)
	} else {
		buf, err = marshalYamlOutput(v)
	}
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func marshalYamlOutput(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func marshalJSONOutput(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *CLI) addOutputFormatFlag(fs *pflag.FlagSet) {
	fs.StringVarP(
		&c.outputFormat, "format", "f", "yml", "The output format (yml, json)")
}

func (c *CLI) updateLogLevel() {
	switch c.logLevel() {
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

	log.WithField("logLevel", c.logLevel()).Debug("updated log level")
}

func (c *CLI) preRun(cmd *cobra.Command, args []string) {

	if c.cfgFile != "" && gotil.FileExists(c.cfgFile) {
		validateConfig(c.cfgFile)
		if err := c.config.ReadConfigFile(c.cfgFile); err != nil {
			panic(err)
		}
		os.Setenv("REXRAY_CONFIG_FILE", c.cfgFile)
		cmd.Flags().Parse(os.Args[1:])
	}

	c.updateLogLevel()

	if !c.config.IsSet(
		apitypes.ConfigIgVolOpsMountPath) {
		c.config.Set(
			apitypes.ConfigIgVolOpsMountPath,
			util.LibFilePath("volumes"))
	}

	if !c.config.IsSet(apitypes.ConfigIgVolOpsPathCache) {
		c.config.Set(
			apitypes.ConfigIgVolOpsPathCache,
			false)
	}

	c.config = c.config.Scope("rexray.modules.default-docker")

	if isHelpFlag(cmd) {
		cmd.Help()
		panic(&helpFlagPanic{})
	}

	if permErr := c.checkCmdPermRequirements(cmd); permErr != nil {
		if term.IsTerminal() {
			printColorizedError(permErr)
		} else {
			printNonColorizedError(permErr)
		}

		fmt.Println()
		cmd.Help()
		panic(&printedErrorPanic{})
	}

	if c.isInitDriverManagersCmd(cmd) {
		c.ctx = context.Background()
		// if c.service != "" && !c.config.IsSet(apitypes.ConfigService) {
		// 	c.ctx = c.ctx.WithServiceName(c.service)
		// }
		if c.runAsync {
			c.ctx = c.ctx.WithValue("async", true)
		}

		r, rs, err, _ := libstorage.New(c.config.Scope("rexray"))
		if err == nil {
			c.r = r
			c.rs = rs
		} else {
			if term.IsTerminal() {
				printColorizedError(err)
			} else {
				printNonColorizedError(err)
			}
			fmt.Println()

			helpCmd := cmd
			if cmd == c.volumeCmd {
				helpCmd = c.volumeGetCmd
			} else if cmd == c.snapshotCmd {
				helpCmd = c.snapshotGetCmd
			} else if cmd == c.deviceCmd {
				helpCmd = c.deviceGetCmd
			} else if cmd == c.adapterCmd {
				helpCmd = c.adapterGetTypesCmd
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

func (c *CLI) checkCmdPermRequirements(cmd *cobra.Command) error {
	if cmd == c.installCmd {
		return checkOpPerms("installed")
	}

	if cmd == c.uninstallCmd {
		return checkOpPerms("uninstalled")
	}

	if cmd == c.serviceStartCmd {
		return checkOpPerms("started")
	}

	if cmd == c.serviceStopCmd {
		return checkOpPerms("stopped")
	}

	if cmd == c.serviceRestartCmd {
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

func (c *CLI) isInitDriverManagersCmd(cmd *cobra.Command) bool {
	return cmd.Parent() != nil &&
		cmd != c.versionCmd &&
		cmd != c.envCmd &&
		c.isServiceCmd(cmd) &&
		c.isModuleCmd(cmd)
}

func (c *CLI) isServiceCmd(cmd *cobra.Command) bool {
	return cmd != c.serviceCmd &&
		cmd != c.serviceInitSysCmd &&
		cmd != c.installCmd &&
		cmd != c.uninstallCmd &&
		cmd != c.serviceStatusCmd &&
		cmd != c.serviceStopCmd &&
		!(cmd == c.serviceStartCmd && (c.client != "" || c.fg || c.force))
}

func (c *CLI) isModuleCmd(cmd *cobra.Command) bool {
	return cmd != c.moduleCmd &&
		cmd != c.moduleTypesCmd &&
		cmd != c.moduleInstancesCmd &&
		cmd != c.moduleInstancesListCmd
}

func (c *CLI) logLevel() string {
	return c.config.GetString("rexray.logLevel")
}

func (c *CLI) host() string {
	return c.config.GetString("rexray.host")
}

func store() apitypes.Store {
	return apiutils.NewStore()
}
