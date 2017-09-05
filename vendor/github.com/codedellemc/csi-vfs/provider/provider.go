package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/codedellemc/gocsi/csi"
	"github.com/codedellemc/goioc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/codedellemc/csi-vfs/service"
)

var (
	errServerStopped = errors.New("server stopped")
	errServerStarted = errors.New("server started")
	csiVFSRootCtxKey = interface{}("csi.vfs.root")
)

// ServiceProvider is a gRPC endpoint that provides the CSI
// services: Controller, Identity, Node.
type ServiceProvider interface {

	// Serve accepts incoming connections on the listener lis, creating
	// a new ServerTransport and service goroutine for each. The service
	// goroutine read gRPC requests and then call the registered handlers
	// to reply to them. Serve returns when lis.Accept fails with fatal
	// errors.  lis will be closed when this method returns.
	// Serve always returns non-nil error.
	Serve(ctx context.Context, lis net.Listener) error

	// Stop stops the gRPC server. It immediately closes all open
	// connections and listeners.
	// It cancels all active RPCs on the server side and the corresponding
	// pending RPCs on the client side will get notified by connection
	// errors.
	Stop(ctx context.Context)

	// GracefulStop stops the gRPC server gracefully. It stops the server
	// from accepting new connections and RPCs and blocks until all the
	// pending RPCs are finished.
	GracefulStop(ctx context.Context)
}

func init() {
	goioc.Register(service.Name, func() interface{} { return &provider{} })
}

// New returns a new service provider.
func New(opts ...grpc.ServerOption) ServiceProvider {
	return &provider{serverOpts: opts}
}

type provider struct {
	sync.Mutex
	server     *grpc.Server
	closed     bool
	service    service.Service
	serverOpts []grpc.ServerOption
}

// config is an interface that matches a possible config object that
// could possibly be pulled out of the context given to the provider's
// Serve function
type config interface {
	GetString(key string) string
}

// ctxConfigKey is an interface-wrapped key used to access a possible
// config object in the context given to the provider's Serve function
var ctxConfigKey = interface{}("csi.config")

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (p *provider) Serve(ctx context.Context, li net.Listener) error {
	if err := func() error {
		p.Lock()
		defer p.Unlock()
		if p.closed {
			return errServerStopped
		}
		if p.server != nil {
			return errServerStarted
		}
		p.server = grpc.NewServer(p.serverOpts...)
		return nil
	}(); err != nil {
		return errServerStarted
	}

	var (
		bindfs  string
		dataDir string
		devDir  string
		mntDir  string
		volDir  string
		volGlob string
	)

	if c, ok := ctx.Value(ctxConfigKey).(config); ok {
		bindfs = c.GetString("csi.bindfs")
		dataDir = c.GetString("csi.data")
		devDir = c.GetString("csi.dev")
		mntDir = c.GetString("csi.mnt")
		volDir = c.GetString("csi.vol")
		volGlob = c.GetString("csi.volGlob")
	}

	p.service = service.New(dataDir, devDir, mntDir, volDir, volGlob, bindfs)

	// Register the services.
	csi.RegisterControllerServer(p.server, p.service)
	csi.RegisterIdentityServer(p.server, p.service)
	csi.RegisterNodeServer(p.server, p.service)

	// Start the grpc server
	log.WithFields(map[string]interface{}{
		"service": service.Name,
		"address": fmt.Sprintf(
			"%s://%s", li.Addr().Network(), li.Addr().String()),
	}).Info("serving")
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
	p.server.Stop()
	p.closed = true
	log.WithField("service", service.Name).Info("stopped")
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
	p.server.GracefulStop()
	p.closed = true
	log.WithField("service", service.Name).Info("shutdown")
}
