package server

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	golog "log"
	"net"
	"net/http"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/gorilla/mux"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

const defaultEndpointConfig = `
libstorage:
  server:
    endpoints:
      localhost:
        address: %s
`

func (s *server) initDefaultEndpoint() error {

	var endpointConfig string

	if host := s.config.GetString(types.ConfigHost); host != "" {

		s.ctx.WithField("host", host).Info("initializing default endpoint")
		endpointConfig = fmt.Sprintf(defaultEndpointConfig, host)

	} else {

		aem := types.ParseEndpointType(
			s.config.GetString(types.ConfigServerAutoEndpointMode))
		s.ctx.WithField("autoEndpointMode", aem).Info(
			"initializing default endpoint")
		endpointConfig = fmt.Sprintf(defaultEndpointConfig, aem)
	}

	return s.config.ReadConfig(bytes.NewReader([]byte(endpointConfig)))
}

var (
	tcpPortLock = &sync.Mutex{}
)

func (s *server) initEndpoints(ctx types.Context) error {

	endpointsObj := s.config.Get(types.ConfigEndpoints)
	if endpointsObj == nil {
		if err := s.initDefaultEndpoint(); err != nil {
			return goof.WithError("no endpoints defined", err)
		}
		endpointsObj = s.config.Get(types.ConfigEndpoints)
	}

	endpoints, ok := endpointsObj.(map[string]interface{})
	if !ok {
		return goof.New("endpoints invalid type")
	}

	if len(endpoints) == 0 {
		if err := s.initDefaultEndpoint(); err != nil {
			return err
		}
	}

	for endpointName := range endpoints {

		endpoint := fmt.Sprintf("%s.%s", types.ConfigEndpoints, endpointName)
		address := fmt.Sprintf("%s.address", endpoint)
		laddr := s.config.GetString(address)
		if laddr == "" {
			return goof.WithField("endpoint", endpoint, "missing address")
		}

		laddrET := types.ParseEndpointType(laddr)
		switch laddrET {

		case types.TCPEndpoint:

			var tcpPort int
			func() {
				tcpPortLock.Lock()
				defer tcpPortLock.Unlock()
				tcpPort = gotil.RandomTCPPort()
			}()

			laddr = fmt.Sprintf("tcp://127.0.0.1:%d", tcpPort)
			s.ctx.WithField("endpoint", endpoint).Info(
				"initializing auto tcp endpoint")

		case types.UnixEndpoint:

			laddr = fmt.Sprintf("unix://%s", utils.GetTempSockFile(ctx))
			s.ctx.WithField("endpoint", endpoint).Info(
				"initializing auto unix endpoint")

		}

		s.ctx.WithFields(log.Fields{
			"endpoint": endpoint, "address": laddr}).Debug("endpoint info")

		s.addrs = append(s.addrs, laddr)

		proto, addr, err := gotil.ParseAddress(laddr)
		if err != nil {
			return err
		}

		logFields := map[string]interface{}{
			"endpoint": endpointName,
			"address":  laddr,
		}

		tlsConfig, err := utils.ParseTLSConfig(
			s.ctx, s.config, proto, logFields, types.ConfigServer)
		if err != nil {
			return err
		}

		ctx.WithFields(logFields).Info("configured endpoint")

		srv, err := s.newHTTPServer(proto, addr, tlsConfig)
		if err != nil {
			return err
		}

		ctx.Info("server created")
		s.servers = append(s.servers, srv)
	}

	return nil
}

func (s *server) initRouters() error {
	for r := range registry.Routers() {
		r.Init(s.config)
		s.addRouter(r)
		s.ctx.WithFields(log.Fields{
			"router":      r.Name(),
			"len(routes)": len(r.Routes()),
		}).Info("initialized router")
	}

	// now that the routers are initialized, initialize the router middleware
	s.initRouteMiddleware()

	return nil
}

func (s *server) addRouter(r types.Router) {
	s.routers = append(s.routers, r)
}

