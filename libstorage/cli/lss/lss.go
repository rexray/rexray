package lss

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"

	"github.com/akutz/gotil"
	flag "github.com/spf13/pflag"

	"github.com/AVENTER-UG/rexray/core"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	apiconfig "github.com/AVENTER-UG/rexray/libstorage/api/utils/config"

	// load the drivers
	_ "github.com/AVENTER-UG/rexray/libstorage/imports/config"
	_ "github.com/AVENTER-UG/rexray/libstorage/imports/storage"
)

var (
	cliFlags        *flag.FlagSet
	flagHost        *string
	flagConfig      *string
	flagLogLvl      *string
	flagHelp        *bool
	flagVerbose     *bool
	flagVersion     *bool
	flagEnv         *bool
	flagPrintConfig *bool
	config          gofig.Config
)

func init() {
	cliFlags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagConfig = cliFlags.StringP("config", "c", "", "path")
	flagHost = cliFlags.StringP("host", "h", "", "<proto>://<addr>")
	flagLogLvl = cliFlags.StringP("log", "l", "info", "error|warn|info|debug")
	flagHelp = cliFlags.BoolP("help", "?", false, "print usage")
	flagVersion = cliFlags.Bool("version", false, "print version info")
	flagEnv = cliFlags.Bool("env", false, "print env info")
	flagPrintConfig = cliFlags.Bool("printConfig", false, "print config info")
	flagVerbose = cliFlags.BoolP("verbose", "v", false, "print verbose usage")
	flag.CommandLine.AddFlagSet(cliFlags)
}

