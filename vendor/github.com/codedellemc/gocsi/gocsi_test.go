package gocsi_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	gomegaTypes "github.com/onsi/gomega/types"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

const (
	mockPkg    = "github.com/codedellemc/gocsi/mock"
	pluginName = "mock"
)

var (
	mockBinPath = os.Getenv("GOCSI_MOCK")
	csiEndpoint = os.Getenv(gocsi.CSIEndpoint)
)

func init() {
	// Do not worry about initializing the properties used to
	// start a mock server if CSI_ENDPOINT is already defined
	if csiEndpoint != "" {
		return
	}
	if mockBinPath == "" {
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		mockBinPath = path.Join(path.Dir(exe), "mock")
		out, err := exec.Command(
			"go", "build", "-o", mockBinPath, mockPkg).CombinedOutput()
		if err != nil {
			panic(fmt.Errorf("error: build mock failed: %v\n%v",
				err, string(out)))
		}
	}
	if _, err := os.Stat(mockBinPath); err != nil {
		panic(err)
	}
}

func startMockServer(
	ctx context.Context) (*grpc.ClientConn, func(), error) {

	return startMockServerWithOptions(ctx, false)
}

func startMockServerWithOptions(
	ctx context.Context,
	impliedSockFile bool) (*grpc.ClientConn, func(), error) {

	// If CSI_ENDPOINT was defined then use the external server.
	if csiEndpoint != "" {
		fmt.Fprintf(GinkgoWriter, "created csi client: %s\n", csiEndpoint)
		client, err := newGrpcClient(ctx, csiEndpoint)
		if err != nil {
			panic(err)
		}
		return client, func() {}, nil
	}

	f, _ := ioutil.TempFile("", "")
	sockFile := f.Name()
	Ω(f.Close()).ShouldNot(HaveOccurred())
	Ω(os.RemoveAll(sockFile)).ShouldNot(HaveOccurred())
	endpoint := sockFile
	if !impliedSockFile {
		endpoint = fmt.Sprintf("unix://%s", endpoint)
	}
	fmt.Fprintf(GinkgoWriter, "determined csi endpoint: %s\n", endpoint)

	cmd := exec.Command(mockBinPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("%s=%s", gocsi.CSIEndpoint, endpoint))

	var (
		err      error
		stdout   io.Reader
		stderr   io.Reader
		debug, _ = strconv.ParseBool(os.Getenv("GOCSI_TEST_DEBUG"))
	)

	stdout, err = cmd.StdoutPipe()
	Ω(err).ShouldNot(HaveOccurred())

	if debug {
		stderr, err = cmd.StderrPipe()
		Ω(err).ShouldNot(HaveOccurred())
	}

	if err = cmd.Start(); err != nil {
		return nil, nil, err
	}

	chServed := make(chan bool)

	go func() {
		scan := bufio.NewScanner(stdout)
		for scan.Scan() {
			l := scan.Text()
			if strings.Contains(l, pluginName+".Serve:") {
				chServed <- true
			}
			if debug {
				fmt.Fprintf(os.Stdout, l)
			}
		}
		close(chServed)
	}()

	if debug {
		go func() {
			scan := bufio.NewScanner(stderr)
			for scan.Scan() {
				os.Stderr.Write(scan.Bytes())
			}
		}()
	}

	started := <-chServed

	if !started {
		return nil, nil, cmd.Wait()
	}

	client, err := newGrpcClient(ctx, endpoint)
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "error creating csi client: %v", err)
		return nil, nil, cmd.Wait()
	}
	fmt.Fprintf(GinkgoWriter, "created csi client: %s\n", endpoint)

	stopMock := func() {
		Ω(cmd.Process.Signal(os.Interrupt)).ShouldNot(HaveOccurred())
		Ω(cmd.Wait()).ShouldNot(HaveOccurred())
		os.RemoveAll(sockFile)
	}

	return client, stopMock, nil
}

func newCSIVersion(major, minor, patch uint32) *csi.Version {
	return &csi.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

var mockSupportedVersions = []*csi.Version{
	newCSIVersion(0, 1, 0),
	newCSIVersion(0, 2, 0),
	newCSIVersion(1, 0, 0),
	newCSIVersion(1, 1, 0),
}

// CTest is an alias to retrieve the current Ginko test description.
var CTest = ginkgo.CurrentGinkgoTestDescription

type gocsiErrMatcher struct {
	exp *gocsi.Error
}

func (m *gocsiErrMatcher) Match(actual interface{}) (bool, error) {
	act, ok := actual.(*gocsi.Error)
	if !ok {
		return false, errors.New("gocsiErrMatcher expects a *gocsi.Error")
	}
	if m.exp.Code != act.Code {
		return false, nil
	}
	if m.exp.Description != act.Description {
		return false, nil
	}
	if m.exp.FullMethod != act.FullMethod {
		return false, nil
	}
	return true, nil
}
func (m *gocsiErrMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nto be equal to\n\t%#v", actual, m.exp)
}
func (m *gocsiErrMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to be equal to\n\t%#v", actual, m.exp)
}

// Σ is a custom Ginkgo matcher that compares two GoCSI errors.
func Σ(a *gocsi.Error) gomegaTypes.GomegaMatcher {
	return &gocsiErrMatcher{exp: a}
}

func newGrpcClient(
	ctx context.Context,
	endpoint string) (*grpc.ClientConn, error) {

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(gocsi.ChainUnaryClient(
			gocsi.ClientCheckReponseError,
			gocsi.ClientResponseValidator)),
		grpc.WithDialer(
			func(target string, timeout time.Duration) (net.Conn, error) {
				proto, addr, err := gocsi.ParseProtoAddr(target)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout(proto, addr, timeout)
			}),
	}

	return grpc.DialContext(ctx, endpoint, dialOpts...)
}
