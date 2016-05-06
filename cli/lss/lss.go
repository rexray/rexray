package lss

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	flag "github.com/spf13/pflag"

	"github.com/emccode/libstorage/api/server"
	apitypes "github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	apiconfig "github.com/emccode/libstorage/api/utils/config"

	// load the drivers
	_ "github.com/emccode/libstorage/imports/config"
	_ "github.com/emccode/libstorage/imports/remote"
	_ "github.com/emccode/libstorage/imports/routers"
)

var (
	cliFlags    *flag.FlagSet
	flagHost    *string
	flagConfig  *string
	flagLogLvl  *string
	flagHelp    *bool
	flagVerbose *bool
	config      gofig.Config
)

func init() {
	cliFlags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagConfig = cliFlags.StringP("config", "c", "", "path")
	flagHost = cliFlags.StringP("host", "h", "", "<proto>://<addr>")
	flagLogLvl = cliFlags.StringP("log", "l", "info", "error|warn|info|debug")
	flagHelp = cliFlags.BoolP("help", "?", false, "print usage")
	flagVerbose = cliFlags.BoolP("verbose", "v", false, "print verbose usage")
	flag.CommandLine.AddFlagSet(cliFlags)
}

// Run the server.
func Run() {
	flag.Usage = printUsage
	flag.Parse()

	cfg, err := apiconfig.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %v\n", os.Args[0], err)
		os.Exit(1)
	}
	if flagConfig != nil && gotil.FileExists(*flagConfig) {
		if err := cfg.ReadConfigFile(*flagConfig); err != nil {
			fmt.Fprintf(os.Stderr, "%s: error: %v\n", os.Args[0], err)
			os.Exit(1)
		}
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

	driversAndServices := []string{}
	for _, ds := range flag.Args() {
		dsp := strings.Split(ds, ":")
		driversAndServices = append(driversAndServices, dsp[0])
		if len(dsp) > 1 {
			driversAndServices = append(driversAndServices, dsp[1])
		}
	}

	server.CloseOnAbort()

	host := config.GetString(apitypes.ConfigHost)
	if host == "" {
		host = fmt.Sprintf("unix://%s", utils.GetTempSockFile())
	}

	if err := server.Run(
		host,
		false,
		driversAndServices...); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-options] ", os.Args[0])
	fmt.Fprintf(os.Stderr, "<driver>[:<service>] [<driver>[:<service>]...]")
	fmt.Fprintf(os.Stderr, "\n\n")

	fmt.Fprintln(os.Stderr, cliFlags.FlagUsages())
	fmt.Fprintln(os.Stderr, hostUsage)
	fmt.Fprintln(os.Stderr, logUsage)
	fmt.Fprintf(os.Stderr, driversUsage, os.Args[0])

	if flagVerbose != nil && *flagVerbose {
		for fsn, fs := range config.FlagSets() {
			fmt.Fprintln(os.Stderr, fsn)
			fmt.Fprintln(os.Stderr, fs.FlagUsages())
		}
		fmt.Fprintln(os.Stderr)
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
