package volume

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/docker/go-connections/sockets"
)

func TestHandler(t *testing.T) {
	p := &testPlugin{}
	h := NewHandler(p)
	l := sockets.NewInmemSocket("test", 0)
	go h.Serve(l)
	defer l.Close()

	client := &http.Client{Transport: &http.Transport{
		Dial: l.Dial,
	}}

	// Create
	_, err := pluginRequest(client, createPath, &CreateRequest{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if p.create != 1 {
		t.Fatalf("expected create 1, got %d", p.create)
	}

	// Get
	resp, err := pluginRequest(client, getPath, &GetRequest{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	var gResp *GetResponse
	if err := json.NewDecoder(resp).Decode(&gResp); err != nil {
		t.Fatal(err)
	}
	if gResp.Volume.Name != "foo" {
		t.Fatalf("expected volume `foo`, got %v", gResp.Volume)
	}
	if p.get != 1 {
		t.Fatalf("expected get 1, got %d", p.get)
	}

	// List
	resp, err = pluginRequest(client, listPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	var lResp *ListResponse
	if err := json.NewDecoder(resp).Decode(&lResp); err != nil {
		t.Fatal(err)
	}
	if len(lResp.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %v", lResp.Volumes)
	}
	if lResp.Volumes[0].Name != "foo" {
		t.Fatalf("expected volume `foo`, got %v", lResp.Volumes[0])
	}
	if p.list != 1 {
		t.Fatalf("expected list 1, got %d", p.list)
	}

	// Path
	if _, err := pluginRequest(client, hostVirtualPath, &PathRequest{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	if p.path != 1 {
		t.Fatalf("expected path 1, got %d", p.path)
	}

	// Mount
	if _, err := pluginRequest(client, mountPath, &MountRequest{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	if p.mount != 1 {
		t.Fatalf("expected mount 1, got %d", p.mount)
	}

	// Unmount
	if _, err := pluginRequest(client, unmountPath, &UnmountRequest{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	if p.unmount != 1 {
		t.Fatalf("expected unmount 1, got %d", p.unmount)
	}

	// Remove
	_, err = pluginRequest(client, removePath, &RemoveRequest{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if p.remove != 1 {
		t.Fatalf("expected remove 1, got %d", p.remove)
	}

	// Capabilities
	resp, err = pluginRequest(client, capabilitiesPath, nil)
	var cResp *CapabilitiesResponse
	if err := json.NewDecoder(resp).Decode(&cResp); err != nil {
		t.Fatal(err)
	}

	if p.capabilities != 1 {
		t.Fatalf("expected remove 1, got %d", p.capabilities)
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
	volumes      []string
	create       int
	get          int
	list         int
	path         int
	mount        int
	unmount      int
	remove       int
	capabilities int
}

func (p *testPlugin) Create(req *CreateRequest) error {
	p.create++
	p.volumes = append(p.volumes, req.Name)
	return nil
}

func (p *testPlugin) Get(req *GetRequest) (*GetResponse, error) {
	p.get++
	for _, v := range p.volumes {
		if v == req.Name {
			return &GetResponse{Volume: &Volume{Name: v}}, nil
		}
	}
	return &GetResponse{}, fmt.Errorf("no such volume")
}

func (p *testPlugin) List() (*ListResponse, error) {
	p.list++
	var vols []*Volume
	for _, v := range p.volumes {
		vols = append(vols, &Volume{Name: v})
	}
	return &ListResponse{Volumes: vols}, nil
}

func (p *testPlugin) Remove(req *RemoveRequest) error {
	p.remove++
	for i, v := range p.volumes {
		if v == req.Name {
			p.volumes = append(p.volumes[:i], p.volumes[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no such volume")
}

func (p *testPlugin) Path(req *PathRequest) (*PathResponse, error) {
	p.path++
	for _, v := range p.volumes {
		if v == req.Name {
			return &PathResponse{}, nil
		}
	}
	return &PathResponse{}, fmt.Errorf("no such volume")
}

func (p *testPlugin) Mount(req *MountRequest) (*MountResponse, error) {
	p.mount++
	for _, v := range p.volumes {
		if v == req.Name {
			return &MountResponse{}, nil
		}
	}
	return &MountResponse{}, fmt.Errorf("no such volume")
}

func (p *testPlugin) Unmount(req *UnmountRequest) error {
	p.unmount++
	for _, v := range p.volumes {
		if v == req.Name {
			return nil
		}
	}
	return fmt.Errorf("no such volume")
}

func (p *testPlugin) Capabilities() *CapabilitiesResponse {
	p.capabilities++
	return &CapabilitiesResponse{Capabilities: Capability{Scope: "local"}}
}
