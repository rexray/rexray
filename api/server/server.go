package server

import (
	"crypto/tls"
	"fmt"
	"io"
	golog "log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/gorilla/mux"

	// imported to load routers
	_ "github.com/emccode/libstorage/api/server/router"

	// imported to load drivers
	_ "github.com/emccode/libstorage/drivers"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/server/httputils"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/types/drivers"
	"github.com/emccode/libstorage/api/utils"
)

type server struct {
	name         string
	ctx          context.Context
	addrs        []string
	config       gofig.Config
	servers      []*HTTPServer
	services     map[string]httputils.Service
	closeSignal  chan int
	closedSignal chan int
	closeOnce    *sync.Once

	routers        []httputils.Router
	routeHandlers  map[string][]httputils.Middleware
	globalHandlers []httputils.Middleware

	logHTTPEnabled   bool
	logHTTPRequests  bool
	logHTTPResponses bool

	stdOut io.WriteCloser
	stdErr io.WriteCloser
}

func newServer(config gofig.Config) (*server, error) {

	s := &server{
		name:         randomServerName(),
		ctx:          context.Background(),
		config:       config,
		closeSignal:  make(chan int),
		closedSignal: make(chan int),
		closeOnce:    &sync.Once{},
	}

	s.ctx = s.ctx.WithContextID("server", s.name)
	s.ctx.Log().Debug("initializing server")

	if err := s.initEndpoints(); err != nil {
		return nil, err
	}
	s.ctx.Log().Debug("initialized endpoints")

	if err := s.initServices(); err != nil {
		return nil, err
	}
	s.ctx.Log().Debug("initialized services")

	s.logHTTPEnabled = config.GetBool("libstorage.server.http.logging.enabled")
	if s.logHTTPEnabled {

		s.logHTTPRequests = config.GetBool(
			"libstorage.server.http.logging.logrequest")
		s.logHTTPResponses = config.GetBool(
			"libstorage.server.http.logging.logresponse")

		s.stdOut = getLogIO(
			"libstorage.server.http.logging.out", config)
		s.stdErr = getLogIO(
			"libstorage.server.http.logging.err", config)
	}

	s.initGlobalMiddleware()

	if err := s.initRouters(); err != nil {
		return nil, err
	}

	return s, nil
}

// Serve starts serving the configured libStorage endpoints. This function
// returns a channel on which errors are received. Reading this channel is
// also the prescribed manner for clients wishing to block until the server is
// shutdown as the error channel will be closed when the server is stopped.
func Serve(config gofig.Config) (io.Closer, <-chan error) {

	s, err := newServer(config)
	if err != nil {
		errs := make(chan error)
		go func() {
			errs <- err
			close(errs)
		}()
		return nil, errs
	}

	errs := make(chan error, len(s.servers))
	srvErrs := make(chan error, len(s.servers))

	for _, srv := range s.servers {
		srv.srv.Handler = s.createMux(srv.Context())
		go func(srv *HTTPServer) {
			srv.Context().Log().Info("api listening")
			if err := srv.Serve(); err != nil {
				if !strings.Contains(
					err.Error(), "use of closed network connection") {
					srvErrs <- err
				}
			}
		}(srv)
	}

	go func() {
		s.ctx.Log().Debugln("waiting for err or close signal")
		select {
		case err := <-srvErrs:
			errs <- err
			s.ctx.Log().Debug("received server error")
		case <-s.closeSignal:
			s.ctx.Log().Debug("received close signal")
		}
		close(errs)
		s.ctx.Log().Debugln("closed server error channel")
		s.closedSignal <- 1
	}()

	// wait a second for all the configured endpoints to start. this isn't
	// pretty, but the underlying golang http package doesn't really provide
	// a better option
	timeout := time.NewTimer(time.Second * 1)
	<-timeout.C

	s.ctx.Log().Info("server started")
	return s, errs
}

// Close closes servers and thus stop receiving requests
func (s *server) Close() (err error) {
	s.closeOnce.Do(
		func() {
			err = s.close()
			s.closeSignal <- 1
			<-s.closedSignal
		})
	return
}

