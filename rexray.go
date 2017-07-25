// Package rexray provides visibility and management of external/underlying
// storage via guest storage introspection. Available as a Go package, CLI tool,
// and Linux service, and with built-in third-party support for tools such as
// Docker, REX-Ray is easily integrated into any workflow. For example, here's
// how to list storage for a guest hosted on Amazon Web Services (AWS) with
// REX-Ray:
//
//     [0]akutz@pax:~$ export REXRAY_STORAGEDRIVERS=ec2
//     [0]akutz@pax:~$ export AWS_ACCESSKEY=access_key
//     [0]akutz@pax:~$ export AWS_SECRETKEY=secret_key
//     [0]akutz@pax:~$ rexray volume get
//
//     - providername: ec2
//       instanceid: i-695bb6ab
//       volumeid: vol-dedbadc3
//       devicename: /dev/sda1
//       region: us-west-1
//       status: attached
//     - providername: ec2
//       instanceid: i-695bb6ab
//       volumeid: vol-04c4b219
//       devicename: /dev/xvdb
//       region: us-west-1
//       status: attached
//
//     [0]akutz@pax:~$
//
// Using REX-Ray as a library is easy too. To perform the same volume listing
// as above, simply use the following snippet:
//
//     import "github.com/codedellemc/rexray"
//
//     r := rexray.NewWithEnv(map[string]string{
//         "REXRAY_STORAGEDRIVERS": "ec2",
//         "AWS_ACCESSKEY": "access_key",
//         "AWS_SECRETKEY": "secret_key"})
//
//     r.InitDrivers()
//
//     volumes, err := r.Storage.GetVolumeMapping()
//
package rexray

import (
	"path"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"

	"github.com/codedellemc/rexray/util"

	// load the libstorage packages
	_ "github.com/codedellemc/rexray/libstorage/imports/config"
)

func init() {
	gofigCore.SetGlobalConfigPath(util.EtcDirPath())
	gofigCore.SetUserConfigPath(path.Join(gotil.HomeDir(), util.DotDirName))
	r := gofigCore.NewRegistration("Global")
	r.SetYAML(`
rexray:
    logLevel: warn
`)
	r.Key(gofig.String, "h", "",
		"The libStorage host.", "rexray.host",
		"host")
	r.Key(gofig.String, "s", "",
		"The libStorage service.", "rexray.service",
		"service")
	r.Key(gofig.String, "l", "warn",
		"The log level (error, warn, info, debug)", "rexray.logLevel",
		"logLevel")
	gofigCore.Register(r)
}
