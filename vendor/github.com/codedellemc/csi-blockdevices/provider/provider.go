package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
	"github.com/codedellemc/goioc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/codedellemc/csi-blockdevices/services"
)

const (
	nodeEnvVar = "X_CSI_BD_NODEONLY"
	ctlrEnvVar = "X_CSI_BDCONTROLLERONLY"
)

var (
	errServerStarted = errors.New(
		services.SpName + ": the server has been started")
	errServerStopped = errors.New(
		services.SpName + ": the server has been stopped")
)

func init() {
	goioc.Register(services.SpName, newProvider)
}

type provider struct {
	sync.Mutex
	server *grpc.Server
	closed bool
	plugin *services.StoragePlugin
}

func newProvider() interface{} {
	return &provider{}
}

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (p *provider) Serve(ctx context.Context, li net.Listener) error {
	log.WithField("name", services.SpName).Info(".Serve")
	if err := func() error {
		p.Lock()
		defer p.Unlock()
		if p.closed {
			return errServerStopped
		}
		if p.server != nil {
			return errServerStarted
		}
		p.server = grpc.NewServer(
			grpc.UnaryInterceptor(gocsi.ChainUnaryServer(
				gocsi.ServerRequestIDInjector,
				gocsi.NewServerRequestLogger(os.Stdout, os.Stderr),
				gocsi.NewServerResponseLogger(os.Stdout, os.Stderr),
				gocsi.NewServerRequestVersionValidator(services.CSIVersions),
				gocsi.ServerRequestValidator)))
		return nil
	}(); err != nil {
		return errServerStarted
	}

	p.plugin = &services.StoragePlugin{}
	p.plugin.Init()

	// Always host the Identity Service
	csi.RegisterIdentityServer(p.server, p.plugin)

	_, nodeSvc := os.LookupEnv(nodeEnvVar)
	_, ctrlSvc := os.LookupEnv(ctlrEnvVar)

	if nodeSvc && ctrlSvc {
		log.Errorf("Cannot specify both %s and %s",
			nodeEnvVar, ctlrEnvVar)
		return fmt.Errorf("Cannot specify both %s and %s",
			nodeEnvVar, ctlrEnvVar)
	}

	switch {
	case nodeSvc:
		csi.RegisterNodeServer(p.server, p.plugin)
		log.Debug("Added Node Service")
	case ctrlSvc:
		csi.RegisterControllerServer(p.server, p.plugin)
		log.Debug("Added Controller Service")
	default:
		csi.RegisterControllerServer(p.server, p.plugin)
		log.Debug("Added Controller Service")
		csi.RegisterNodeServer(p.server, p.plugin)
		log.Debug("Added Node Service")
	}

	// start the grpc server
	return p.server.Serve(li)
}

// Stop stops the gRPC server. It immediately closes all open
// connections and listeners.
// It cancels all active RPCs on the server side and the corresponding
// pending RPCs on the client side will get notified by connection
// errors.
func (p *provider) Stop(ctx context.Context) {
	if p.server == nil {
		return
	}

	p.Lock()
	defer p.Unlock()
	log.WithField("name", services.SpName).Info(".Stop")
	p.server.Stop()
	p.closed = true
}

// GracefulStop stops the gRPC server gracefully. It stops the server
// from accepting new connections and RPCs and blocks until all the
// pending RPCs are finished.
func (p *provider) GracefulStop(ctx context.Context) {
	if p.server == nil {
		return
	}

	p.Lock()
	defer p.Unlock()
	log.WithField("name", services.SpName).Info(".GracefulStop")
	p.server.GracefulStop()
	p.closed = true
}