func (s *server) close() error {
	s.ctx.Log().Info("shutting down server")

	for _, srv := range s.servers {
		srv.ctx.Log().Info("shutting down endpoint")
		if err := srv.Close(); err != nil {
			srv.Context().Log().Error(err)
		}
		if srv.l.Addr().Network() == "unix" {
			laddr := srv.l.Addr().String()
			srv.Context().Log().WithField(
				"path", laddr).Debug("removed unix socket")
			os.RemoveAll(laddr)
		}
		srv.Context().Log().Debug("shutdown endpoint complete")
	}

	if s.stdOut != nil {
		if err := s.stdOut.Close(); err != nil {
			log.Error(err)
		}
	}

	if s.stdErr != nil {
		if err := s.stdErr.Close(); err != nil {
			log.Error(err)
		}
	}

	s.ctx.Log().Debug("shutdown server complete")

	return nil
}

func (s *server) initEndpoints() error {

	endpointsObj := s.config.Get("libstorage.server.endpoints")
	if endpointsObj == nil {
		return goof.New("no endpoints defined")
	}

	endpoints, ok := endpointsObj.(map[string]interface{})
	if !ok {
		return goof.New("endpoints invalid type")
	}

	for endpointName := range endpoints {

		endpoint := fmt.Sprintf("libstorage.server.endpoints.%s", endpointName)
		address := fmt.Sprintf("%s.address", endpoint)

		laddr := s.config.GetString(address)
		if laddr == "" {
			return goof.WithField("endpoint", endpoint, "missing address")
		}

		proto, addr, err := gotil.ParseAddress(laddr)
		if err != nil {
			return err
		}

		logFields := map[string]interface{}{
			"endpoint": endpointName,
			"address":  laddr,
		}

		tlsConfig, tlsFields, err :=
			utils.ParseTLSConfig(s.config.Scope(endpoint))
		if err != nil {
			return err
		}
		for k, v := range tlsFields {
			logFields[k] = v
		}

		log.WithFields(logFields).Info("configured endpoint")

		srv, err := s.newHTTPServer(proto, addr, tlsConfig)
		if err != nil {
			return err
		}

		srv.Context().Log().Info("server created")
		s.servers = append(s.servers, srv)
	}

	return nil
}

type service struct {
	driver drivers.StorageDriver
	config gofig.Config
	name   string
}

func (s *service) Config() gofig.Config {
	return s.config
}

func (s *service) Driver() drivers.StorageDriver {
	return s.driver
}

func (s *service) Name() string {
	return s.name
}

func (s *server) initServices() error {

	cfgSvcs := s.config.Get("libstorage.server.services")
	cfgSvcsMap, ok := cfgSvcs.(map[string]interface{})
	if !ok {
		return goof.New("invalid format libstorage.server.services")
	}
	log.WithField("count", len(cfgSvcsMap)).Debug("got services map")

	s.services = map[string]httputils.Service{}

	for name := range cfgSvcsMap {
		name = strings.ToLower(name)

		log.WithField("name", name).Debug("processing service config")

		scope := fmt.Sprintf("libstorage.server.services.%s", name)
		log.WithField("scope", scope).Debug("getting scoped config for service")
		sc := s.config.Scope(scope)

		driverName := sc.GetString("driver")
		if driverName == "" {
			driverName = sc.GetString("libstorage.driver")
			if driverName == "" {
				return goof.WithField(
					"service", name, "error getting driver name")
			}
		}

		driver, err := registry.NewStorageDriver(driverName)
		if err != nil {
			return err
		}

		if err := driver.Init(sc); err != nil {
			return err
		}

		svc := &service{
			name:   name,
			driver: driver,
			config: sc,
		}

		log.WithFields(log.Fields{
			"service": svc.Name(),
			"driver":  driver.Name(),
		}).Info("created new service")

		s.services[name] = svc
	}

	return nil
}

func (s *server) initRouters() error {
	for r := range registry.Routers() {
		r.Init(s.config, s.services)
		s.addRouter(r)
		log.WithFields(log.Fields{
			"router":      r.Name(),
			"len(routes)": len(r.Routes()),
		}).Info("initialized router")
	}

	// now that the routers are initialized, initialize the router middleware
	s.initRouteMiddleware()

	return nil
}

