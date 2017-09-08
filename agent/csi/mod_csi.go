package csi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	dvol "github.com/docker/go-plugins-helpers/volume"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"

	"github.com/codedellemc/rexray/agent"
	apictx "github.com/codedellemc/rexray/libstorage/api/context"
	"github.com/codedellemc/rexray/libstorage/api/registry"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
)

const (
	modName = "csi"
)

type csiServer interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type mod struct {
	lsc    apitypes.Client
	ctx    apitypes.Context
	config gofig.Config
	name   string
	addr   string
	desc   string
	gs     *grpc.Server
	cs     *csiService
}

var (
	loadGoPluginsFunc func(context.Context, ...string) error
	separators        = regexp.MustCompile(`[ &_=+:]`)
	dashes            = regexp.MustCompile(`[\-]+`)
	illegalPath       = regexp.MustCompile(`[^[:alnum:]\~\-\./]`)
)

const configFormat = `
rexray:
  modules:
    default-csi:
      type:     csi
      desc:     The default CSI module.
      host:     %s
      disabled: false
`

func init() {
	agent.RegisterModule(modName, newModule)
	registry.RegisterConfigReg(
		"CSI",
		func(ctx apitypes.Context, r gofig.ConfigRegistration) {
			if v := os.Getenv("CSI_ENDPOINT"); v != "" {
				ctx.WithField("CSI_ENDPOINT", v).Info(
					"configuring default CSI module")
				r.SetYAML(fmt.Sprintf(configFormat, v))
			}
			r.Key(gofig.String, "", "libstorage", "",
				"csi.driver", "csiDriver", "X_CSI_DRIVER")
			r.Key(gofig.String, "", "", "",
				"csi.goplugins", "csiGoPlugins", "X_CSI_GO_PLUGINS")
		})
}

func newModule(
	ctx apitypes.Context,
	c *agent.Config) (agent.Module, error) {

	host := strings.Trim(c.Address, " ")

	if host == "" {
		return nil, errors.New("error: host is required")
	}

	c.Address = host
	config := c.Config

	m := &mod{
		ctx:    ctx,
		config: config,
		lsc:    c.Client,
		name:   c.Name,
		desc:   c.Description,
		addr:   host,
	}

	// Determine what kind of driver this CSI module uses.
	csiDriver := config.GetString("csi.driver")
	ctx.WithFields(map[string]interface{}{
		"mod.name":   c.Name,
		"csi.driver": csiDriver,
	}).Info("configuring csi module's driver")

	// Create the CSI service that will answer incoming requests.
	if m.cs = newService(ctx, c.Name, csiDriver); m.cs == nil {
		return nil, fmt.Errorf("invalid csi driver: %s", csiDriver)
	}

	// Create a gRPC server used to advertise the CSI service.
	m.gs = newGrpcServer(ctx)
	csi.RegisterControllerServer(m.gs, m.cs)
	csi.RegisterIdentityServer(m.gs, m.cs)
	csi.RegisterNodeServer(m.gs, m.cs)

	return m, nil
}

func newGrpcServer(ctx apitypes.Context) *grpc.Server {
	lout := newLogger(ctx.Infof)
	lerr := newLogger(ctx.Errorf)
	return grpc.NewServer(grpc.UnaryInterceptor(gocsi.ChainUnaryServer(
		gocsi.ServerRequestIDInjector,
		gocsi.NewServerRequestLogger(lout, lerr),
		gocsi.NewServerResponseLogger(lout, lerr),
		gocsi.ServerRequestValidator)))
}

var loadGoPluginsFuncOnce sync.Once

func doLoadGoPluginsFuncOnce(
	ctx apitypes.Context,
	config gofig.Config) (err error) {

	loadGoPluginsFuncOnce.Do(func() {
		if loadGoPluginsFunc != nil {
			err = loadGoPluginsFunc(
				ctx,
				config.GetStringSlice("csi.goplugins")...)
		}
	})
	return
}

const protoUnix = "unix"

