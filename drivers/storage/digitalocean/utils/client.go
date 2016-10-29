// +build !libstorage_storage_driver libstorage_storage_driver_digitalocean

package utils

import (
	"github.com/codedellemc/libstorage/api"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// Client returns a new DigitalOcean client
func Client(token string) (*godo.Client, error) {
	tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	client, err := godo.New(oauth2.NewClient(
		oauth2.NoContext, tokenSrc),
		godo.SetUserAgent(userAgent()))

	return client, err
}

func userAgent() string {
	return "libstorage/" + api.Version.SemVer
}
