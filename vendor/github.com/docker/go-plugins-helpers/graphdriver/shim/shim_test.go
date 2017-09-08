package shim

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/go-connections/sockets"
	graphPlugin "github.com/docker/go-plugins-helpers/graphdriver"
)

type testGraphDriver struct{}

// ProtoDriver
var _ graphdriver.ProtoDriver = &testGraphDriver{}

func (t *testGraphDriver) String() string {
	return ""
}

// FIXME(samoht): this doesn't seem to be called by the plugins
func (t *testGraphDriver) CreateReadWrite(id, parent string, opts *graphdriver.CreateOpts) error {
	return nil
}
func (t *testGraphDriver) Create(id, parent string, opts *graphdriver.CreateOpts) error {
	return nil
}
func (t *testGraphDriver) Remove(id string) error {
	return nil
}
func (t *testGraphDriver) Get(id, mountLabel string) (dir string, err error) {
	return "", nil
}
func (t *testGraphDriver) Put(id string) error {
	return nil
}
func (t *testGraphDriver) Exists(id string) bool {
	return false
}
func (t *testGraphDriver) Status() [][2]string {
	return nil
}
func (t *testGraphDriver) GetMetadata(id string) (map[string]string, error) {
	return nil, nil
}
func (t *testGraphDriver) Cleanup() error {
	return nil
}
func (t *testGraphDriver) Capabilities() graphdriver.Capabilities {
	return graphdriver.Capabilities{}
}

func Init(root string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	d := graphdriver.NewNaiveDiffDriver(&testGraphDriver{}, uidMaps, gidMaps)
	return d, nil
}

func TestGraphDriver(t *testing.T) {
	h := NewHandlerFromGraphDriver(Init)
	l := sockets.NewInmemSocket("test", 0)
	go h.Serve(l)
	defer l.Close()

	client := &http.Client{Transport: &http.Transport{
		Dial: l.Dial,
	}}

	resp, err := pluginRequest(client, "/GraphDriver.Init", &graphPlugin.InitRequest{Home: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatalf("error while creating GraphDriver: %v", err)
	}
}

func pluginRequest(client *http.Client, method string, req *graphPlugin.InitRequest) (*graphPlugin.ErrorResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := client.Post("http://localhost"+method, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var gResp graphPlugin.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&gResp)
	if err != nil {
		return nil, err
	}

	return &gResp, nil
}
