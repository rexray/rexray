package server

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	apicnfg "github.com/emccode/libstorage/api/utils/config"
	"github.com/emccode/libstorage/api/utils/semaphore"
)

var (
	portLock = &sync.Mutex{}
	servers  []*server
)

func start(host string, tls bool, driversAndServices ...string) (
	gofig.Config, io.Closer, <-chan error) {

	if host == "" {
		portLock.Lock()
		defer portLock.Unlock()

		port := 7979
		if !gotil.IsTCPPortAvailable(port) {
			port = gotil.RandomTCPPort()
		}
		host = fmt.Sprintf("tcp://localhost:%d", port)
	}

	config := NewConfig(host, tls, driversAndServices...)
	server, errs := serve(config)

	if server != nil {
		servers = append(servers, server)
	}

	return config, server, errs
}

func startWithConfig(config gofig.Config) (io.Closer, <-chan error) {
	server, errs := serve(config)

	if server != nil {
		servers = append(servers, server)
	}

	return server, errs
}

type server struct {
	name         string
	ctx          types.Context
	addrs        []string
	config       gofig.Config
	servers      []*HTTPServer
	services     map[string]types.StorageService
	closeSignal  chan int
	closedSignal chan int
	closeOnce    *sync.Once

	routers        []types.Router
	routeHandlers  map[string][]types.Middleware
	globalHandlers []types.Middleware

	logHTTPEnabled   bool
	logHTTPRequests  bool
	logHTTPResponses bool

	stdOut io.WriteCloser
	stdErr io.WriteCloser
}

func newServer(config gofig.Config) (*server, error) {

	if config == nil {
		var err error
		if config, err = apicnfg.NewConfig(); err != nil {
			return nil, err
		}
	}

	s := &server{
		name:         randomServerName(),
		ctx:          context.Background(),
		config:       config,
		closeSignal:  make(chan int),
		closedSignal: make(chan int),
		closeOnce:    &sync.Once{},
	}

	s.ctx = s.ctx.WithContextSID(types.CtxKeyServerName, s.name)
	s.ctx = s.ctx.WithValue(types.CtxKeyServerName, s.name)
	s.ctx.Info("initializing server")

	if err := s.initEndpoints(); err != nil {
		return nil, err
	}
	s.ctx.Info("initialized endpoints")

	if err := services.Init(s.ctx, s.config); err != nil {
		return nil, err
	}
	s.ctx.Info("initialized services")

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
	return serve(config)
}
func serve(config gofig.Config) (*server, <-chan error) {
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
			srv.Context().Info("api listening")
			if err := srv.Serve(); err != nil {
				if !strings.Contains(
					err.Error(), "use of closed network connection") {
					srvErrs <- err
				}
			}
		}(srv)
	}

	go func() {
		s.ctx.Info("waiting for err or close signal")
		select {
		case err := <-srvErrs:
			errs <- err
			s.ctx.Error("received server error")
		case <-s.closeSignal:
			s.ctx.Debug("received close signal")
		}
		close(errs)
		s.ctx.Info("closed server error channel")
		s.closedSignal <- 1
	}()

	// wait a second for all the configured endpoints to start. this isn't
	// pretty, but the underlying golang http package doesn't really provide
	// a better option
	timeout := time.NewTimer(time.Second * 1)
	<-timeout.C

	s.ctx.Info("server started")
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
	s.ctx.Info("shutting down server")

	for _, srv := range s.servers {
		srv.ctx.Info("shutting down endpoint")
		if err := srv.Close(); err != nil {
			srv.Context().Error(err)
		}
		if srv.l.Addr().Network() == "unix" {
			laddr := srv.l.Addr().String()
			srv.Context().WithField(
				"path", laddr).Debug("removed unix socket")
			os.RemoveAll(laddr)
		}
		srv.Context().Debug("shutdown endpoint complete")
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

	s.ctx.Debug("shutdown server complete")

	return nil
}

// CloseOnAbort is a helper function that can be called by programs, such as
// tests or a command line or service application.
func CloseOnAbort() {
	// make sure all servers get closed even if the test is abrubptly aborted
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		fmt.Println("received abort signal")
		for range Close() {
		}
		semaphore.Unlink(types.LSX)
		os.Exit(1)
	}()
}

// Close closes all servers. This function can be used when a calling program
// traps UNIX signals or when it exits gracefully.
func Close() <-chan error {
	errs := make(chan error)
	go func() {
		for _, server := range servers {
			if err := server.Close(); err != nil {
				errs <- err
			}
		}
		close(errs)
		log.Info("all servers closed")
	}()
	return errs
}
