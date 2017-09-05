package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

const (
	// defaultVersion is the default CSI_VERSION string if none
	// is provided via a CLI argument or environment variable
	defaultVersion = "0.1.0"

	// maxUint32 is the maximum value for a uint32. this is
	// defined as math.MaxUint32, but it's redefined here
	// in order to avoid importing the math package for just
	// a constant value
	maxUint32 = 4294967295

	// maxInt32 is the maximum value for an int32. this is
	// defined as math.MaxInt32, but it's redefined here
	// in order to avoid importing the math package for just
	// a constant value
	maxInt32 = 2147483647
)

var appName = path.Base(os.Args[0])

func main() {

	// the program should have at least two args:
	//
	//     args[0]  path of executable
	//     args[1]  csi rpc
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(1)
	}

	// match the name of the rpc or one of its aliases
	rpc := os.Args[1]
	c := func(ccc ...[]*cmd) *cmd {
		for _, cc := range ccc {
			for _, c := range cc {
				if strings.EqualFold(rpc, c.Name) {
					rpc = c.Name
					return c
				}
				for _, a := range c.Aliases {
					if strings.EqualFold(rpc, a) {
						rpc = a
						return c
					}
				}
			}
		}
		return nil
	}(controllerCmds, identityCmds, nodeCmds)

	// assert that a command for the requested rpc was found
	if c == nil {
		fmt.Fprintf(os.Stderr, "error: invalid rpc: %s\n", rpc)
		usage(os.Stderr)
		os.Exit(1)
	}

	if c.Action == nil {
		panic("nil rpc action")
	}
	if c.Flags == nil {
		panic("nil rpc flags")
	}

	ctx := context.Background()

	// parse the command line with the command's flag set
	cflags := c.Flags(ctx, rpc)
	if err := cflags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// assert that the endpoint value is required
	if args.endpoint == "" {
		fmt.Fprintln(os.Stderr, "error: endpoint is required")
		cflags.Usage()
		os.Exit(1)
	}

	// assert that the version is required and valid
	versionRX := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	versionMatch := versionRX.FindStringSubmatch(args.szVersion)
	if len(versionMatch) == 0 {
		fmt.Fprintf(
			os.Stderr,
			"error: invalid version: %s\n",
			args.szVersion)
		os.Exit(1)
	}
	versionMajor, _ := strconv.Atoi(versionMatch[1])
	if versionMajor > maxUint32 {
		fmt.Fprintf(
			os.Stderr, "error: MAJOR > uint32: %v\n", versionMajor)
		os.Exit(1)
	}
	versionMinor, _ := strconv.Atoi(versionMatch[2])
	if versionMinor > maxUint32 {
		fmt.Fprintf(
			os.Stderr, "error: MINOR > uint32: %v\n", versionMinor)
		os.Exit(1)
	}
	versionPatch, _ := strconv.Atoi(versionMatch[3])
	if versionPatch > maxUint32 {
		fmt.Fprintf(
			os.Stderr, "error: PATCH > uint32: %v\n", versionPatch)
		os.Exit(1)
	}
	args.version = &csi.Version{
		Major: uint32(versionMajor),
		Minor: uint32(versionMinor),
		Patch: uint32(versionPatch),
	}

	// initialize a grpc client
	gclient, err := newGrpcClient(ctx, args.endpoint, args.insecure)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// if a service is specified then add it to the context
	// as gRPC metadata
	if args.service != "" {
		ctx = metadata.NewContext(
			ctx, metadata.Pairs("csi.service", args.service))
	}

	// execute the command
	if err := c.Action(ctx, cflags, gclient); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if _, ok := err.(*errUsage); ok {
			cflags.Usage()
		}
		os.Exit(1)
	}
}

func newGrpcClient(
	ctx context.Context,
	endpoint string,
	insecure bool) (*grpc.ClientConn, error) {

	dialOpts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(gocsi.ChainUnaryClient(
			gocsi.ClientCheckReponseError,
			gocsi.ClientResponseValidator)),
		grpc.WithDialer(
			func(target string, timeout time.Duration) (net.Conn, error) {
				proto, addr, err := gocsi.ParseProtoAddr(target)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout(proto, addr, timeout)
			}),
	}

	if insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	return grpc.DialContext(ctx, endpoint, dialOpts...)
}

///////////////////////////////////////////////////////////////////////////////
//                            Default Formats                                //
///////////////////////////////////////////////////////////////////////////////

// mapSzOfSzFormat is the default Go template format for
// emitting a map[string]string
const mapSzOfSzFormat = `{{range $k, $v := .}}` +
	`{{printf "%s=%s\t" $k $v}}{{end}}{{"\n"}}`

// volumeInfoFormat is the default Go template format for
// emitting a *csi.VolumeInfo
const volumeInfoFormat = `{{with .GetId}}{{range $k, $v := .GetValues}}` +
	`{{printf "%s=%s\t" $k $v}}{{end}}{{end}}{{"\n"}}`

