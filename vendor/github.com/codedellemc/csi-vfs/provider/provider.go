package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path"
	"sync"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
	"github.com/codedellemc/gocsi/mount"
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
func New(
	opts []grpc.ServerOption,
	interceptors []grpc.UnaryServerInterceptor) ServiceProvider {

	return &provider{interceptors: interceptors, serverOpts: opts}
}

type provider struct {
	sync.Mutex
	server       *grpc.Server
	closed       bool
	service      service.Service
	interceptors []grpc.UnaryServerInterceptor
	serverOpts   []grpc.ServerOption
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

func (p *provider) newGrpcServer(
	idemp gocsi.IdempotencyProvider) *grpc.Server {

	var interceptors []grpc.UnaryServerInterceptor
	if len(p.interceptors) > 0 {
		interceptors = append(interceptors, p.interceptors...)
	}
	interceptors = append(
		interceptors, gocsi.NewIdempotentInterceptor(idemp))

	iopt := gocsi.ChainUnaryServer(interceptors...)

	var serverOpts []grpc.ServerOption
	if len(p.serverOpts) > 0 {
		serverOpts = append(serverOpts, p.serverOpts...)
	}

	serverOpts = append(serverOpts, grpc.UnaryInterceptor(iopt))

	return grpc.NewServer(serverOpts...)
}

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (p *provider) Serve(ctx context.Context, li net.Listener) error {

	var (
		bindfs  string
		dataDir string
		devDir  string
		mntDir  string
		volDir  string
		volGlob string
	)

	if c, ok := ctx.Value(ctxConfigKey).(config); ok {
		bindfs = c.GetString("csi.vfs.bindfs")
		dataDir = c.GetString("csi.vfs.data")
		devDir = c.GetString("csi.vfs.dev")
		mntDir = c.GetString("csi.vfs.mnt")
		volDir = c.GetString("csi.vfs.vol")
		volGlob = c.GetString("csi.vfs.volGlob")
	}

	if err := func() error {
		p.Lock()
		defer p.Unlock()
		if p.closed {
			return errServerStopped
		}
		if p.server != nil {
			return errServerStarted
		}

		idemp := newIdempotentProvider(dataDir, devDir, mntDir, volDir)
		p.server = p.newGrpcServer(idemp)
		return nil
	}(); err != nil {
		return errServerStarted
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

func newIdempotentProvider(
	data, dev, mnt, vol string) gocsi.IdempotencyProvider {
	p := &vfsIdemProvider{
		data: data,
		dev:  dev,
		mnt:  mnt,
		vol:  vol,
	}
	service.InitConfig(&p.data, &p.dev, &p.mnt, &p.vol, nil, nil)
	log.WithFields(map[string]interface{}{
		"data": p.data,
		"dev":  p.dev,
		"mnt":  p.mnt,
		"vol":  p.vol,
	}).Info("created new vfs idempotent provider")
	return p
}

var (
	errMissingIDKeyPath = errors.New("missing id key path")
)

type vfsIdemProvider struct {
	data string
	dev  string
	mnt  string
	vol  string
}

func (i *vfsIdemProvider) GetVolumeName(id *csi.VolumeID) (string, error) {
	volPath, ok := id.Values["path"]
	if !ok {
		return "", errMissingIDKeyPath
	}
	return path.Base(volPath), nil
}

func (i *vfsIdemProvider) GetVolumeInfo(name string) (*csi.VolumeInfo, error) {
	volPath := path.Join(i.vol, name)
	if !service.FileExists(volPath) {
		return nil, nil
	}
	return &csi.VolumeInfo{
		Id: &csi.VolumeID{
			Values: map[string]string{"path": volPath},
		},
	}, nil
}

func (i *vfsIdemProvider) IsControllerPublished(
	id *csi.VolumeID) (*csi.PublishVolumeInfo, error) {

	volPath, ok := id.Values["path"]
	if !ok {
		return nil, errMissingIDKeyPath
	}

	volName := path.Base(volPath)
	devPath := path.Join(i.dev, volName)
	minfo, err := mount.GetMounts()
	if err != nil {
		return nil, err
	}

	for _, mi := range minfo {
		if mi.Device == volPath && mi.Path == devPath {
			return &csi.PublishVolumeInfo{
				Values: map[string]string{
					"path": devPath,
				},
			}, nil
		}
	}

	return nil, nil
}

func (i *vfsIdemProvider) IsNodePublished(
	id *csi.VolumeID, targetPath string) (bool, error) {

	volPath, ok := id.Values["path"]
	if !ok {
		return false, errMissingIDKeyPath
	}
	volName := path.Base(volPath)
	mntPath := path.Join(i.mnt, volName)

	minfo, err := mount.GetMounts()
	if err != nil {
		return false, err
	}

	for _, mi := range minfo {
		if mi.Device == mntPath && mi.Path == targetPath {
			return true, nil
		}
	}

	return false, nil
}
