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

	gocontext "golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"

	"github.com/emccode/gournal"
	glogrus "github.com/emccode/gournal/logrus"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/server/services"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	apicnfg "github.com/emccode/libstorage/api/utils/config"

	// imported to load routers
	_ "github.com/emccode/libstorage/imports/routers"

	// imported to load remote storage drivers
	_ "github.com/emccode/libstorage/imports/remote"
)

var (
	servers []*server
)

type server struct {
	name         string
	adminToken   string
	ctx          types.Context
	addrs        []string
	config       gofig.Config
	servers      []*HTTPServer
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

func newServer(goCtx gocontext.Context, config gofig.Config) (*server, error) {

	adminTokenUUID, err := types.NewUUID()
	if err != nil {
		return nil, err
	}
	adminToken := adminTokenUUID.String()
	serverName := randomServerName()

	ctx := context.New(goCtx)
	ctx = ctx.WithValue(context.ServerKey, serverName)
	ctx = ctx.WithValue(context.AdminTokenKey, adminToken)

	if lvl, ok := context.GetLogLevel(ctx); ok {
		switch lvl {
		case log.DebugLevel:
			ctx = context.WithValue(
				ctx, gournal.LevelKey(),
				gournal.DebugLevel)
		case log.InfoLevel:
			ctx = context.WithValue(
				ctx, gournal.LevelKey(),
				gournal.InfoLevel)
		case log.WarnLevel:
			ctx = context.WithValue(
				ctx, gournal.LevelKey(),
				gournal.WarnLevel)
		case log.ErrorLevel:
			ctx = context.WithValue(
				ctx, gournal.LevelKey(),
				gournal.ErrorLevel)
		case log.FatalLevel:
			ctx = context.WithValue(
				ctx, gournal.LevelKey(),
				gournal.FatalLevel)
		case log.PanicLevel:
			ctx = context.WithValue(
				ctx, gournal.LevelKey(),
				gournal.PanicLevel)
		}
	}

	if logger, ok := ctx.Value(context.LoggerKey).(*log.Logger); ok {
		ctx = context.WithValue(
			ctx, gournal.AppenderKey(),
			glogrus.NewWithOptions(
				logger.Out, logger.Level, logger.Formatter))
	}

	if config == nil {
		var err error
		if config, err = apicnfg.NewConfig(); err != nil {
			return nil, err
		}
	}
	config = config.Scope(types.ConfigServer)

	s := &server{
		ctx:          ctx,
		name:         serverName,
		adminToken:   adminToken,
		config:       config,
		closeSignal:  make(chan int),
		closedSignal: make(chan int),
		closeOnce:    &sync.Once{},
	}

	if logger, ok := s.ctx.Value(context.LoggerKey).(*log.Logger); ok {
		s.PrintServerStartupHeader(logger.Out)
	} else {
		s.PrintServerStartupHeader(os.Stdout)
	}

	if lvl, err := log.ParseLevel(
		config.GetString(types.ConfigLogLevel)); err == nil {
		context.SetLogLevel(s.ctx, lvl)
	}

	logFields := log.Fields{}
	logConfig, err := utils.ParseLoggingConfig(
		config, logFields, "libstorage.server")
	if err != nil {
		return nil, err
	}

	// always update the server context's log level
	context.SetLogLevel(s.ctx, logConfig.Level)
	s.ctx.WithFields(logFields).Info("configured logging")

	s.ctx.Info("initializing server")

	if err := s.initEndpoints(s.ctx); err != nil {
		return nil, err
	}
	s.ctx.Info("initialized endpoints")

	if err := services.Init(s.ctx, s.config); err != nil {
		return nil, err
	}
	s.ctx.Info("initialized services")

	if logConfig.HTTPRequests || logConfig.HTTPResponses {
		s.logHTTPEnabled = true
		s.logHTTPRequests = logConfig.HTTPRequests
		s.logHTTPResponses = logConfig.HTTPResponses
		s.stdOut = getLogIO(logConfig.Stdout, types.ConfigLogStdout)
		s.stdErr = getLogIO(logConfig.Stderr, types.ConfigLogStderr)
	}

	s.initGlobalMiddleware()

	if err := s.initRouters(); err != nil {
		return nil, err
	}

	servers = append(servers, s)

	return s, nil
}

// Serve starts serving the configured libStorage endpoints. This function
// returns a channel on which errors are received. Reading this channel is
// also the prescribed manner for clients wishing to block until the server is
// shutdown as the error channel will be closed when the server is stopped.
func Serve(
	goCtx gocontext.Context,
	config gofig.Config) (types.Server, <-chan error, error) {

	s, err := newServer(goCtx, config)
	if err != nil {
		return nil, nil, err
	}

	errs := make(chan error, len(s.servers))
	srvErrs := make(chan error, len(s.servers))

	for _, srv := range s.servers {
		srv.srv.Handler = s.createMux(srv.ctx)
		go func(srv *HTTPServer) {
			srv.ctx.Info("api listening")
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

	if logger, ok := s.ctx.Value(context.LoggerKey).(*log.Logger); ok {
		s.PrintServerStartupFooter(logger.Out)
	} else {
		s.PrintServerStartupFooter(os.Stdout)
	}

	return s, errs, nil
}

// Name returns the name of the server.
func (s *server) Name() string {
	return s.name
}

// Addrs returns the server's configured endpoint addresses.
func (s *server) Addrs() []string {
	return s.addrs
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
			srv.ctx.Error(err)
		}
		if srv.l.Addr().Network() == "unix" {
			laddr := srv.l.Addr().String()
			srv.ctx.WithField(
				"path", laddr).Debug("removed unix socket")
			os.RemoveAll(laddr)
		}
		srv.ctx.Debug("shutdown endpoint complete")
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
