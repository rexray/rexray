package csi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"

	"github.com/codedellemc/rexray/agent"
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
			r.Key(gofig.String, "", "libstorage", "", "csi.driver")
			r.Key(gofig.String, "", "", "", "csi.goplugins")
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
	m.gs = newGrpcServer(m.ctx)
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

func (m *mod) Start() error {

	doLoadGoPluginsFuncOnce(m.ctx, m.config)

	proto, addr, err := gotil.ParseAddress(m.Address())
	if err != nil {
		return err
	}

	// ensure the sock file directory is created & remove
	// any stale sock files with the same path
	if proto == "unix" {
		os.MkdirAll(filepath.Dir(addr), 0755)
		os.RemoveAll(addr)
	}

	// create a listener
	l, err := net.Listen(proto, addr)
	if err != nil {
		return err
	}

	go func() {
		go func() {
			if err := m.cs.Serve(m.ctx, nil); err != nil {
				panic(err)
			}
		}()
		err := m.gs.Serve(l)
		if proto == "unix" {
			os.RemoveAll(addr)
		}
		if err != grpc.ErrServerStopped {
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
