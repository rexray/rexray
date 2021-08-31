package tests

import (
	"fmt"
	"os"
	"path"

	apiclient "github.com/AVENTER-UG/rexray/libstorage/client"
)

func (t *testRunner) copyTLSCertKeyCacerts() {
	err := t.copyFile(
		t.pathConfig.DefaultTLSCertFile, t.serverCrt, 0644)
	Ω(err).ShouldNot(HaveOccurred())
	err = t.copyFile(
		t.pathConfig.DefaultTLSKeyFile, t.serverKey, 0644)
	Ω(err).ShouldNot(HaveOccurred())
	err = t.copyFile(
		t.pathConfig.DefaultTLSTrustedRootsFile, t.cacerts, 0644)
	Ω(err).ShouldNot(HaveOccurred())
}

func (t *testRunner) beforeEachTLSUnix() {
	if t.proto != protoUnix {
		return
	}
	os.Setenv("LIBSTORAGE_TLS_SOCKITTOME", "true")
}

func (t *testRunner) beforeEachTLS() {
	t.beforeEachTLSUnix()
	os.Setenv("LIBSTORAGE_TLS_SERVERNAME", "libstorage-server")
	os.Setenv("LIBSTORAGE_TLS_CERTFILE", t.serverCrt)
	os.Setenv("LIBSTORAGE_TLS_KEYFILE", t.serverKey)
	os.Setenv("LIBSTORAGE_TLS_TRUSTEDCERTSFILE", t.cacerts)
}

func (t *testRunner) beforeEachTLSAuto() {
	t.beforeEachTLSUnix()
	os.Setenv("LIBSTORAGE_TLS_SERVERNAME", "libstorage-server")
	t.copyTLSCertKeyCacerts()
}

func (t *testRunner) beforeEachTLSInsecure() {
	t.beforeEachTLSUnix()
	t.configFileData = []byte(fmt.Sprintf(`
libstorage:
  host: %[1]s://%[2]s
  client:
    tls: insecure
  server:
    tls:
      certFile:         %[4]s
      keyFile:          %[5]s
      trustedCertsFile: %[6]s
    endpoints:
      localhost:
        address: %[1]s://%[2]s
    services:
      %[3]s:
        driver: %[3]s
`, t.proto, t.laddr, t.driverName, t.serverCrt, t.serverKey, t.cacerts))
}

func (t *testRunner) appendUnixSockToKnownHosts() {
	f, err := os.OpenFile(
		t.knownHosts,
		os.O_APPEND|os.O_WRONLY,
		0644)
	Ω(err).ShouldNot(HaveOccurred())
	defer f.Close()
	fmt.Fprintf(f, "%s sha256 %s\n", t.laddr, localhostFingerprint)
}

func (t *testRunner) copyKnownHosts(dstKnownHosts string) {
	t.knownHosts = dstKnownHosts
	Ω(t.knownHosts).ShouldNot(BeAnExistingFile())
	Ω(t.copyFile(t.knownHosts, suiteKnownHosts, 0644)).ShouldNot(HaveOccurred())
	Ω(t.knownHosts).Should(BeARegularFile())
	if t.proto == protoUnix {
		t.appendUnixSockToKnownHosts()
	}
}

func (t *testRunner) beforeEachTLSKnownHosts() {
	t.beforeEachTLSUnix()
	t.copyKnownHosts(path.Join(t.pathConfig.TLS, ".known_hosts"))

	t.configFileData = []byte(fmt.Sprintf(`
libstorage:
  host: %[1]s://%[2]s
  client:
    tls:
      knownHosts: %[4]s
  server:
    tls:
      certFile:         %[5]s
      keyFile:          %[6]s
      trustedCertsFile: %[7]s
    endpoints:
      localhost:
        address: %[1]s://%[2]s
    services:
      %[3]s:
        driver: %[3]s
`, t.proto, t.laddr, t.driverName, t.knownHosts,
		t.serverCrt, t.serverKey, t.cacerts))
}

func (t *testRunner) beforeEachTLSKnownHostsAuto() {
	t.beforeEachTLSUnix()
	t.configFileData = []byte(fmt.Sprintf(`
libstorage:
  host: %[1]s://%[2]s
  server:
    tls:
      certFile:         %[4]s
      keyFile:          %[5]s
      trustedCertsFile: %[6]s
    endpoints:
      localhost:
        address: %[1]s://%[2]s
    services:
      %[3]s:
        driver: %[3]s
`, t.proto, t.laddr, t.driverName, t.serverCrt, t.serverKey, t.cacerts))

	t.copyKnownHosts(t.pathConfig.DefaultTLSKnownHosts)
}

func (t *testRunner) beforeEachTLSAutoKnownHostsAuto() {
	t.beforeEachTLSUnix()
	t.configFileData = []byte(fmt.Sprintf(`
libstorage:
  host: %[1]s://%[2]s
  server:
    endpoints:
      localhost:
        address: %[1]s://%[2]s
    services:
      %[3]s:
        driver: %[3]s
`, t.proto, t.laddr, t.driverName))

	t.copyKnownHosts(t.pathConfig.DefaultTLSKnownHosts)
	t.copyTLSCertKeyCacerts()
}

func (t *testRunner) justBeforeEachTLSRemoveKnownHosts() {
	Ω(t.knownHosts).Should(BeARegularFile())
	Ω(os.RemoveAll(t.knownHosts)).ToNot(HaveOccurred())
	Ω(t.knownHosts).ShouldNot(BeAnExistingFile())
	t.client, t.err = apiclient.New(t.ctx, t.config)
}

func (t *testRunner) justBeforeEachTLSClientNoTLS() {
	t.config.Set("libstorage.client.tls", false)
	t.client, t.err = apiclient.New(t.ctx, t.config)
}

func (t *testRunner) itTLSClientError() {
	Ω(t.err).Should(HaveOccurred())
	Ω(t.client).Should(BeNil())
}
