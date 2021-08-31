package cli

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	gofig "github.com/akutz/gofig/types"
	glog "github.com/akutz/golf/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiutils "github.com/AVENTER-UG/rexray/libstorage/api/utils"

	"github.com/AVENTER-UG/rexray/util"
)

var initCmdFuncs []func(*CLI)

func init() {
	log.SetFormatter(&glog.TextFormatter{TextFormatter: log.TextFormatter{}})
}

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

	tokenCmd       *cobra.Command
	tokenNewCmd    *cobra.Command
	tokenDecodeCmd *cobra.Command

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
	serviceStopCmd    *cobra.Command
	serviceRestartCmd *cobra.Command
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

	scriptsCmd          *cobra.Command
	scriptsListCmd      *cobra.Command
	scriptsInstallCmd   *cobra.Command
	scriptsUninstallCmd *cobra.Command

	flexRexCmd          *cobra.Command
	flexRexInstallCmd   *cobra.Command
	flexRexUninstallCmd *cobra.Command
	flexRexStatusCmd    *cobra.Command

	verify                  bool
	key                     string
	alg                     string
	attach                  bool
	expires                 time.Duration
	amount                  bool
	quiet                   bool
	dryRun                  bool
	continueOnError         bool
	outputFormat            string
	outputTemplate          string
	outputTemplateTabs      bool
	fg                      bool
	nopid                   bool
	fork                    bool
	force                   bool
	cfgFile                 string
	snapshotID              string
	volumeID                string
	runAsync                bool
	volumeAttached          bool
	volumeAvailable         bool
	volumePath              bool
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
	encrypted               bool
	encryptionKey           string
	idempotent              bool
	scriptPath              string
	serverCertFile          string
	serverKeyFile           string
	serverCAFile            string
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

// New returns a new CLI using the current process's arguments.
func New(ctx apitypes.Context, config gofig.Config) *CLI {
	return NewWithArgs(ctx, config, os.Args[1:]...)
}

