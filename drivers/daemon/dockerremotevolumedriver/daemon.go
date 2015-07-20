package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emccode/rexray/drivers/daemon"
	"github.com/emccode/rexray/volume"
)

const driverName = "dockerremotevolumedriver"

func init() {
	daemondriver.Register(driverName, Init)
}

type Driver struct{}

func Init() (daemondriver.Driver, error) {
	if os.Getenv("REXRAY_DEBUG") == "true" {
		log.Println("Daemon Driver Initialized: " + driverName)
	}
	return &Driver{}, nil
}

var (
	ErrMissingHost      = errors.New("Missing host parameter")
	ErrBadHostSpecified = errors.New("Bad host specified, ie. unix:///run/docker/plugins/rexray.sock or tcp://127.0.0.1:8080")
	ErrBadProtocol      = errors.New("Bad protocol specified with host, ie. unix:// or tcp://")
)

var daemonConfig struct {
	unixListener net.Listener
	httpServer   *http.Server
}

type volumeDriverResponse struct {
	NetworkName string `json:"Networkname,ommitempty"`
	Err         error  `json:",ommitempty"`
}

type pluginRequest struct {
	Name       string `json:"Name,ommitempty"`
	InstanceID string `json:"Instanceid,ommitempty"`
}

func (driver *Driver) Start(host string) error {

	if host == "" {
		host = "unix:///run/docker/plugins/rexray.sock"
	}

	protoAndAddr := strings.Split(host, "://")
	if len(protoAndAddr) != 2 {
		return ErrBadHostSpecified
	}

	mux := http.NewServeMux()

	var unixPath string
	if protoAndAddr[0] == "unix" {
		path := protoAndAddr[1]
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		_ = os.RemoveAll(path)
		unixPath = fmt.Sprintf("%s://%s", "unix", path)
	} else if protoAndAddr[0] == "tcp" {
	} else {
		return ErrBadProtocol
	}

	mux.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Implements": ["RemoteVolumeDriver"]}`)
	})

	mux.HandleFunc("/RemoteVolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		err := volume.Create(pr.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/RemoteVolumeDriver.Remove", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		err := volume.Remove(pr.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/RemoteVolumeDriver.NetworkName", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		if pr.InstanceID == "" {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", errors.New("Missing InstanceID").Error()), 500)
			return
		}

		networkName, err := volume.NetworkName(pr.Name, pr.InstanceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Networkname\": \"%s\"}", networkName))
	})

	mux.HandleFunc("/RemoteVolumeDriver.Attach", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		if pr.InstanceID == "" {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", errors.New("Missing InstanceID").Error()), 500)
			return
		}

		networkName, err := volume.Attach(pr.Name, pr.InstanceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Networkname\": \"%s\"}", networkName))
	})

	mux.HandleFunc("/RemoteVolumeDriver.Detach", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		if pr.InstanceID == "" {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", errors.New("Missing InstanceID").Error()), 500)
			return
		}

		err := volume.Detach(pr.Name, pr.InstanceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	if err := os.MkdirAll("/etc/docker/plugins", 0755); err != nil {
		return err
	}

	var specPath string
	if protoAndAddr[0] == "unix" {
		listener, err := net.Listen("unix", protoAndAddr[1])
		if err != nil {
			return err
		}

		daemonConfig.unixListener = listener
		go http.Serve(daemonConfig.unixListener, mux)
		specPath = unixPath
	} else {
		host = strings.Replace(host, "tcp://", "", 1)
		daemonConfig.httpServer = &http.Server{
			Addr:    host,
			Handler: mux,
		}
		go daemonConfig.httpServer.ListenAndServe()

		specPath = daemonConfig.httpServer.Addr
	}

	if err := ioutil.WriteFile("/etc/docker/plugins/rexray.spec", []byte(specPath), 0644); err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Listening for HTTP (%s)", specPath))
	select {}

}