func (m *mod) Start() error {

	ctx := m.ctx

	doLoadGoPluginsFuncOnce(ctx, m.config)

	var (
		addr          string
		proto         string
		isMultiplexed bool
	)

	// If multiplexing Docker+CSI then the path to the sock file is
	// determined using the same logic as the Docker module.
	if isMultiplexed = strings.EqualFold(
		m.Address(), "rexray.sock"); isMultiplexed {

		proto = protoUnix
		addr = path.Join(
			apictx.MustPathConfig(ctx).Home,
			"/run/docker/plugins/rexray.sock")
		ctx.WithField("sockFile", addr).Info("multiplexed csi+docker endpoint")
	} else {

		var err error
		if proto, addr, err = gotil.ParseAddress(m.Address()); err != nil {
			return err
		}
		ctx.WithField("sockFile", addr).Info("csi endpoint")
	}

	// ensure the sock file directory is created & remove
	// any stale sock files with the same path
	if proto == protoUnix {
		os.MkdirAll(filepath.Dir(addr), 0755)
		os.RemoveAll(addr)
	}

	// create a listener
	l, err := net.Listen(proto, addr)
	if err != nil {
		return err
	}

	var (
		tcpm  cmux.CMux
		httpl net.Listener
		grpcl net.Listener
		http2 net.Listener
	)

	// If multiplexing Docker+CSI then create the multiplexer and the routers.
	if isMultiplexed {
		// Create a cmux object.
		tcpm = cmux.New(l)

		// Declare the match for different services required.
		httpl = tcpm.Match(cmux.HTTP1Fast())
		grpcl = tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings(
			"content-type", "application/grpc"))
		http2 = tcpm.Match(cmux.HTTP2())
	}

	go func() {
		go func() {
			if err := m.cs.Serve(ctx, nil); err != nil {
				panic(err)
			}
		}()

		// Alias the listener to use.
		ll := l

		// If multiplexing Docker+CSI then use the multiplexed router.
		if isMultiplexed {
			ll = grpcl
		}

		err := m.gs.Serve(ll)

		// If not multiplexing Docker+CSI and it's a UNIX protocol
		// then remove the sock file.
		if !isMultiplexed && proto == protoUnix {
			os.RemoveAll(addr)
		}

		if err != nil && err != grpc.ErrServerStopped {
			// If not multiplexing Docker+CSI then panic on error,
			// otherwise leave that to the multiplexer.
			if !isMultiplexed {
				panic(err)
			} else {
				ctx.WithError(err).Warn(
					"failed to start csi grpc server")
			}
		}
	}()

	// If not multiplexing Docker+CSI then nothing below is required.
	if !isMultiplexed {
		return nil
	}

	go func() {
		dh := dvol.NewHandler(&dockerVolDriver{cs: m.cs, ctx: ctx})
		go func() {
			if err := dh.Serve(httpl); err != nil {
				ctx.WithError(err).Warn(
					"failed to start http1 docker->csi proxy")
			}
		}()
		go func() {
			if err := dh.Serve(http2); err != nil {
				ctx.WithError(err).Warn(
					"failed to start http2 docker->csi proxy")
			}
		}()
	}()

	go func() {
		// Start cmux serving.
		err := tcpm.Serve()
		if proto == protoUnix {
			os.RemoveAll(addr)
		}
		if err != nil && !strings.Contains(err.Error(),
			"use of closed network connection") {
			panic(err)
		}
	}()

	return nil
}

func (m *mod) Stop() error {
	m.gs.GracefulStop()
	m.cs.GracefulStop(m.ctx)
	return nil
}

func (m *mod) Name() string {
	return m.name
}

func (m *mod) Description() string {
	return m.desc
}

func (m *mod) Address() string {
	return m.addr
}

type logger struct {
	f func(msg string, args ...interface{})
	w io.Writer
}

func newLogger(f func(msg string, args ...interface{})) *logger {
	l := &logger{f: f}
	r, w := io.Pipe()
	l.w = w
	go func() {
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			f(scan.Text())
		}
	}()
	return l
}

func (l *logger) Write(data []byte) (int, error) {
	return l.w.Write(data)
}