// NewWithArgs returns a new CLI using the specified arguments.
func NewWithArgs(
	ctx apitypes.Context, config gofig.Config, a ...string) *CLI {

	s := "REX-Ray:\n" +
		"  A guest-based storage introspection tool that enables local\n" +
		"  visibility and management from cloud and storage platforms."

	if config == nil {
		config = util.NewConfig(ctx)
	}

	noLibStorage := false
	if strings.EqualFold("false", os.Getenv("LIBSTORAGE")) {
		config.Set("libstorage.disabled", true)
		noLibStorage = true
		ctx.Info("libstorage disabled by env var")
	} else if config.GetBool("libstorage.disabled") {
		os.Setenv("LIBSTORAGE", "false")
		noLibStorage = true
		ctx.Info("libstorage disabled by config")
	}
	if noLibStorage && config.Get("csi.driver") == "libstorage" {
		config.Set("csi.driver", "csi-vfs")
	}

	c := &CLI{
		l:      log.New(),
		ctx:    ctx,
		config: config,
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

	for _, f := range initCmdFuncs {
		f(c)
	}

	c.initUsageTemplates()

	return c
}

// Execute executes the CLI using the current process's arguments.
func Execute(ctx apitypes.Context, config gofig.Config) {
	New(ctx, config).Execute()
}

// ExecuteWithArgs executes the CLI using the specified arguments.
func ExecuteWithArgs(
	ctx apitypes.Context, config gofig.Config, a ...string) {

	NewWithArgs(ctx, config, a...).Execute()
}

// Execute executes the CLI.
func (c *CLI) Execute() {
	defer func() {
		if c.activateLibStorage {
			util.WaitUntilLibStorageStopped(c.ctx, c.rsErrs)
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
func (c *CLI) addQuietFlag(fs *pflag.FlagSet) {
	fs.BoolVarP(&c.quiet, "quiet", "q", false, "Suppress table headers")
}

func (c *CLI) addDryRunFlag(fs *pflag.FlagSet) {
	fs.BoolVarP(&c.dryRun, "dryRun", "n", false,
		"Show what action(s) will occur, but do not execute them")
}

func (c *CLI) addContinueOnErrorFlag(fs *pflag.FlagSet) {
	fs.BoolVar(&c.continueOnError, "continueOnError", false,
		"Continue processing a collection upon error")
}

func (c *CLI) addIdempotentFlag(fs *pflag.FlagSet) {
	fs.BoolVarP(&c.idempotent, "idempotent", "i", false,
		"Make this command idempotent.")
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
	// Disable patch caching for the CLI
	c.config = c.config.Scope(configRexrayCLI)
	c.config.Set(apitypes.ConfigIgVolOpsPathCacheEnabled, false)

	c.activateLibStorage = true
	c.preRun(cmd, args)
}

const (
	configRexrayCLI = "rexray.cli"
)

func (c *CLI) preRun(cmd *cobra.Command, args []string) {
	c.updateLogLevel()

	if v := c.rrHost(); v != "" {
		c.config.Set(apitypes.ConfigHost, v)
	}
	if v := c.rrService(); v != "" {
		c.config.Set(apitypes.ConfigService, v)
	}

	if isHelpFlags(cmd) {
		cmd.Help()
		os.Exit(0)
	}

	if permErr := c.checkCmdPermRequirements(cmd); permErr != nil {
		if util.IsTerminal(os.Stderr) {
			printColorizedError(permErr)
		} else {
			printNonColorizedError(permErr)
		}

		fmt.Println()
		cmd.Help()
		os.Exit(1)
	}

	c.ctx.WithField("val", os.Args).Debug("os.args")

	if c.activateLibStorage {

		if c.runAsync {
			c.ctx = c.ctx.WithValue("async", true)
		}

		c.ctx.WithField("cmd", cmd.Name()).Debug("activating libStorage")

		var err error
		c.ctx, c.config, c.rsErrs, err = util.ActivateLibStorage(
			c.ctx, c.config)
		if err == nil {
			c.ctx.WithField("cmd", cmd.Name()).Debug(
				"creating libStorage client")
			c.r, err = util.NewClient(c.ctx, c.config)
			err = c.handleKnownHostsError(err)
		}

		if err != nil {
			if util.IsTerminal(os.Stderr) {
				printColorizedError(err)
			} else {
				printNonColorizedError(err)
			}
			fmt.Println()
			cmd.Help()
			os.Exit(1)
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
		blueBg, "https://github.com/AVENTER-UG/rexray")
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
		"  - The REX-ray website at https://github.com/AVENTER-UG/rexray\n")
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

// handles the known_hosts error,
// if error is ErrKnownHosts, stop execution to prevent unstable state
func (c *CLI) handleKnownHostsError(err error) error {
	if err == nil {
		return nil
	}
	urlErr, ok := err.(*url.Error)
	if !ok {
		return err
	}

	var khErr *apitypes.ErrKnownHost
	var khConflictErr *apitypes.ErrKnownHostConflict
	switch err := urlErr.Err.(type) {
	case *apitypes.ErrKnownHost:
		khErr = err
	case *apitypes.ErrKnownHostConflict:
		khConflictErr = err
	}

	if khErr == nil && khConflictErr == nil {
		return err
	}

	pathConfig := context.MustPathConfig(c.ctx)
	knownHostPath := pathConfig.UserDefaultTLSKnownHosts

	if khConflictErr != nil {
		// it's an ErrKnownHostConflict
		fmt.Fprintf(
			os.Stderr,
			hostKeyCheckFailedFormat,
			khConflictErr.PeerFingerprint,
			knownHostPath,
			khConflictErr.KnownHostName)
		os.Exit(1)
	}

	// it's an ErrKnownHost

	if !util.AssertTrustedHost(
		c.ctx,
		khErr.HostName,
		khErr.PeerAlg,
		khErr.PeerFingerprint) {
		fmt.Fprintln(
			os.Stderr,
			"Aborting request, remote host not trusted.")
		os.Exit(1)
	}

	if err := util.AddKnownHost(
		c.ctx,
		knownHostPath,
		khErr.HostName,
		khErr.PeerAlg,
		khErr.PeerFingerprint,
	); err == nil {
		fmt.Fprintf(
			os.Stderr,
			"Permanently added host %s to known_hosts file %s\n",
			khErr.HostName, knownHostPath)
		fmt.Fprintln(
			os.Stderr,
			"It is safe to retry your last rexray command.")
	} else {
		fmt.Fprintf(
			os.Stderr,
			"Failed to add entry to known_hosts file: %v",
			err)
	}

	os.Exit(1) // do not continue
	return nil
}

func store() apitypes.Store {
	return apiutils.NewStore()
}

func checkOpPerms(op string) error {
	//if os.Geteuid() != 0 {
	//	return goof.Newf("REX-Ray can only be %s by root", op)
	//}
	return nil
}

const hostKeyCheckFailedFormat = `@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@ WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED! @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now (man-in-the-middle attack)!
It is also possible that the RSA host key has just been changed.
The fingerprint for the RSA key sent by the remote host is
%[1]x.
Please contact your system administrator.
Add correct host key in %[2]s to get rid of this message.
Offending key in %[2]s
RSA host key for %[3]s has changed and you have requested strict checking.
Host key verification failed.
`
