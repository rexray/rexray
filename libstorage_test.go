package libstorage

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/client"
)

func TestMain(m *testing.M) {

	// make sure all servers get closed even if the test is abrubptly aborted
	trapAbort()

	os.MkdirAll(clientToolDir, 0755)
	ioutil.WriteFile(localDevicesFile, localDevicesFileBuf, 0644)

	if debug, _ := strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG")); debug {
		log.SetLevel(log.DebugLevel)
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_SERVER_HTTP_LOGGING_LOGRESPONSE", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGRESPONSE", "true")
	}

	exitCode := m.Run()

	if !closeAllServers() && exitCode == 0 {
		exitCode = 1
	}
	os.RemoveAll(clientToolDir)
	os.RemoveAll(localDevicesFile)
	os.Exit(exitCode)
}

func trapAbort() {
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
		closeAllServers()
		fmt.Println("all servers closed")
		os.Exit(1)
	}()
}

const (
	localDevicesFile = "/tmp/libstorage/partitions"
	clientToolDir    = "/tmp/libstorage/bin"

	testServer1Name = "testService1"
	testServer2Name = "testService2"
	testServer3Name = "testService3"
)

var (
	tlsPath = fmt.Sprintf(
		"%s/src/github.com/emccode/libstorage/.tls", os.Getenv("GOPATH"))
	serverCrt    = fmt.Sprintf("%s/libstorage-server.crt", tlsPath)
	serverKey    = fmt.Sprintf("%s/libstorage-server.key", tlsPath)
	clientCrt    = fmt.Sprintf("%s/libstorage-client.crt", tlsPath)
	clientKey    = fmt.Sprintf("%s/libstorage-client.key", tlsPath)
	trustedCerts = fmt.Sprintf("%s/libstorage-ca.crt", tlsPath)
)

var localDevicesFileBuf = []byte(`
major minor  #blocks  name
  11        0    4050944 sr0
   8        0   67108864 sda
   8        1     512000 sda1
   8        2   66595840 sda2
 253        0    4079616 dm-0
 253        1   42004480 dm-1
 253        2   20508672 dm-2
 1024       1   20508672 xvda
   7        0  104857600 loop0
   7        1    2097152 loop1
 253        3  104857600 dm-3
`)

var servers []io.Closer

func closeAllServers() bool {
	noErrors := true
	for _, server := range servers {
		if err := server.Close(); err != nil {
			noErrors = false
			fmt.Printf("error closing server: %v\n", err)
		}
	}
	return noErrors
}

func getServer(
	host string, tls bool, t *testing.T) (gofig.Config, io.Closer, <-chan error) {

	if host == "" {
		host = fmt.Sprintf("tcp://localhost:%d", gotil.RandomTCPPort())
	}
	config := getConfig(host, tls, t)
	server, errs := Serve(config)
	if server != nil {
		servers = append(servers, server)
	}
	return config, server, errs
}

func getClient(host string, tls bool, t *testing.T) client.Client {

	config, _, _ := getServer(host, tls, t)

	c, err := Dial(nil, config)
	if err != nil {
		t.Fatalf("error dialing libStorage service at '%s' %v",
			config.Get("libstorage.host"), err)
	}
	return c
}

type testFunc func(client client.Client, t *testing.T)

func testTCP(tf testFunc, t *testing.T) {
	client := getClient("", false, t)
	tf(client, t)
}

func testTCPTLS(tf testFunc, t *testing.T) {
	client := getClient("", true, t)
	tf(client, t)
}

func testSock(tf testFunc, t *testing.T) {
	sock := fmt.Sprintf("unix://%s", getTempSockFile())
	client := getClient(sock, false, t)
	tf(client, t)
}

func testSockTLS(tf testFunc, t *testing.T) {
	sock := fmt.Sprintf("unix://%s", getTempSockFile())
	client := getClient(sock, true, t)
	tf(client, t)
}

func getTempSockFile() string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	name := f.Name()
	os.RemoveAll(name)
	return fmt.Sprintf("%s.sock", name)
}

func testLogAsJSON(i interface{}, t *testing.T) {
	buf, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	t.Logf("%s\n", string(buf))
}
