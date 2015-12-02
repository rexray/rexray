package volumedriver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/daemon/module"
)

const (
	modAddress     = "unix:///run/docker/plugins/rexray.sock"
	modPort        = 7980
	modName        = "DockerVolumeDriverModule"
	modDescription = "The REX-Ray Docker VolumeDriver module"
)

type mod struct {
	id   int32
	r    *core.RexRay
	name string
	addr string
	desc string
}

func init() {
	//tcpAddr := fmt.Sprintf("tcp://:%d", ModPort)

	_, fsPath, parseAddrErr := gotil.ParseAddress(modAddress)
	if parseAddrErr != nil {
		panic(parseAddrErr)
	}

	fsPathDir := filepath.Dir(fsPath)
	os.MkdirAll(fsPathDir, 0755)

	mc := &module.Config{
		Address: modAddress,
		Config:  gofig.New(),
	}

	module.RegisterModule(modName, true, newMod, []*module.Config{mc})
}

func (m *mod) ID() int32 {
	return m.id
}

func newMod(id int32, cfg *module.Config) (module.Module, error) {
	return &mod{
		id:   id,
		r:    core.New(cfg.Config),
		name: modName,
		desc: modDescription,
		addr: cfg.Address,
	}, nil
}

const driverName = "dockervolumedriver"

var (
	errMissingHost      = goof.New("Missing host parameter")
	errBadHostSpecified = goof.New("Bad host specified, ie. unix:///run/docker/plugins/rexray.sock or tcp://127.0.0.1:8080")
	errBadProtocol      = goof.New("Bad protocol specified with host, ie. unix:// or tcp://")
)

type pluginRequest struct {
	Name string          `json:"Name,omitempty"`
	Opts core.VolumeOpts `json:"Opts,omitempty"`
}

func (m *mod) Start() error {

	proto, addr, parseAddrErr := gotil.ParseAddress(m.Address())
	if parseAddrErr != nil {
		return parseAddrErr
	}

	const validProtoPatt = "(?i)^unix|tcp$"
	isProtoValid, matchProtoErr := regexp.MatchString(validProtoPatt, proto)
	if matchProtoErr != nil {
		return goof.WithFieldsE(goof.Fields{
			"protocol":       proto,
			"validProtoPatt": validProtoPatt,
		}, "error matching protocol", matchProtoErr)
	}
	if !isProtoValid {
		return goof.WithField("protocol", proto, "invalid protocol")
	}

	if err := m.r.InitDrivers(); err != nil {
		return goof.WithFieldsE(goof.Fields{
			"m":   m,
			"m.r": m.r,
		}, "error initializing drivers", err)
	}

	if err := os.MkdirAll("/etc/docker/plugins", 0755); err != nil {
		return err
	}

	var specPath string
	var startFunc func() error

	mux := m.buildMux()

	if proto == "unix" {
		sockFile := addr
		sockFileDir := filepath.Dir(sockFile)
		mkSockFileDirErr := os.MkdirAll(sockFileDir, 0755)
		if mkSockFileDirErr != nil {
			return mkSockFileDirErr
		}

		_ = os.RemoveAll(sockFile)

		specPath = m.Address()
		startFunc = func() error {
			l, lErr := net.Listen("unix", sockFile)
			if lErr != nil {
				return lErr
			}
			defer l.Close()
			defer os.Remove(sockFile)

			return http.Serve(l, mux)
		}
	} else {
		specPath = addr
		startFunc = func() error {
			s := &http.Server{
				Addr:           addr,
				Handler:        mux,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}
			return s.ListenAndServe()
		}
	}

	go func() {
		sErr := startFunc()
		if sErr != nil {
			panic(sErr)
		}
	}()

	writeSpecErr := ioutil.WriteFile(
		"/etc/docker/plugins/rexray.spec", []byte(specPath), 0644)
	if writeSpecErr != nil {
		return writeSpecErr
	}

	return nil
}

func (m *mod) Stop() error {
	return nil
}

func (m *mod) Name() string {
	return m.name
}

func (m *mod) Description() string {
	return m.desc
}

func (m *mod) Address() string {
	return m.addr
}

func (m *mod) buildMux() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Implements": ["VolumeDriver"]}`)
	})

	mux.HandleFunc("/VolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Create: error decoding json")
			return
		}

		err := m.r.Volume.Create(pr.Name, pr.Opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Create: error creating volume")
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Remove", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Remove: error decoding json")
			return
		}

		err := m.r.Volume.Remove(pr.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Remove: error removing volume")
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Path", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Path: error decoding json")
			return
		}

		mountPath, err := m.r.Volume.Path(pr.Name, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Path: error returning path")
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Mount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Mount: error decoding json")
			return
		}

		mountPath, err := m.r.Volume.Mount(pr.Name, "", false, "", false)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Mount: error mounting volume")
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Unmount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Unmount: error decoding json")
			return
		}

		err := m.r.Volume.Unmount(pr.Name, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			log.WithField("error", err).Error("/VolumeDriver.Unmount: error unmounting volume")
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	return mux
}
