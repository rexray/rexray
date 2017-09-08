package graphdriver

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	graphDriver "github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/go-connections/sockets"
)

const url = "http://localhost"

func TestHandler(t *testing.T) {
	p := &testPlugin{}
	h := NewHandler(p)
	l := sockets.NewInmemSocket("test", 0)
	go h.Serve(l)
	defer l.Close()

	client := &http.Client{Transport: &http.Transport{
		Dial: l.Dial,
	}}

	// Init
	_, err := pluginRequest(client, initPath, &InitRequest{Home: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if p.init != 1 {
		t.Fatalf("expected init 1, got %d", p.init)
	}

	// Create
	_, err = pluginRequest(client, createPath, &CreateRequest{ID: "foo", Parent: "bar"})
	if err != nil {
		t.Fatal(err)
	}
	if p.create != 1 {
		t.Fatalf("expected create 1, got %d", p.create)
	}

	// CreateReadWrite
	_, err = pluginRequest(client, createRWPath, &CreateRequest{ID: "foo", Parent: "bar", MountLabel: "toto"})
	if err != nil {
		t.Fatal(err)
	}
	if p.createRW != 1 {
		t.Fatalf("expected createReadWrite 1, got %d", p.createRW)
	}

	// Remove
	_, err = pluginRequest(client, removePath, RemoveRequest{ID: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if p.remove != 1 {
		t.Fatalf("expected remove 1, got %d", p.remove)
	}

	// Get
	resp, err := pluginRequest(client, getPath, GetRequest{ID: "foo", MountLabel: "bar"})
	if err != nil {
		t.Fatal(err)
	}
	var gResp *GetResponse
	if err := json.NewDecoder(resp).Decode(&gResp); err != nil {
		t.Fatal(err)
	}
	if gResp.Dir != "baz" {
		t.Fatalf("expected dir = 'baz', got %s", gResp.Dir)
	}
	if p.get != 1 {
		t.Fatalf("expected get 1, got %d", p.get)
	}

	// Put
	_, err = pluginRequest(client, putPath, PutRequest{ID: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if p.put != 1 {
		t.Fatalf("expected put 1, got %d", p.put)
	}

	// Exists
	resp, err = pluginRequest(client, existsPath, ExistsRequest{ID: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	var eResp *ExistsResponse
	if err := json.NewDecoder(resp).Decode(&eResp); err != nil {
		t.Fatal(err)
	}
	if !eResp.Exists {
		t.Fatalf("got error testing for existence of graph drivers: %v", eResp.Exists)
	}
	if p.exists != 1 {
		t.Fatalf("expected exists 1, got %d", p.exists)
	}

	// Status
	resp, err = pluginRequest(client, statusPath, StatusRequest{})
	if err != nil {
		t.Fatal(err)
	}
	var sResp *StatusResponse
	if err := json.NewDecoder(resp).Decode(&sResp); err != nil {
		t.Fatal(err)
	}
	if p.status != 1 {
		t.Fatalf("expected get 1, got %d", p.status)
	}

	// GetMetadata
	resp, err = pluginRequest(client, getMetadataPath, GetMetadataRequest{ID: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	var gmResp *GetMetadataResponse
	if err := json.NewDecoder(resp).Decode(&gmResp); err != nil {
		t.Fatal(err)
	}
	if p.getMetadata != 1 {
		t.Fatalf("expected getMetadata 1, got %d", p.getMetadata)
	}

	// Cleanup
	_, err = pluginRequest(client, cleanupPath, CleanupRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if p.cleanup != 1 {
		t.Fatalf("expected cleanup 1, got %d", p.cleanup)
	}

	// Diff
	_, err = pluginRequest(client, diffPath, DiffRequest{ID: "foo", Parent: "bar"})
	if err != nil {
		t.Fatal(err)
	}
	if p.diff != 1 {
		t.Fatalf("expected diff 1, got %d", p.diff)
	}

	// Changes
	resp, err = pluginRequest(client, changesPath, ChangesRequest{ID: "foo", Parent: "bar"})
	if err != nil {
		t.Fatal(err)
	}
	var cResp *ChangesResponse
	if err := json.NewDecoder(resp).Decode(&cResp); err != nil {
		t.Fatal(err)
	}
	if p.status != 1 {
		t.Fatalf("expected get 1, got %d", p.get)
	}

	// ApplyDiff
	b := new(bytes.Buffer)
	stream := bytes.NewReader(b.Bytes())
	resp, err = pluginRequest(client, applyDiffPath, &ApplyDiffRequest{ID: "foo", Parent: "bar", Stream: stream})
	if err != nil {
		t.Fatal(err)
	}
	var adResp *ApplyDiffResponse
	if err := json.NewDecoder(resp).Decode(&adResp); err != nil {
		t.Fatal(err)
	}
	if p.status != 1 {
		t.Fatalf("expected applyDiff 1, got %d", p.applyDiff)
	}

	// DiffSize
	resp, err = pluginRequest(client, diffSizePath, DiffSizeRequest{ID: "foo", Parent: "bar"})
	if err != nil {
		t.Fatal(err)
	}
	var dsResp *DiffSizeResponse
	if err := json.NewDecoder(resp).Decode(&dsResp); err != nil {
		t.Fatal(err)
	}
	if p.diffSize != 1 {
		t.Fatalf("expected diffSize 1, got %d", p.diffSize)
	}

	// Capabilities
	resp, err = pluginRequest(client, capabilitiesPath, CapabilitiesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	var caResp *CapabilitiesResponse
	if err := json.NewDecoder(resp).Decode(&caResp); err != nil {
		t.Fatal(err)
	}
	if caResp.Capabilities.ReproducesExactDiffs != true {
		t.Fatalf("got error getting capabilities for graph drivers: %v", caResp.Capabilities)
	}
	if p.capabilities != 1 {
		t.Fatalf("expected get 1, got %d", p.get)
	}
}

func pluginRequest(client *http.Client, method string, req interface{}) (io.Reader, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if req == nil {
		b = []byte{}
	}
	resp, err := client.Post("http://localhost"+method, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

type testPlugin struct {
	init         int
	create       int
	createRW     int
	remove       int
	get          int
	put          int
	exists       int
	status       int
	getMetadata  int
	cleanup      int
	diff         int
	changes      int
	applyDiff    int
	diffSize     int
	capabilities int
}

var _ Driver = &testPlugin{}

func (p *testPlugin) Init(string, []string, []idtools.IDMap, []idtools.IDMap) error {
	p.init++
	return nil
}

func (p *testPlugin) Create(string, string, string, map[string]string) error {
	p.create++
	return nil
}

func (p *testPlugin) CreateReadWrite(string, string, string, map[string]string) error {
	p.createRW++
	return nil
}

func (p *testPlugin) Remove(string) error {
	p.remove++
	return nil
}

func (p *testPlugin) Get(string, string) (string, error) {
	p.get++
	return "baz", nil
}

func (p *testPlugin) Put(string) error {
	p.put++
	return nil
}

func (p *testPlugin) Exists(string) bool {
	p.exists++
	return true
}

func (p *testPlugin) Status() [][2]string {
	p.status++
	return nil
}

func (p *testPlugin) GetMetadata(string) (map[string]string, error) {
	p.getMetadata++
	return nil, nil
}

func (p *testPlugin) Cleanup() error {
	p.cleanup++
	return nil
}

func (p *testPlugin) Diff(string, string) io.ReadCloser {
	p.diff++
	b := new(bytes.Buffer)
	x := ioutil.NopCloser(bytes.NewReader(b.Bytes()))
	return x
}

func (p *testPlugin) Changes(string, string) ([]Change, error) {
	p.changes++
	return nil, nil
}

func (p *testPlugin) ApplyDiff(string, string, io.Reader) (int64, error) {
	p.applyDiff++
	return 42, nil
}

func (p *testPlugin) DiffSize(string, string) (int64, error) {
	p.diffSize++
	return 42, nil
}

func (p *testPlugin) Capabilities() graphDriver.Capabilities {
	p.capabilities++
	return graphDriver.Capabilities{ReproducesExactDiffs: true}
}
