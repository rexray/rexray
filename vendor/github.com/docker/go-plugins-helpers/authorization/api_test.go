package authorization

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-plugins-helpers/sdk"
	"github.com/stretchr/testify/require"
)

type TestPlugin struct {
	Plugin
}

func (p *TestPlugin) AuthZReq(r Request) Response {
	return Response{
		Allow: false,
		Msg:   "You are not authorized",
		Err:   "",
	}
}

func (p *TestPlugin) AuthZRes(r Request) Response {
	return Response{
		Allow: false,
		Msg:   "You are not authorized",
		Err:   "",
	}
}

func TestActivate(t *testing.T) {
	response, err := http.Get("http://localhost:32456/Plugin.Activate")

	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		t.Fatal(err)
	}

	if string(body) != manifest+"\n" {
		t.Fatalf("Expected %s, got %s\n", manifest+"\n", string(body))
	}
}

func TestAuthZReq(t *testing.T) {
	request := `{"User":"bob","UserAuthNMethod":"","RequestMethod":"POST","RequestURI":"http://127.0.0.1/v.1.23/containers/json","RequestBody":"","RequestHeader":"","RequestStatusCode":""}`

	response, err := http.Post(
		"http://localhost:32456/AuthZPlugin.AuthZReq",
		sdk.DefaultContentTypeV1_1,
		strings.NewReader(request),
	)

	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		t.Fatal(err)
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatal(err)
	}

	if r.Msg != "You are not authorized" {
		t.Fatal("Authorization message does not match")
	}

	if r.Allow {
		t.Fatal("The request has been allowed while should not be")
	}

	if r.Err != "" {
		t.Fatal("Authorization Error should be empty")
	}
}

func TestAuthZRes(t *testing.T) {
	request := `{"User":"bob","UserAuthNMethod":"","RequestMethod":"POST","RequestURI":"http://127.0.0.1/v.1.23/containers/json","RequestBody":"","RequestHeader":"","RequestStatusCode":"", "ResponseBody":"","ResponseHeader":"","ResponseStatusCode":200}`

	response, err := http.Post(
		"http://localhost:32456/AuthZPlugin.AuthZRes",
		sdk.DefaultContentTypeV1_1,
		strings.NewReader(request),
	)

	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		t.Fatal(err)
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatal(err)
	}

	if r.Msg != "You are not authorized" {
		t.Fatal("Authorization message does not match")
	}

	if r.Allow {
		t.Fatal("The request has been allowed while should not be")
	}

	if r.Err != "" {
		t.Fatal("Authorization Error should be empty")
	}
}

func TestPeerCertificateMarshalJSON(t *testing.T) {
	template := &x509.Certificate{
		IsCA: true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"Earth"},
			Organization: []string{"Mother Nature"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 5, 5),

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publickey := &privatekey.PublicKey

	// create a self-signed certificate. template = parent
	var parent = template
	raw, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, privatekey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(raw)
	require.NoError(t, err)

	var certs = []*x509.Certificate{cert}
	addr := "www.authz.com/auth"
	req, err := http.NewRequest("GET", addr, nil)
	require.NoError(t, err)

	req.RequestURI = addr
	req.TLS = &tls.ConnectionState{}
	req.TLS.PeerCertificates = certs
	req.Header.Add("header", "value")

	for _, c := range req.TLS.PeerCertificates {
		pcObj := PeerCertificate(*c)

		t.Run("Marshalling :", func(t *testing.T) {
			raw, err = pcObj.MarshalJSON()
			require.NotNil(t, raw)
			require.Nil(t, err)
		})

		t.Run("UnMarshalling :", func(t *testing.T) {
			err := pcObj.UnmarshalJSON(raw)
			require.Nil(t, err)
			require.Equal(t, "Earth", pcObj.Subject.Country[0])
			require.Equal(t, true, pcObj.IsCA)

		})

	}

}

func TestMain(m *testing.M) {
	d := &TestPlugin{}
	h := NewHandler(d)
	go h.ServeTCP("test", ":32456", "", nil)

	os.Exit(m.Run())
}