func (s *server) makeHTTPHandler(
	ctx types.Context,
	route types.Route) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set(types.ServerNameHeader, s.name)

		ctx := context.WithRequestRoute(ctx, req, route)

		if req.TLS != nil {
			if len(req.TLS.PeerCertificates) > 0 {
				userName := req.TLS.PeerCertificates[0].Subject.CommonName
				ctx = ctx.WithValue(context.UserKey, userName)
			}
		}

		ctx.Info("http request")

		vars := mux.Vars(req)
		if vars == nil {
			vars = map[string]string{}
		}
		store := utils.NewStoreWithVars(vars)

		handlerFunc := s.handleWithMiddleware(ctx, route)
		if err := handlerFunc(ctx, w, req, store); err != nil {
			ctx.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *server) createMux(ctx types.Context) *mux.Router {
	m := mux.NewRouter()
	for _, apiRouter := range s.routers {
		for _, r := range apiRouter.Routes() {

			ctx := ctx.WithValue(context.RouteKey, r)

			f := s.makeHTTPHandler(ctx, r)
			mr := m.Path(r.GetPath())
			mr = mr.Name(r.GetName())
			mr = mr.Methods(r.GetMethod())
			mr = mr.Queries(r.GetQueries()...)
			mr.Handler(f)

			if l, ok := context.GetLogLevel(ctx); ok && l >= log.DebugLevel {
				ctx.WithFields(log.Fields{
					"path":         r.GetPath(),
					"method":       r.GetMethod(),
					"queries":      r.GetQueries(),
					"len(queries)": len(r.GetQueries()),
				}).Debug("registered route")
			} else {
				ctx.Info("registered route")
			}
		}
	}
	return m
}

func (s *server) newHTTPServer(
	proto, laddr string, tlsConfig *types.TLSConfig) (*HTTPServer, error) {

	var (
		l   net.Listener
		err error
	)

	if tlsConfig != nil {
		l, err = tls.Listen(proto, laddr, &tlsConfig.Config)
	} else {
		l, err = net.Listen(proto, laddr)
	}

	if err != nil {
		return nil, err
	}

	host := fmt.Sprintf("%s://%s", proto, laddr)
	ctx := s.ctx.WithValue(context.HostKey, host)
	ctx = ctx.WithValue(context.TLSKey, tlsConfig != nil)

	logger := ctx.Value(context.LoggerKey).(*log.Logger)
	errLogger := &httpServerErrLogger{logger}

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
	ctx types.Context
}

// Serve starts listening for inbound requests.
func (s *HTTPServer) Serve() error {
	return s.srv.Serve(s.l)
}

// Close closes the HTTPServer from listening for the inbound requests.
func (s *HTTPServer) Close() error {
	return s.l.Close()
}

// Context returns this server's types.
func (s *HTTPServer) Context() types.Context {
	return s.ctx
}

func getLogIO(ctx types.Context, path, propName string) io.WriteCloser {

	if path != "" {
		logio, err := os.OpenFile(
			path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			ctx.WithError(err).Error("error opening log file")
		}
		ctx.WithFields(log.Fields{
			"logType": propName,
			"logPath": path,
		}).Debug("using log file")
		return logio
	}
	if logger, ok := ctx.Value(context.LoggerKey).(*log.Logger); ok {
		return logger.Writer()
	}
	if propName == types.ConfigLogStdout {
		return &writeCloser{types.Stdout}
	}
	return &writeCloser{types.Stderr}
}

type httpServerErrLogger struct {
	logger *log.Logger
}

func (l *httpServerErrLogger) Write(p []byte) (n int, err error) {
	l.logger.Error(string(p))
	return len(p), nil
}

type writeCloser struct {
	writer io.Writer
}

func (wc *writeCloser) Write(p []byte) (n int, err error) {
	return wc.writer.Write(p)
}

func (wc *writeCloser) Close() error {
	if c, ok := wc.writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
