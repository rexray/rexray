package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	glog "github.com/akutz/golf/logrus"
	"github.com/akutz/gotil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/emccode/libstorage/api/context"
	apiserver "github.com/emccode/libstorage/api/server"
	apitypes "github.com/emccode/libstorage/api/types"
	apiutils "github.com/emccode/libstorage/api/utils"
	apiclient "github.com/emccode/libstorage/client"

	"github.com/emccode/rexray/cli/cli/term"
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
	l                  *log.Logger
	r                  apitypes.Client
	rs                 apitypes.Server
	rsErrs             <-chan error
	c                  *cobra.Command
	config             gofig.Config
	ctx                apitypes.Context
	activateLibStorage bool

	envCmd     *cobra.Command
	versionCmd *cobra.Command

	installCmd   *cobra.Command
	uninstallCmd *cobra.Command

	moduleCmd                *cobra.Command
	moduleTypesCmd           *cobra.Command
	moduleInstancesCmd       *cobra.Command
	moduleInstancesListCmd   *cobra.Command
	moduleInstancesCreateCmd *cobra.Command
	moduleInstancesStartCmd  *cobra.Command

	serviceCmd        *cobra.Command
	serviceStartCmd   *cobra.Command
	serviceRestartCmd *cobra.Command
	serviceStopCmd    *cobra.Command
	serviceStatusCmd  *cobra.Command
	serviceInitSysCmd *cobra.Command

	adapterCmd             *cobra.Command
	adapterGetTypesCmd     *cobra.Command
	adapterGetInstancesCmd *cobra.Command

	volumeCmd        *cobra.Command
	volumeListCmd    *cobra.Command
	volumeCreateCmd  *cobra.Command
	volumeRemoveCmd  *cobra.Command
	volumeAttachCmd  *cobra.Command
	volumeDetachCmd  *cobra.Command
	volumeMountCmd   *cobra.Command
	volumeUnmountCmd *cobra.Command
	volumePathCmd    *cobra.Command

	snapshotCmd       *cobra.Command
	snapshotGetCmd    *cobra.Command
	snapshotCreateCmd *cobra.Command
	snapshotRemoveCmd *cobra.Command
	snapshotCopyCmd   *cobra.Command

	deviceCmd        *cobra.Command
	deviceGetCmd     *cobra.Command
	deviceMountCmd   *cobra.Command
	devuceUnmountCmd *cobra.Command
	deviceFormatCmd  *cobra.Command

	outputFormat            string
	outputTemplate          string
	outputTemplateTabs      bool
	fg                      bool
	fork                    bool
	force                   bool
	cfgFile                 string
	snapshotID              string
	volumeID                string
	runAsync                bool
	volumeAttached          bool
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
func New(ctx apitypes.Context) *CLI {
	return NewWithArgs(ctx, os.Args[1:]...)
}

// NewWithArgs returns a new CLI using the specified arguments.
func NewWithArgs(ctx apitypes.Context, a ...string) *CLI {

	validateConfig(util.EtcFilePath("config.yml"))
	validateConfig(fmt.Sprintf("%s/.rexray/config.yml", gotil.HomeDir()))

	s := "REX-Ray:\n" +
		"  A guest-based storage introspection tool that enables local\n" +
		"  visibility and management from cloud and storage platforms."

	c := &CLI{
		l:      log.New(),
		ctx:    ctx,
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
func Execute(ctx apitypes.Context) {
	New(ctx).Execute()
}

// ExecuteWithArgs executes the CLI using the specified arguments.
func ExecuteWithArgs(ctx apitypes.Context, a ...string) {
	NewWithArgs(ctx, a...).Execute()
}

// Execute executes the CLI.
func (c *CLI) Execute() {

	defer func() {
		if c.activateLibStorage {
			util.WaitUntilLibStorageStopped(c.ctx, c.rsErrs)
		}
	}()

	defer func() {
		r := recover()
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

func (c *CLI) addOutputFormatFlag(fs *pflag.FlagSet) {
	fs.StringVarP(
		&c.outputFormat, "format", "f", "tmpl",
		"The output format (tmpl, json, jsonp)")
	fs.StringVarP(
		&c.outputTemplate, "template", "", "",
		"The Go template to use when --format is set to 'tmpl'")
	fs.BoolVarP(
		&c.outputTemplateTabs, "templateTabs", "", true,
		"Set to true to use a Go tab writer with the output template")
}

func (c *CLI) updateLogLevel() {
	lvl, err := log.ParseLevel(c.logLevel())
	if err != nil {
		return
	}
	c.ctx.WithField("level", lvl).Debug("updating log level")
	log.SetLevel(lvl)
	c.config.Set(apitypes.ConfigLogLevel, lvl.String())
	context.SetLogLevel(c.ctx, lvl)
	log.WithField("logLevel", lvl).Info("updated log level")
}

func (c *CLI) preRunActivateLibStorage(cmd *cobra.Command, args []string) {
	c.activateLibStorage = true
	c.preRun(cmd, args)
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

	// disable path caching for the CLI
	c.config.Set(apitypes.ConfigIgVolOpsPathCacheEnabled, false)

	if v := c.rrHost(); v != "" {
		c.config.Set(apitypes.ConfigHost, v)
	}
	if v := c.rrService(); v != "" {
		c.config.Set(apitypes.ConfigService, v)
	}

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

	c.ctx.WithField("val", os.Args).Debug("os.args")

	if c.activateLibStorage {

		c.ctx.WithField("cmd", cmd.Name()).Debug("activating libStorage")

		if c.runAsync {
			c.ctx = c.ctx.WithValue("async", true)
		}

		apiserver.DisableStartupInfo = true

		var err error

		// activate libStorage if necessary
		c.ctx, c.config, _, err = util.ActivateLibStorage(c.ctx, c.config)

		if err == nil {
			c.r, err = apiclient.New(c.ctx, c.config)
		}

		if err != nil {
			if term.IsTerminal() {
				printColorizedError(err)
			} else {
				printNonColorizedError(err)
			}
			fmt.Println()
			cmd.Help()
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

func (c *CLI) rrHost() string {
	return c.config.GetString("rexray.host")
}

func (c *CLI) rrService() string {
	return c.config.GetString("rexray.service")
}

func (c *CLI) logLevel() string {
	return c.config.GetString("rexray.logLevel")
}

func store() apitypes.Store {
	return apiutils.NewStore()
}
