package main

import (
	"bufio"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"golang.org/x/net/context"

	"github.com/codedellemc/csi-vfs/provider"
	"github.com/codedellemc/csi-vfs/service"
)

////////////////////////////////////////////////////////////////////////////////
//                                 CLI                                        //
////////////////////////////////////////////////////////////////////////////////

// main is ignored when this package is built as a go plug-in
func main() {
	l, err := gocsi.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Fatalln("failed to listen")
	}

	// Define a lambda that can be used in the exit handler
	// to remove a potential UNIX sock file.
	rmSockFile := func() {
		if l == nil || l.Addr() == nil {
			return
		}
		if l.Addr().Network() == "unix" {
			sockFile := l.Addr().String()
			os.RemoveAll(sockFile)
			log.WithField("path", sockFile).Info("removed sock file")
		}
	}

	sp := provider.New(newGrpcInterceptorChain())
	ctx := context.Background()

	trapSignals(func() {
		sp.GracefulStop(ctx)
		rmSockFile()
		log.Info("server stopped gracefully")
	}, func() {
		sp.Stop(ctx)
		rmSockFile()
		log.Info("server aborted")
	})

	if err := sp.Serve(ctx, l); err != nil {
		log.WithError(err).Fatal("grpc failed")
		os.Exit(1)
	}
}

func newGrpcInterceptorChain() grpc.ServerOption {
	lout := newLogger(log.Infof)
	lerr := newLogger(log.Errorf)
	return grpc.UnaryInterceptor(gocsi.ChainUnaryServer(
		gocsi.ServerRequestIDInjector,
		gocsi.NewServerRequestLogger(lout, lerr),
		gocsi.NewServerResponseLogger(lout, lerr),
		gocsi.NewServerRequestVersionValidator(service.SupportedVersions),
		gocsi.ServerRequestValidator))
}

type serviceProvider interface {
	Serve(ctx context.Context, lis net.Listener) error
	Stop(ctx context.Context)
	GracefulStop(ctx context.Context)
}

func trapSignals(onExit, onAbort func()) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc)
	go func() {
		for s := range sigc {
			ok, graceful := isExitSignal(s)
			if !ok {
				continue
			}
			if !graceful {
				log.WithField("signal", s).Error("received signal; aborting")
				if onAbort != nil {
					onAbort()
				}
				os.Exit(1)
			}
			log.WithField("signal", s).Info("received signal; shutting down")
			if onExit != nil {
				onExit()
			}
			os.Exit(0)
		}
	}()
}

// isExitSignal returns a flag indicating whether a signal is SIGKILL, SIGHUP,
// SIGINT, SIGTERM, or SIGQUIT. The second return value is whether it is a
// graceful exit. This flag is true for SIGTERM, SIGHUP, SIGINT, and SIGQUIT.
func isExitSignal(s os.Signal) (bool, bool) {
	switch s {
	case syscall.SIGKILL:
		return true, false
	case syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT:
		return true, true
	default:
		return false, false
	}
}

type logger struct {
	f func(msg string, args ...interface{})
	w io.Writer
}

func newLogger(f func(msg string, args ...interface{})) *logger {
	l := &logger{f: f}
	r, w := io.Pipe()
	l.w = w
	go func() {
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			f(scan.Text())
		}
	}()
	return l
}

func (l *logger) Write(data []byte) (int, error) {
	return l.w.Write(data)
}
