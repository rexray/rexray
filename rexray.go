package rexray

import (
	"io"
	"os"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/config"

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
	c := config.New()
	if err := c.ReadConfigFile(path); err != nil {
		return nil, err
	}
	var err error
	var r *core.RexRay
	if r, err = core.New(c); err != nil {
		return nil, err
	}
	return r, nil
}

// NewWithConfigReader creates a new REX-Ray instance and configures it with a
// custom configuration stream.
func NewWithConfigReader(in io.Reader) (*core.RexRay, error) {
	c := config.New()
	if err := c.ReadConfig(in); err != nil {
		return nil, err
	}
	var err error
	var r *core.RexRay
	if r, err = core.New(c); err != nil {
		return nil, err
	}
	return r, nil
}

// New creates a new REX-Ray instance and configures using the standard
// configuration workflow: environment variables followed by global and user
// configuration files.
func New() (*core.RexRay, error) {
	c := config.New()
	var err error
	var r *core.RexRay
	if r, err = core.New(c); err != nil {
		return nil, err
	}
	return r, nil
}
