package main

import (
	"net"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/csi-blockdevices/services"
	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

const (
	debugEnvVar = "BDPLUGIN_DEBUG"
	nodeEnvVar  = "BDPLUGIN_NODEONLY"
	ctlrEnvVar  = "BDPLUGIN_CONTROLLERONLY"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	if _, d := os.LookupEnv(debugEnvVar); d {
		log.SetLevel(log.DebugLevel)
	}

	s := &sp{}

	go func() {
		_ = <-c
		if s.server != nil {
			log.WithField("name", services.SpName).Info(".GracefulStop")
			s.server.GracefulStop()

			// make sure sock file got cleaned up
			proto, addr, _ := gocsi.GetCSIEndpoint()
			if proto == "unix" && addr != "" {
				if _, err := os.Stat(addr); !os.IsNotExist(err) {
					s.server.Stop()
					if err := os.Remove(addr); err != nil {
						log.WithError(err).Warn(
							"Unable to remove sock file")
					}
				}
			}
		}
	}()

	l, err := gocsi.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	ctx := context.Background()

	if err := s.Serve(ctx, l); err != nil {
		if err != grpc.ErrServerStopped {
			log.WithError(err).Fatal("grpc failed")
		}
	}
}

type sp struct {
	server *grpc.Server
	plugin *services.StoragePlugin
}

// ServiceProvider.Serve
func (s *sp) Serve(ctx context.Context, li net.Listener) error {
	log.WithField("name", services.SpName).Info(".Serve")
	s.server = grpc.NewServer(grpc.UnaryInterceptor(gocsi.ChainUnaryServer(
		gocsi.ServerRequestIDInjector,
		gocsi.NewServerRequestLogger(os.Stdout, os.Stderr),
		gocsi.NewServerResponseLogger(os.Stdout, os.Stderr),
		gocsi.NewServerRequestVersionValidator(services.CSIVersions),
		gocsi.ServerRequestValidator)))

	s.plugin = &services.StoragePlugin{}
	s.plugin.Init()

	// Always host the Identity Service
	csi.RegisterIdentityServer(s.server, s.plugin)

	_, nodeSvc := os.LookupEnv(nodeEnvVar)
	_, ctrlSvc := os.LookupEnv(ctlrEnvVar)

	if nodeSvc && ctrlSvc {
		log.Fatalf("Cannot specify both %s and %s",
			nodeEnvVar, ctlrEnvVar)
	}

	switch {
	case nodeSvc:
		csi.RegisterNodeServer(s.server, s.plugin)
		log.Debug("Added Node Service")
	case ctrlSvc:
		csi.RegisterControllerServer(s.server, s.plugin)
		log.Debug("Added Controller Service")
	default:
		csi.RegisterControllerServer(s.server, s.plugin)
		log.Debug("Added Controller Service")
		csi.RegisterNodeServer(s.server, s.plugin)
		log.Debug("Added Node Service")
	}

	// start the grpc server
	return s.server.Serve(li)
}
