package shim

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/docker/docker/volume"
	"github.com/docker/go-connections/sockets"
	volumeplugin "github.com/docker/go-plugins-helpers/volume"
)

type testVolumeDriver struct{}
type testVolume struct{}

func (testVolume) Name() string           { return "" }
func (testVolume) Path() string           { return "" }
func (testVolume) Mount() (string, error) { return "", nil }
func (testVolume) Unmount() error         { return nil }
func (testVolume) DriverName() string     { return "" }

func (testVolumeDriver) Name() string                                            { return "" }
func (testVolumeDriver) Create(string, map[string]string) (volume.Volume, error) { return nil, nil }
func (testVolumeDriver) Remove(volume.Volume) error                              { return nil }
func (testVolumeDriver) List() ([]volume.Volume, error)                          { return nil, nil }
func (testVolumeDriver) Get(name string) (volume.Volume, error)                  { return nil, nil }
func (testVolumeDriver) Scope() string                                           { return "local" }

func TestVolumeDriver(t *testing.T) {
	h := NewHandlerFromVolumeDriver(testVolumeDriver{})
	l := sockets.NewInmemSocket("test", 0)
	go h.Serve(l)
	defer l.Close()

	client := &http.Client{Transport: &http.Transport{
		Dial: l.Dial,
	}}

	resp, err := pluginRequest(client, "/VolumeDriver.Create", &volumeplugin.CreateRequest{Name: "foo"})
	if err != nil {
		t.Fatalf(err.Error())
	}

	if resp.Err != "" {
		t.Fatalf("error while creating volume: %v", err)
	}
}

func pluginRequest(client *http.Client, method string, req *volumeplugin.CreateRequest) (*volumeplugin.ErrorResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := client.Post("http://localhost"+method, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var vResp volumeplugin.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&vResp)
	if err != nil {
		return nil, err
	}

	return &vResp, nil
}
