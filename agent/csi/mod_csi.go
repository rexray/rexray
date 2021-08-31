package csi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	gofig "github.com/akutz/gofig/types"
	dvol "github.com/docker/go-plugins-helpers/volume"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"

	"github.com/AVENTER-UG/rexray/agent"
	"github.com/AVENTER-UG/rexray/core"
	apictx "github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	rrutils "github.com/AVENTER-UG/rexray/util"
)

type csiServer interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type mod struct {
	lsc           apitypes.Client
	ctx           apitypes.Context
	config        gofig.Config
	name          string
	addr          string
	desc          string
	gs            *grpc.Server
	cs            *csiService
	lis           net.Listener
	cancel        context.CancelFunc
	waitForCancel sync.WaitGroup
}

var loadGoPluginsFunc func(context.Context, ...string) error

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
	agent.RegisterModule("csi", newModule)
	agent.RegisterModule("docker", newModule)

	registry.RegisterConfigReg(
		"CSI",
		func(ctx apitypes.Context, r gofig.ConfigRegistration) {

			pathConfig := apictx.MustPathConfig(ctx)

			csiEndpoint := os.Getenv("CSI_ENDPOINT")
			if csiEndpoint == "" {
				csiEndpoint = path.Join(
					pathConfig.Home,
					"/run/docker/plugins/rexray.sock")
			}

			if csiEndpoint != "" {
				// Register the default CSI module.
				r.SetYAML(fmt.Sprintf(configFormat, csiEndpoint))
				ctx.WithField("CSI_ENDPOINT", csiEndpoint).Info(
					"configured default CSI module")
			}

			// Register the CSI module's configuration properties.
			r.Key(gofig.String, "", "", "", "csi.endpoint")
			r.Key(gofig.String, "", "libstorage", "",
				"csi.driver", "csiDriver", "X_CSI_DRIVER")
			r.Key(gofig.String, "", "", "",
				"csi.goplugins", "csiGoPlugins", "X_CSI_GO_PLUGINS")
			r.Key(gofig.Bool, "", false, "",
				"csi.nodocker", "csiNoDocker", "X_CSI_NO_DOCKER")
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
	lout := rrutils.NewWriterFor(ctx.Infof)
	lerr := rrutils.NewWriterFor(ctx.Errorf)
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

	// Create the cancellation context for this module.
	m.ctx, m.cancel = apictx.WithCancel(m.ctx)
	ctx := m.ctx

	doLoadGoPluginsFuncOnce(ctx, m.config)

	// Check to see if the provided address is the same as that of
	// the default docker module's address. If so then multiplex
	// both Docker *and* CSI connections.
	isMultiplexed := !m.config.GetBool("csi.nodocker")

	// Use the GoCSI package to parse the address since its parsing
	// function will handle an implied UNIX sock file by virtue of a
	// vanilla file path.
	proto, addr, err := gocsi.ParseProtoAddr(m.Address())
	if err != nil {
		return err
	}

	if isMultiplexed {
		ctx.WithField("sockFile", addr).Info("multiplexed csi+docker endpoint")
	} else {
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
	m.lis = l

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

		m.waitForCancel.Add(3)
		go func() {
			<-ctx.Done()

			httpl.Close()
			grpcl.Close()
			http2.Close()

			m.waitForCancel.Done()
			m.waitForCancel.Done()
			m.waitForCancel.Done()
		}()
	}

	// Start the CSI endpoint
	go func() {
		go func() {
			if err := m.cs.Serve(ctx, nil); err != nil {
				if err.Error() != http.ErrServerClosed.Error() {
					panic(err)
				}
			}
		}()

		// Alias the listener to use.
		ll := l

		// If multiplexing Docker+CSI then use the multiplexed router.
		if isMultiplexed {
			ll = grpcl
		}

		err := m.gs.Serve(ll)
		if err != nil && err != grpc.ErrServerStopped {
			// If not multiplexing Docker+CSI then panic on error,
			// otherwise leave that to the multiplexer.
			if !isMultiplexed {
				if !strings.Contains(err.Error(),
					"use of closed network connection") {
					panic(err)
				}
			} else if !strings.Contains(err.Error(),
				"mux: listener closed") {
				ctx.WithError(err).Warn(
					"failed to start csi grpc server")
			}
		}
	}()

	// If not multiplexing Docker+CSI then nothing below is required.
	if !isMultiplexed {
		return nil
	}

	// Determine the type of Docker driver to create: bridge or legacy.
	var (
		dockerVolDriver       dvol.Driver
		warmDockerBridgeCache func()
		useDockerBridge       = !core.DockerLegacyMode
	)

	// If bridge mode is enabled then check if it should be disabled for
	// the following libStorage storage platforms:
	//
	// * EFS
	// * Isilon
	// * S3FS
	if useDockerBridge {

		// If the CSI service type is libStorage then check if legacy
		// mode is required.
		if strings.EqualFold(m.cs.serviceType, "libstorage") {
			svcName, ok := apictx.ServiceName(ctx)
			if !ok {
				return errors.New("missing service name")
			}
			svcInfo, err := m.lsc.API().ServiceInspect(ctx, svcName)
			if err != nil {
				err = fmt.Errorf(
					"mod-csi: docker mode: "+
						"libStorage.ServiceInspect: %s: failed: %v",
					svcName, err)
				ctx.Error(err)
				return err
			}

			ctx.Infof("mod-csi: docker mode: service=%s", svcName)

			if svcInfo.Driver == nil {
				err = fmt.Errorf(
					"mod-csi: docker mode: service: %s: nil driver", svcName)
				ctx.Error(err)
				return err
			}

			ctx.Debugf("mod-csi: docker mode: driver=%s", svcInfo.Driver.Name)

			// If the libStorage driver is EFS, Isilon, or S3FS then
			// Docker Bridge mode is not supported.
			if ok, _ := regexp.MatchString(
				svcInfo.Driver.Name, `(?i)^efs|isilon|s3fs$`); ok {
				useDockerBridge = false
			}
		}
	}

	// Create a Docker driver either with the Docker Bridge or the
	// Docker Legacy driver.
	{
		var err error
		if useDockerBridge {
			ctx.Info("mod-csi: docker mode: bridge")

			dockerVolDriver, err = newDockerBridge(ctx, m.config, m.cs)
			if err != nil {
				return err
			}

			// Add one for the docker cache list call.
			m.waitForCancel.Add(1)

			// Define the function used to warm the Docker Bridge's
			// volume cache.
			warmDockerBridgeCache = func() {
				defer m.waitForCancel.Done()

				// Loop every one second until a successful attempt
				// at listing the volumes using the bridge. This caches
				// the volume name-to-ID mappings.
				for {
					if _, err := dockerVolDriver.List(); err == nil {
						ctx.Info("mod-csi: docker bridge: warmed cache")
						break
					}
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Duration(1) * time.Second):
					}
				}
			}
		} else {
			ctx.Info("mod-csi: docker mode: legacy")

			dockerVolDriver, err = newDockerLegacy(ctx, m.config, m.lsc)
			if err != nil {
				return err
			}
		}
	}

	// Start the Docker Volume API
	go func() {

		var warnMsgDType string
		if useDockerBridge {
			go warmDockerBridgeCache()
			warnMsgDType = "bridge"
		} else {
			warnMsgDType = "legacy"
		}

		dh := dvol.NewHandler(dockerVolDriver)
		go func() {
			if err := dh.Serve(httpl); err != nil {
				if !strings.Contains(err.Error(),
					"mux: listener closed") {
					ctx.Warnf("mod-csi: failed to start http1 %s: %v",
						warnMsgDType, err)
				}
			}
		}()
		go func() {
			if err := dh.Serve(http2); err != nil {
				if !strings.Contains(err.Error(),
					"mux: listener closed") {
					ctx.Warnf("mod-csi: failed to start http2 %s: %v",
						warnMsgDType, err)
				}
			}
		}()
	}()

	// Start multiplexing connections to either the CSI endpoint or
	// Docker Volume API
	go func() {
		// Start cmux serving.
		err := tcpm.Serve()
		if err != nil && !strings.Contains(err.Error(),
			"use of closed network connection") {
			panic(err)
		}
	}()

	return nil
}

func (m *mod) Stop() error {
	// Invoke the module's context cancellation function and wait
	// for its participants to finish their business.
	m.cancel()
	m.waitForCancel.Wait()

	m.gs.GracefulStop()
	m.cs.GracefulStop(m.ctx)
	if m.lis != nil {
		addr := m.lis.Addr()
		if addr.Network() == protoUnix {
			os.RemoveAll(addr.String())
		}
	}
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
