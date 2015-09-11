package volumedriver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/emccode/rexray/daemon/module"

	"github.com/emccode/rexray/config"
	osm "github.com/emccode/rexray/os"
	"github.com/emccode/rexray/storage"
	"github.com/emccode/rexray/util"
	"github.com/emccode/rexray/volume"
)

const MOD_ADDR = "unix:///run/docker/plugins/rexray.sock"
const MOD_PORT = 7980
const MOD_NAME = "DockerVolumeDriverModule"
const MOD_DESC = "The REX-Ray Docker VolumeDriver module"

type Module struct {
	id   int32
	vdm  *volume.VolumeDriverManager
	name string
	addr string
	desc string
}

func init() {
	//tcpAddr := fmt.Sprintf("tcp://:%d", MOD_PORT)

	_, fsPath, parseAddrErr := util.ParseAddress(MOD_ADDR)
	if parseAddrErr != nil {
		panic(parseAddrErr)
	}

	fsPathDir := filepath.Dir(fsPath)
	os.MkdirAll(fsPathDir, 0755)

	mc := &module.ModuleConfig{
		Address: MOD_ADDR,
		Config:  config.New(),
	}

	module.RegisterModule(MOD_NAME, true, Init, []*module.ModuleConfig{mc})
}

func (mod *Module) Id() int32 {
	return mod.id
}

func Init(id int32, cfg *module.ModuleConfig) (module.Module, error) {

	osdm, osdmErr := osm.NewOSDriverManager(cfg.Config)
	if osdmErr != nil {
		return nil, osdmErr
	}
	if len(osdm.Drivers) == 0 {
		return nil, errors.New("no os drivers initialized")
	}

	sdm, sdmErr := storage.NewStorageDriverManager(cfg.Config)
	if sdmErr != nil {
		return nil, sdmErr
	}
	if len(sdm.Drivers) == 0 {
		return nil, errors.New("no storage drivers initialized")
	}

	vdm, vdmErr := volume.NewVolumeDriverManager(cfg.Config, osdm, sdm)
	if vdmErr != nil {
		return nil, vdmErr
	}
	if len(vdm.Drivers) == 0 {
		return nil, errors.New("no volume drivers initialized")
	}

	return &Module{
		id:   id,
		vdm:  vdm,
		name: MOD_NAME,
		desc: MOD_DESC,
		addr: cfg.Address,
	}, nil
}

const driverName = "dockervolumedriver"

var (
	ErrMissingHost      = errors.New("Missing host parameter")
	ErrBadHostSpecified = errors.New("Bad host specified, ie. unix:///run/docker/plugins/rexray.sock or tcp://127.0.0.1:8080")
	ErrBadProtocol      = errors.New("Bad protocol specified with host, ie. unix:// or tcp://")
)

type pluginRequest struct {
	Name string `json:"Name,ommitempty"`
}

func (mod *Module) Start() error {

	proto, addr, parseAddrErr := util.ParseAddress(mod.Address())
	if parseAddrErr != nil {
		return parseAddrErr
	}

	const validProtoPatt = "(?i)^unix|tcp$"
	isProtoValid, matchProtoErr := regexp.MatchString(validProtoPatt, proto)
	if matchProtoErr != nil {
		return errors.New(fmt.Sprintf(
			"Error matching protocol %s with pattern '%s' ERR: %v",
			proto, validProtoPatt, matchProtoErr))
	}
	if !isProtoValid {
		return errors.New(fmt.Sprintf("Invalid protocol %s", proto))
	}

	if err := os.MkdirAll("/etc/docker/plugins", 0755); err != nil {
		return err
	}

	var specPath string
	var startFunc func() error

	mux := mod.buildMux()

	if proto == "unix" {
		sockFile := addr
		sockFileDir := filepath.Dir(sockFile)
		mkSockFileDirErr := os.MkdirAll(sockFileDir, 0755)
		if mkSockFileDirErr != nil {
			return mkSockFileDirErr
		}

		_ = os.RemoveAll(sockFile)

		specPath = mod.Address()
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

func (mod *Module) Stop() error {
	return nil
}

func (mod *Module) Name() string {
	return mod.name
}

func (mod *Module) Description() string {
	return mod.desc
}

func (mod *Module) Address() string {
	return mod.addr
}

func (mod *Module) buildMux() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{"Implements": ["VolumeDriver"]}`)
	})

	mux.HandleFunc("/VolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		err := mod.vdm.Create(pr.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Remove", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		err := mod.vdm.Remove(pr.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Path", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		mountPath, err := mod.vdm.Path(pr.Name, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Mount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		mountPath, err := mod.vdm.Mount(pr.Name, "", false, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Unmount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		err := mod.vdm.Unmount(pr.Name, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	return mux
}
