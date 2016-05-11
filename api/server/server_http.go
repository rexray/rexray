package server

import (
	"crypto/tls"
	"fmt"
	"io"
	golog "log"
	"net"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/gorilla/mux"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

func (s *server) initEndpoints(ctx types.Context) error {

	endpointsObj := s.config.Get(types.ConfigEndpoints)
	if endpointsObj == nil {
		return goof.New("no endpoints defined")
	}

	endpoints, ok := endpointsObj.(map[string]interface{})
	if !ok {
		return goof.New("endpoints invalid type")
	}

	for endpointName := range endpoints {

		endpoint := fmt.Sprintf("%s.%s", types.ConfigEndpoints, endpointName)
		address := fmt.Sprintf("%s.address", endpoint)
		s.ctx.WithFields(log.Fields{
			"endpoint": endpoint, "address": address}).Debug("endpoing info")

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

		tlsConfig, err :=
			utils.ParseTLSConfig(s.config.Scope(endpoint), logFields)
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
		log.WithFields(log.Fields{
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

		ctx.Debug("http request")

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

func getLogIO(propName string, config gofig.Config) io.WriteCloser {

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
