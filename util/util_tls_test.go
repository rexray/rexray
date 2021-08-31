package util

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/akutz/gotil"
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
)

func TestCreateSelfCert(t *testing.T) {
	certPath := "/tmp/libstorage/tls/file.crt"
	keyPath := "/tmp/libstorage/tls/file.key"
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}

	if err := CreateSelfCert(
		context.Background(),
		certPath,
		keyPath,
		host); err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(certPath)
		os.Remove(keyPath)
	}()

	tlsKeypair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("failed to create TLS keypair: %v", err)
	}

	s := httptest.NewUnstartedServer(http.HandlerFunc(
		func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(200)
		},
	))
	defer s.Close()

	s.TLS = &tls.Config{
		Certificates: []tls.Certificate{tlsKeypair},
	}
	s.StartTLS()

	crtPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(crtPEM)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            certPool,
			},
		},
	}
	if _, err := client.Get(s.URL); err != nil {
		t.Fatal(err)
	}

}

func TestAddKnownHost(t *testing.T) {
	knownHostPath := "/tmp/libstorage/tls/known_hosts"
	algo := "sha256"
	host := "localhost"
	fingerprint := []byte("4bd46851a059aab36255863c8d679b6")

	err := AddKnownHost(
		context.Background(),
		knownHostPath,
		host, algo,
		fingerprint)
	defer os.Remove(knownHostPath)

	if err != nil {
		t.Fatal(err)
	}

	if !gotil.FileExists(knownHostPath) {
		t.Fatal("knwon_hosts file not getting created")
	}
}
