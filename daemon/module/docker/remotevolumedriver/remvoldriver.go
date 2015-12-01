package remotevolumedriver

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

	"github.com/emccode/rexray/daemon/module"

	"github.com/akutz/gofig"
	"github.com/emccode/rexray/core"
	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/util"
)

const (
	modAddress     = "unix:///run/docker/plugins/rexray-remote.sock"
	modPort        = 7981
	modName        = "DockerRemoteVolumeDriverModule"
	modDescription = "The REX-Ray Docker RemoteVolumeDriver module"

	modOptsStorageAdapter = "storageAdapter"
)

type mod struct {
	id   int32
	r    *core.RexRay
	name string
	addr string
	desc string
	stor string
}

func init() {
	//tcpAddr := fmt.Sprintf("tcp://:%d", ModPort)

	_, fsPath, parseAddrErr := util.ParseAddress(modAddress)
	if parseAddrErr != nil {
		panic(parseAddrErr)
	}

	fsPathDir := filepath.Dir(fsPath)
	os.MkdirAll(fsPathDir, 0755)

	mc := &module.Config{
		Address: modAddress,
		Config:  gofig.New(),
	}

	module.RegisterModule(modName, true, newModule, []*module.Config{mc})
}

func optVal(opts map[string]string, key string) string {
	if opts == nil {
		return ""
	}
	return opts[key]
}

const driverName = "dockerremotevolumedriver"

func (m *mod) ID() int32 {
	return m.id
}

func newModule(id int32, cfg *module.Config) (module.Module, error) {
	return &mod{
		id:   id,
		r:    core.New(cfg.Config),
		name: modName,
		desc: modDescription,
		addr: cfg.Address,
	}, nil
}

var (
	errMissingHost      = errors.New("Missing host parameter")
	errBadHostSpecified = errors.New("Bad host specified, ie. unix:///run/docker/plugins/rexray.sock or tcp://127.0.0.1:8080")
	errBadProtocol      = errors.New("Bad protocol specified with host, ie. unix:// or tcp://")
)

type pluginRequest struct {
	Name       string          `json:"Name,omitempty"`
	Opts       core.VolumeOpts `json:"Opts,omitempty"`
	InstanceID string          `json:"Instanceid,omitempty"`
}

func (m *mod) Start() error {

	proto, addr, parseAddrErr := util.ParseAddress(m.Address())
	if parseAddrErr != nil {
		return parseAddrErr
	}

	const validProtoPatt = "(?i)^unix|tcp$"
	isProtoValid, matchProtoErr := regexp.MatchString(validProtoPatt, proto)
	if matchProtoErr != nil {
		return errors.WithFieldsE(errors.Fields{
			"protocol":       proto,
			"validProtoPatt": validProtoPatt,
		}, "error matching protocol", matchProtoErr)
	}
	if !isProtoValid {
		return errors.WithField("protocol", proto, "invalid protocol")
	}

	if err := m.r.InitDrivers(); err != nil {
		return errors.WithFieldsE(errors.Fields{
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
		fmt.Fprintln(w, `{"Implements": ["RemoteVolumeDriver"]}`)
	})

	mux.HandleFunc("/RemoteVolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}
		err := m.r.Volume.Create(pr.Name, pr.Opts)
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

		err := m.r.Volume.Remove(pr.Name)
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

		networkName, err := m.r.Volume.NetworkName(pr.Name, pr.InstanceID)
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

		networkName, err := m.r.Volume.Attach(pr.Name, pr.InstanceID)
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

		err := m.r.Volume.Detach(pr.Name, pr.InstanceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	return mux
}
