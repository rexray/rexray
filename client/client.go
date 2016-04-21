package client

import (
	"fmt"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	apiconfig "github.com/emccode/libstorage/api/utils/config"

	// load the drivers
	_ "github.com/emccode/libstorage/drivers/os"
)

var (
	libstorHome = fmt.Sprintf("%s/.libstorage", gotil.HomeDir())
)

// Client is the libStorage client.
type Client interface {
	apiclient.Client
}

type client struct {
	apiclient.APIClient
	config gofig.Config
}

// New returns a new Client.
func New(config gofig.Config) (Client, error) {
	if config == nil {
		if cfg, err := apiconfig.NewConfig(); err != nil {
			return nil, err
		} else {
			config = cfg
		}
	}
	ac, err := apiclient.Dial(nil, config)
	if err != nil {
		return nil, err
	}
	return &client{APIClient: ac, config: config}, nil
}
