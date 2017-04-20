package util

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/akutz/gotil"
	"github.com/codedellemc/libstorage/api/context"
)

func TestCreateSelfCert(t *testing.T) {
	certPath := "/tmp/libstorage/tls/file.crt"
	keyPath := "/tmp/libstorage/tls/file.key"
	host := "127.0.0.1"

	err := CreateSelfCert(context.Background(), certPath, keyPath, host)
	if err != nil {
		t.Error(err)
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

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	if _, err := client.Get(s.URL); err != nil {
		log.Println(err)
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
