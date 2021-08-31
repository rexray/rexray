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

	log "github.com/sirupsen/logrus"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gournal"
	glogrus "github.com/akutz/gournal/logrus"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/services"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	apicnfg "github.com/AVENTER-UG/rexray/libstorage/api/utils/config"

	// import and load the routers
	_ "github.com/AVENTER-UG/rexray/libstorage/api/server/router/help"
	_ "github.com/AVENTER-UG/rexray/libstorage/api/server/router/root"
	_ "github.com/AVENTER-UG/rexray/libstorage/api/server/router/service"
	_ "github.com/AVENTER-UG/rexray/libstorage/api/server/router/snapshot"
	_ "github.com/AVENTER-UG/rexray/libstorage/api/server/router/tasks"
	_ "github.com/AVENTER-UG/rexray/libstorage/api/server/router/volume"
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
	authConfig   *types.AuthConfig
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
		if config, err = apicnfg.NewConfig(ctx); err != nil {
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
		s.PrintServerStartupHeader(types.Stderr)
	}

	if lvl, err := log.ParseLevel(
		config.GetString(types.ConfigLogLevel)); err == nil {
		context.SetLogLevel(s.ctx, lvl)
	}

	logFields := log.Fields{}
	logConfig, err := utils.ParseLoggingConfig(
		config, logFields, types.ConfigServer)
	if err != nil {
		return nil, err
	}

	// always update the server context's log level
	context.SetLogLevel(s.ctx, logConfig.Level)
	s.ctx.WithFields(logFields).Info("configured logging")

	authFields := log.Fields{}
	authConfig, err := utils.ParseAuthConfig(
		s.ctx, config, authFields, types.ConfigServer)
	if err != nil {
		return nil, err
	}
	s.authConfig = authConfig
	if s.authConfig != nil {
		s.ctx.WithFields(authFields).Info("configured global auth")
	}

	s.ctx.Info("initializing server")

	if err := services.Init(s.ctx, s.config); err != nil {
		return nil, err
	}
	s.ctx.Info("initialized services")

	if err := s.initEndpoints(s.ctx); err != nil {
		return nil, err
	}
	s.ctx.Info("initialized endpoints")

	if logConfig.HTTPRequests || logConfig.HTTPResponses {
		s.logHTTPEnabled = true
		s.logHTTPRequests = logConfig.HTTPRequests
		s.logHTTPResponses = logConfig.HTTPResponses
		s.stdOut = getLogIO(s.ctx, logConfig.Stdout, types.ConfigLogStdout)
		s.stdErr = getLogIO(s.ctx, logConfig.Stderr, types.ConfigLogStderr)
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

	if goCtx == nil {
		goCtx = context.Background()
	}

	ctx := context.New(goCtx)

	if _, ok := context.PathConfig(ctx); !ok {
		pathConfig := utils.NewPathConfig()
		ctx = ctx.WithValue(context.PathConfigKey, pathConfig)
		registry.ProcessRegisteredConfigs(ctx)
	}

	s, err := newServer(ctx, config)
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
		s.PrintServerStartupFooter(types.Stderr)
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