// versionFormat is the default Go template format for emitting a *csi.Version
const versionFormat = `{{.GetMajor}}.{{.GetMinor}}.{{.GetPatch}}{{"\n"}}`

// pluginInfoFormat is the default Go template format for
// emitting a *csi.GetPluginInfoResponse_Result
const pluginInfoFormat = `{{.Name}}{{print "\t"}}{{.VendorVersion}}{{print "\t"}}` +
	`{{with .GetManifest}}{{range $k, $v := .}}` +
	`{{printf "%s=%s\t" $k $v}}{{end}}{{end}}{{"\n"}}`

// capFormat is the default Go template for emitting a
// *csi.{Controller,Node}ServiceCapability
const capFormat = `{{with .GetRpc}}{{.Type}}{{end}}{{"\n"}}`

// valCapFormat is the default Go tempate for emitting a
// *csi.ValidateVolumeCapabilitiesResponse_Result
const valCapFormat = `{{with .GetSupported}}{{print "supported: "}}{{.}}` +
	`{{print "\n"}}{{end}}{{with .GetMessage}}{{print "\tmessage: "}}` +
	`{{.}}{{end}}{{"\n"}}`

///////////////////////////////////////////////////////////////////////////////
//                                Commands                                   //
///////////////////////////////////////////////////////////////////////////////
type errUsage struct {
	msg string
}

func (e *errUsage) Error() string {
	return e.msg
}

type cmd struct {
	Name    string
	Aliases []string
	Action  func(context.Context, *flag.FlagSet, *grpc.ClientConn) error
	Flags   func(context.Context, string) *flag.FlagSet
}

///////////////////////////////////////////////////////////////////////////////
//                                Usage                                      //
///////////////////////////////////////////////////////////////////////////////
func usage(w io.Writer) {
	const h = `usage: {{.Name}} RPC [ARGS...]{{range $Name, $Cmds := .Categories}}

       {{$Name}} RPCs{{range $Cmds}}
         {{.Name}}{{if .Aliases}} ({{join .Aliases ", "}}){{end}}{{end}}{{end}}

Use the -? flag with an RPC for additional help.
`
	f := template.FuncMap{"join": strings.Join}
	t := template.Must(template.New(appName).Funcs(f).Parse(h))
	d := struct {
		Name       string
		Categories map[string][]*cmd
	}{
		appName,
		map[string][]*cmd{
			"CONTROLLER": controllerCmds,
			"IDENTITY":   identityCmds,
			"NODE":       nodeCmds,
		},
	}
	t.Execute(w, d)
}

///////////////////////////////////////////////////////////////////////////////
//                               Global Flags                                //
///////////////////////////////////////////////////////////////////////////////
var args struct {
	service   string
	endpoint  string
	format    string
	help      bool
	insecure  bool
	szVersion string
	version   *csi.Version
}

func flagsGlobal(
	fs *flag.FlagSet,
	formatDefault, formatObjectType string) {

	fs.StringVar(
		&args.endpoint,
		"endpoint",
		os.Getenv("CSI_ENDPOINT"),
		"The endpoint address")

	fs.StringVar(
		&args.service,
		"service",
		"",
		"The name of the CSD service to use.")

	version := defaultVersion
	if v := os.Getenv("CSI_VERSION"); v != "" {
		version = v
	}
	fs.StringVar(
		&args.szVersion,
		"version",
		version,
		"The API version string")

	insecure := true
	if v := os.Getenv("CSI_INSECURE"); v != "" {
		insecure, _ = strconv.ParseBool(v)
	}
	fs.BoolVar(
		&args.insecure,
		"insecure",
		insecure,
		"Disables transport security")

	fmtMsg := &bytes.Buffer{}
	fmt.Fprint(fmtMsg, "The Go template used to print an object.")
	if formatObjectType != "" {
		fmt.Fprintf(fmtMsg, " This command emits a %s.", formatObjectType)
	}
	fs.StringVar(
		&args.format,
		"format",
		formatDefault,
		fmtMsg.String())
}

// stringSliceArg is used for parsing a csv arg into a string slice
type stringSliceArg struct {
	szVal string
	vals  []string
}

func (s *stringSliceArg) String() string {
	return s.szVal
}

func (s *stringSliceArg) Set(val string) error {
	s.vals = append(s.vals, strings.Split(val, ",")...)
	return nil
}

// mapOfStringArg is used for parsing a csv, key=value arg into
// a map[string]string
type mapOfStringArg struct {
	szVal string
	vals  map[string]string
}

func (s *mapOfStringArg) String() string {
	return s.szVal
}

func (s *mapOfStringArg) Set(val string) error {
	if s.vals == nil {
		s.vals = map[string]string{}
	}
	vals := strings.Split(val, ",")
	for _, v := range vals {
		vp := strings.SplitN(v, "=", 2)
		switch len(vp) {
		case 1:
			s.vals[vp[0]] = ""
		case 2:
			s.vals[vp[0]] = vp[1]
		}
	}
	return nil
}
