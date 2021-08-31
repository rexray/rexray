// +build !agent
// +build !client

package util

import (
	"bytes"
	"fmt"
	"os"

	gofig "github.com/akutz/gofig/types"
	apiserver "github.com/AVENTER-UG/rexray/libstorage/api/server"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const defaultServiceConfigFormat = `
rexray:
  libstorage:
    service: %[1]s
    server:
      services:
        %[1]s:
          driver: %[1]s
`

// initDefaultLibStorageServices initializes the config object with a default
// libStorage service if one is not present.
//
// TODO Move this into libStorage in libStorage 0.1.2
func initDefaultLibStorageServices(
	ctx apitypes.Context, config gofig.Config) error {

	if config.IsSet(apitypes.ConfigServices) {
		ctx.Debug(
			"libStorage auto service mode disabled; services defined")
		return nil
	}

	serviceName := config.GetString(apitypes.ConfigService)
	if serviceName == "" {
		ctx.Debug(
			"libStorage auto service mode disabled; service name empty")
		return nil
	}

	ctx.WithField("driver", serviceName).Info(
		"libStorage auto service mode enabled")

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, defaultServiceConfigFormat, serviceName)

	if err := config.ReadConfig(buf); err != nil {
		return err
	}

	return nil
}

func waitUntilLibStorageStopped(ctx apitypes.Context, errs <-chan error) {
	ctx.Debug("waiting until libStorage is stopped")

	// if there is no err channel then do not wait until libStorage is stopped
	// as the absence of the err channel means libStorage was not started in
	// embedded mode
	if errs == nil {
		ctx.Debug("done waiting on err chan; err chan is nil")
		return
	}

	// in a goroutine, range over the apiserver.Close channel until it's closed
	for range apiserver.Close() {
	}
	ctx.Debug("done sending close signals to libStorage")

	// block until the err channel is closed
	for err := range errs {
		if err == nil {
			continue
		}
		ctx.WithError(err).Error("error on closing libStorage server")
	}
	ctx.Debug("done waiting on err chan")
}

func activateLibStorage(
	ctx apitypes.Context,
	config gofig.Config) (apitypes.Context, gofig.Config, <-chan error, error) {

	apiserver.DisableStartupInfo = true

	var (
		host          string
		err           error
		isRunning     bool
		errs          chan error
		serverErrChan <-chan error
		server        apitypes.Server
	)

	if host = config.GetString(apitypes.ConfigHost); host != "" {
		if !config.GetBool(apitypes.ConfigEmbedded) {
			ctx.WithField(
				"host", host,
			).Debug("not starting embedded server; embedded mode disabled")
			return ctx, config, nil, nil
		}
	}

	if host, isRunning = IsLocalServerActive(ctx, config); isRunning {
		ctx = setHost(ctx, config, host)
		ctx.WithField("host", host).Debug(
			"not starting embedded server; already running")
		return ctx, config, nil, nil
	}

	// if no host was specified then see if a set of default services need to
	// be initialized
	if host == "" {
		ctx.Debug("host is empty; initiliazing default services")
		if err = initDefaultLibStorageServices(ctx, config); err != nil {
			ctx.WithError(err).Error("error initializing default services")
			return ctx, config, nil, err
		}
	}

	ctx.Debug("starting embedded libStorage server")

	if server, serverErrChan, err = apiserver.Serve(ctx, config); err != nil {
		return ctx, config, nil, fmt.Errorf("libstorage: %v", err)
	}

	if host == "" {
		host = server.Addrs()[0]
		ctx.WithField("specHost", host).Debug(
			"got host from new server address server.Addrs()[0]")

		host = parseSafeHost(ctx, host)

		ctx.WithField("specHost", host).Debug(
			"got host from new server address; updated")
	}
	ctx = setHost(ctx, config, host)

	errs = make(chan error)
	go func() {
		for err := range serverErrChan {
			if err != nil {
				errs <- err
			}
		}
		if err := os.RemoveAll(SpecFilePath(ctx)); err == nil {
			logHostSpec(ctx, host, "removed spec file")
		}
		close(errs)
	}()

	// write the host to the spec file so that other rex-ray invocations can
	// find it, even if running as an embedded libStorage server
	if err := WriteSpecFile(ctx, host); err != nil {
		specFile := SpecFilePath(ctx)
		if os.IsPermission(err) {
			ctx.WithError(err).Errorf(
				"user does not have write permissions for %s", specFile)
		} else {
			ctx.WithError(err).Errorf(
				"error writing spec file at %s", specFile)
		}
		//WaitUntilLibStorageStopped(ctx, serverErrChan)
		return ctx, config, errs, err
	}
	logHostSpec(ctx, host, "created spec file")

	return ctx, config, errs, nil
}
