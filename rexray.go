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
	"io"
	"os"

	"github.com/akutz/gofig"

	"github.com/emccode/rexray/core"

	// This blank import loads the drivers package
	_ "github.com/emccode/rexray/drivers"
)

// NewWithEnv creates a new REX-Ray instance and configures it with a a custom
// environment.
func NewWithEnv(env map[string]string) (*core.RexRay, error) {
	if env != nil {
		for k, v := range env {
			os.Setenv(k, v)
		}
	}
	return New()
}

// NewWithConfigFile creates a new REX-Ray instance and configures it with a
// custom configuration file.
func NewWithConfigFile(path string) (*core.RexRay, error) {
	c := gofig.New()
	if err := c.ReadConfigFile(path); err != nil {
		return nil, err
	}
	return core.New(c), nil
}

// NewWithConfigReader creates a new REX-Ray instance and configures it with a
// custom configuration stream.
func NewWithConfigReader(in io.Reader) (*core.RexRay, error) {
	c := gofig.New()
	if err := c.ReadConfig(in); err != nil {
		return nil, err
	}
	return core.New(c), nil
}

// New creates a new REX-Ray instance and configures using the standard
// configuration workflow: environment variables followed by global and user
// configuration files.
func New() (*core.RexRay, error) {
	return core.New(gofig.New()), nil
}
