package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/emccode/rexray/volume"
)

func init() {

}

var (
	ErrMissingHost      = errors.New("Missing host parameter")
	ErrBadHostSpecified = errors.New("Bad host specified, ie. unix:///usr/share/docker/plugins/rexray.sock or tcp://127.0.0.1:8080")
	ErrBadProtocol      = errors.New("Bad protocol specified with host, ie. unix:// or tcp://")
)

var daemonConfig struct {
	unixListener net.Listener
	httpServer   *http.Server
}

func Start(host string) error {

	if host == "" {
		host = "unix:///usr/share/docker/plugins/rexray.sock"
	}

	protoAndAddr := strings.Split(host, "://")
	if len(protoAndAddr) != 2 {
		return ErrBadHostSpecified
	}

	mux := http.NewServeMux()

	var unixPath string
	if protoAndAddr[0] == "unix" {
		path := protoAndAddr[1]
		_ = os.RemoveAll(path)
		unixPath = fmt.Sprintf("%s://%s", "unix", path)
	} else if protoAndAddr[0] == "tcp" {
	} else {
		return ErrBadProtocol
	}

	type pluginRequest struct {
		Name string `json:"name,ommitempty"`
	}

	mux.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Implements": ["VolumeDriver"]}`)
	})

	mux.HandleFunc("/VolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Remove", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Path", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", ""))
	})

	mux.HandleFunc("/VolumeDriver.Mount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		mountPath, err := volume.MountVolume(pr.Name, "", false, "ext4")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Umount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, err.Error(), 500)
		}

		// p := hostVolumePath(pr.Name)
		// if err := os.RemoveAll(p); err != nil {
		// 	http.Error(w, err.Error(), 500)
		// }

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	if err := os.MkdirAll("/usr/share/docker/plugins", 0755); err != nil {
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
		daemonConfig.httpServer = &http.Server{
			Addr:    host,
			Handler: mux,
		}
		go daemonConfig.httpServer.ListenAndServe()

		specPath = daemonConfig.httpServer.Addr
	}

	if err := ioutil.WriteFile("/usr/share/docker/plugins/rexray.spec", []byte(specPath), 0644); err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Listening for HTTP (%s)", specPath))
	select {}

}
