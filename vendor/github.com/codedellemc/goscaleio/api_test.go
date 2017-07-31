package goscaleio

import (
	"errors"
	"os"
	"testing"

	"github.com/codedellemc/goscaleio/testutil"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
	client *Client
}

var _ = Suite(&S{})

var testServer = testutil.NewHTTPServer()

var authheader = map[string]string{"x-scaleio-authorization": "012345678901234567890123456789"}

func (s *S) SetUpSuite(c *C) {
	testServer.Start()
	var err error

	os.Setenv("GOSCALEIO_ENDPOINT", "http://localhost:4444/api")

	client, err := NewClient()
	if err != nil {
		panic(err)
	}

	testServer.ResponseMap(1,
		testutil.ResponseMap{
			"/api/auth/login": testutil.Response{201, authheader, vaauthorization},
		},
	)

	_, err = client.Authenticate(&ConfigConnect{Username: "username", Password: "password", Endpoint: "http://localhost:4444/api", Version: "2.0"})
	if err != nil {
		panic(err)
	}

	if client.Token == "" {
		errors.New("missing Token")
	}

}

func (s *S) TearDownTest(c *C) {
	testServer.Flush()
}

func TestClient_Authenticate(t *testing.T) {

	testServer.Start()
	var err error
	os.Setenv("GOSCALEIO_ENDPOINT", "http://localhost:4444/api")
	client, err := NewClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	testServer.ResponseMap(1,
		testutil.ResponseMap{
			"/api/auth/login": testutil.Response{201, authheader, vaauthorization},
		},
	)

	_, err = client.Authenticate(&ConfigConnect{Username: "username", Password: "password", Endpoint: "", Version: "2.0"})
	_ = testServer.WaitRequests(1)
	testServer.Flush()
	if err != nil {
		t.Fatalf("Uncatched error: %v", err)
	}

	if client.Token != "012345678901234567890123456789" {
		t.Fatalf("Token not set correctly on client: %s", client.Token)
	}

}

// status: 201
var vaauthorization = `
  <?xml version="1.0" ?>
  <Session href="http://localhost:4444/api/vchs/session" type="application/xml;class=vnd.vmware.vchs.session" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
      <Link href="http://localhost:4444/api/vchs/services" rel="down" type="application/xml;class=vnd.vmware.vchs.servicelist"/>
      <Link href="http://localhost:4444/api/vchs/session" rel="remove"/>
  </Session>
  `
var vaauthorizationErr = `
  <?xml version="1.0" ?>
  <Session href="http://localhost:4444/api/vchs/session" type="application/xml;class=vnd.vmware.vchs.session" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
      <Link href="http://localhost:4444/api/vchs/session" rel="remove"/>
  </Session>
  `
