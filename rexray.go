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
//     import "github.com/emccode/rexray"
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
	"fmt"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	// load libStorage
	_ "github.com/emccode/libstorage"
	_ "github.com/emccode/libstorage/imports/local"
	_ "github.com/emccode/libstorage/imports/remote"
	"github.com/emccode/rexray/util"
)

func init() {
	gofig.SetGlobalConfigPath(util.EtcDirPath())
	gofig.SetUserConfigPath(fmt.Sprintf("%s/.rexray", gotil.HomeDir()))
	gofig.Register(globalRegistration())
}

func globalRegistration() *gofig.Registration {
	r := gofig.NewRegistration("Global")
	r.Yaml(`
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
	return r
}