// Run the server.
func Run() {
	server.CloseOnAbort()

	flag.Usage = printUsage
	flag.Parse()

	if flagVersion != nil && *flagVersion {
		_, _, thisExeAbsPath := gotil.GetThisPathParts()
		fmt.Fprintf(apitypes.Stdout, "Binary: %s\n", thisExeAbsPath)
		fmt.Fprint(apitypes.Stdout, core.Version.String())
		os.Exit(0)
	}

	if flagEnv != nil && *flagEnv {
		for _, v := range os.Environ() {
			fmt.Fprintf(apitypes.Stdout, "%s\n", v)
		}
		os.Exit(0)
	}

	// if a config is specified then do not care about any other options
	if flagConfig != nil && gotil.FileExists(*flagConfig) {

		config = gofigCore.New()

		if err := config.ReadConfigFile(*flagConfig); err != nil {
			fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
			os.Exit(1)
		}

		if flagPrintConfig != nil && *flagPrintConfig {
			jstr, err := config.ToJSON()
			if err != nil {
				fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
				os.Exit(1)
			}
			fmt.Fprintln(apitypes.Stdout, jstr)
			os.Exit(0)
		}

		s, errs, err := server.Serve(nil, config)
		if err != nil {
			fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
			os.Exit(1)
		}

		err = <-errs
		if err != nil {
			fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
			os.Exit(1)
		}

		s.Close()
		os.Exit(0)
	}

	ctx := context.Background()
	ctx = ctx.WithValue(context.PathConfigKey, utils.NewPathConfig(ctx, "", ""))
	registry.ProcessRegisteredConfigs(ctx)

	cfg, err := apiconfig.NewConfig(ctx)
	if err != nil {
		fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	config = cfg
	for _, fs := range config.FlagSets() {
		flag.CommandLine.AddFlagSet(fs)
	}

	if flagHelp != nil && *flagHelp {
		flag.Usage()
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
	}

	if flagHost != nil {
		os.Setenv("LIBSTORAGE_HOST", *flagHost)
	}

	if flagLogLvl != nil {
		os.Setenv("LIBSTORAGE_LOGGING_LEVEL", *flagLogLvl)
	}

	if lvl, err := log.ParseLevel(
		config.GetString(apitypes.ConfigLogLevel)); err == nil {
		log.SetLevel(lvl)
	}

	if flagPrintConfig != nil && *flagPrintConfig {
		jstr, err := config.ToJSON()
		if err != nil {
			fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
			os.Exit(1)
		}
		fmt.Fprintln(apitypes.Stdout, jstr)
		os.Exit(0)
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "libstorage:\n  server:\n    services:\n")
	for _, ds := range flag.Args() {
		dsp := strings.Split(ds, ":")
		dn := dsp[0]
		sn := dsp[0]
		if len(dsp) > 1 {
			sn = dsp[1]
		}
		fmt.Fprintf(buf, "      %s:\n        driver: %s\n", sn, dn)
	}
	if err := config.ReadConfig(buf); err != nil {
		fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	server.CloseOnAbort()

	_, errs, err := server.Serve(ctx, config)
	if err != nil {
		fmt.Fprintf(apitypes.Stderr, "%s: error: %v\n", os.Args[0], err)
		os.Exit(1)
	}

	<-errs
}

func printUsage() {
	firstLine := fmt.Sprintf("usage: %s", os.Args[0])
	fmt.Fprintf(apitypes.Stderr, "%s\n", firstLine)
	padFmt := fmt.Sprintf("%%%ds\n", len(firstLine))
	fmt.Fprintf(apitypes.Stderr, padFmt, "-c,--config <configFilePath> [--printConfig]")
	fmt.Fprintf(apitypes.Stderr, padFmt, "--version")
	fmt.Fprintf(apitypes.Stderr, padFmt, "--env")
	fmt.Fprintf(apitypes.Stderr, padFmt, "[-options] <driver>[:<service>] [<driver>[:<service>]...]")
	fmt.Fprintf(apitypes.Stderr, "\n")

	fmt.Fprintln(apitypes.Stderr, cliFlags.FlagUsages())
	fmt.Fprintln(apitypes.Stderr, hostUsage)
	fmt.Fprintln(apitypes.Stderr, logUsage)
	fmt.Fprintf(apitypes.Stderr, driversUsage, os.Args[0])

	if flagVerbose != nil && *flagVerbose {
		for fsn, fs := range config.FlagSets() {
			fmt.Fprintln(apitypes.Stderr, fsn)
			fmt.Fprintln(apitypes.Stderr, fs.FlagUsages())
		}
		fmt.Fprintln(apitypes.Stderr)
	}

	os.Exit(1)
}

const (
	driversUsage = `  Drivers and Services

    After all of the flags and options are processed, the remaining
    arguments on the command line are parsed as drivers and services.

    Arguments can be one of two formats:

      <driver>            The name of a driver to be hosted by a
                          service configured with the same name as
                          the driver.

      <driver>:<service>  This format is a colon-delimited pair of
                          the name of the driver to host and the
                          explicit name of the service to configure
                          for hosting the driver.

    For example, the command "%[1]s vfs" would start a new server and
    load the "vfs" driver using a service also named "vfs". Meanwhile,
    the command "%[1]s vfs:s1" would also start a new server and load the
    "vfs" driver, but the service hosting the driver would be named
    "s1".

    Some more examples:

      %[1]s vfs:s1 scaleio:s2

        Launch a server with the vfs and scaleio drivers hosted by the
        services s1 and s2, respectively.

      %[1]s vfs scaleio:scaleio-service-00

        Launch a server with the vfs driver hosted by an eponymous
        service whereas the scaleio driver is hosted by the service
        scaleio-service-00.

`

	hostUsage = `  Host Address

    The -h flag expects a Golang-formatted network address -- this will be
    the address on which the server is hosted. Omitting the host address is
    fine as well as one will be generated automatically.

    Valid formats include TCP addresses such as "tcp://127.0.0.1:7979"
    in order to host the server on the loopback adapter on port 7979. The TCP
    address can also be specified as "tcp://:7979". This creates the server
    and binds it to port 7979 on all available network interfaces.

    This flag also accepts the Golang UNIX socket network address such as
    "unix:///tmp/libstorage.sock".
`

	logUsage = `  Log Level

    The -l flag updates the log level to either error, warn, info, or debug.
`
)