func (s *server) addRouter(r httputils.Router) {
	s.routers = append(s.routers, r)
}

func (s *server) makeHTTPHandler(
	ctx context.Context,
	route httputils.Route) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		fctx := context.NewContext(ctx, req)
		fctx = ctx.WithContextID("route", route.GetName())
		fctx = ctx.WithRoute(route.GetName())

		if req.TLS != nil {
			if len(req.TLS.PeerCertificates) > 0 {
				fctx = ctx.WithContextID("user",
					req.TLS.PeerCertificates[0].Subject.CommonName)
			}
		}

		ctx.Log().Debug("http request")

		vars := mux.Vars(req)
		if vars == nil {
			vars = map[string]string{}
		}
		store := utils.NewStoreWithVars(vars)

		handlerFunc := s.handleWithMiddleware(fctx, route)
		if err := handlerFunc(fctx, w, req, store); err != nil {
			fctx.Log().Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *server) createMux(ctx context.Context) *mux.Router {
	m := mux.NewRouter()
	for _, apiRouter := range s.routers {
		for _, r := range apiRouter.Routes() {
			f := s.makeHTTPHandler(ctx, r)
			mr := m.Path(r.GetPath())
			mr = mr.Name(r.GetName())
			mr = mr.Methods(r.GetMethod())
			mr = mr.Queries(r.GetQueries()...)
			mr.Handler(f)
			if ctx.Log().Level >= log.DebugLevel {
				ctx.Log().WithFields(log.Fields{
					"route":        r.GetName(),
					"path":         r.GetPath(),
					"method":       r.GetMethod(),
					"queries":      r.GetQueries(),
					"len(queries)": len(r.GetQueries()),
				}).Debug("registered route")
			} else {
				ctx.Log().WithField(
					"route", r.GetName()).Info("registered route")
			}
		}
	}
	return m
}

func (s *server) newHTTPServer(
	proto, laddr string, tlsConfig *tls.Config) (*HTTPServer, error) {

	var (
		l   net.Listener
		err error
	)

	if tlsConfig != nil {
		l, err = tls.Listen(proto, laddr, tlsConfig)
	} else {
		l, err = net.Listen(proto, laddr)
	}

	if err != nil {
		return nil, err
	}

	host := fmt.Sprintf("%s://%s", proto, laddr)
	ctx := s.ctx
	ctx = ctx.WithContextID("host", host)
	ctx = ctx.WithContextID("tls", fmt.Sprintf("%v", tlsConfig != nil))
	errLogger := &httpServerErrLogger{ctx.Log()}

	srv := &http.Server{Addr: l.Addr().String()}
	srv.ErrorLog = golog.New(errLogger, "", 0)

	return &HTTPServer{
		srv: srv,
		l:   l,
		ctx: ctx,
	}, nil
}

// HTTPServer contains an instance of http server and the listener.
//
// srv *http.Server, contains configuration to create a http server and a mux
// router with all api end points.
//
// l   net.Listener, is a TCP or Socket listener that dispatches incoming
// request to the router.
type HTTPServer struct {
	srv *http.Server
	l   net.Listener
	ctx context.Context
}

// Serve starts listening for inbound requests.
func (s *HTTPServer) Serve() error {
	return s.srv.Serve(s.l)
}

// Close closes the HTTPServer from listening for the inbound requests.
func (s *HTTPServer) Close() error {
	return s.l.Close()
}

// Context returns this server's context.
func (s *HTTPServer) Context() context.Context {
	return s.ctx
}

func getLogIO(
	propName string,
	config gofig.Config) io.WriteCloser {

	if path := config.GetString(propName); path != "" {
		logio, err := os.OpenFile(
			path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error(err)
		}
		log.WithFields(log.Fields{
			"logType": propName,
			"logPath": path,
		}).Debug("using log file")
		return logio
	}
	return log.StandardLogger().Writer()
}

type httpServerErrLogger struct {
	logger *log.Logger
}

func (l *httpServerErrLogger) Write(p []byte) (n int, err error) {
	l.logger.Error(string(p))
	return len(p), nil
}
