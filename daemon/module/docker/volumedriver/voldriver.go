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
	"strings"
	"time"

	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/emccode/libstorage"
	"github.com/emccode/libstorage/api/context"
	apitypes "github.com/emccode/libstorage/api/types"
	apiutils "github.com/emccode/libstorage/api/utils"

	"github.com/emccode/rexray/daemon/module"
)

const (
	modName = "docker"
)

type mod struct {
	lsc    apitypes.Client
	ctx    apitypes.Context
	config gofig.Config
	name   string
	addr   string
	desc   string
}

var (
	separators  = regexp.MustCompile(`[ &_=+:]`)
	dashes      = regexp.MustCompile(`[\-]+`)
	illegalPath = regexp.MustCompile(`[^[:alnum:]\~\-\./]`)
)

func init() {
	module.RegisterModule(modName, newModule)
}

func newModule(c *module.Config) (module.Module, error) {

	host := strings.Trim(c.Address, " ")

	if host == "" {
		if c.Name == "default-docker" {
			host = "unix:///run/docker/plugins/rexray.sock"
		} else {
			fname := cleanName(c.Name)
			host = fmt.Sprintf("unix:///run/docker/plugins/%s.sock", fname)
		}
	}

	c.Address = host

	/*cc, err := c.Config.Copy()
	if err != nil {
		return nil, err
	}*/

	// TODO The config should be copied, but it was causing an issue with
	// duplicate, explicit values not being removed.
	cc := c.Config

	if !cc.GetBool("rexray.volume.path.disableCache") {
		cc.Set("rexray.volume.path.cache", true)
	}

	ctx := context.Background().WithContextID(
		"module", c.Name,
	).WithServiceName(
		cc.GetString("libstorage.service"),
	)

	return &mod{
		ctx:    ctx,
		lsc:    c.Client,
		config: cc,
		name:   c.Name,
		desc:   c.Description,
		addr:   host,
	}, nil
}

func cleanName(s string) string {
	s = strings.Trim(strings.ToLower(s), " ")
	s = separators.ReplaceAllString(s, "-")
	s = illegalPath.ReplaceAllString(s, "")
	s = dashes.ReplaceAllString(s, "-")
	return s
}

const driverName = "docker"

var (
	errMissingHost      = goof.New("Missing host parameter")
	errBadHostSpecified = goof.New("Bad host specified, ie. unix:///run/docker/plugins/rexray.sock or tcp://127.0.0.1:8080")
	errBadProtocol      = goof.New("Bad protocol specified with host, ie. unix:// or tcp://")
)

type pluginRequest struct {
	Name string            `json:"Name,omitempty"`
	Opts map[string]string `json:"Opts,omitempty"`
}

func (m *mod) Start() error {

	lsc, _, err, _ := libstorage.New(m.config)

	if err != nil {
		for k, v := range m.config.AllSettings() {
			m.ctx.Errorf("%s=%v", k, v)
		}
		m.ctx.Fatal(err)
	}
	m.lsc = lsc

	proto, addr, parseAddrErr := gotil.ParseAddress(m.Address())
	if parseAddrErr != nil {
		return parseAddrErr
	}

	if proto == "unix" {
		dir := filepath.Dir(addr)
		os.MkdirAll(dir, 0755)
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

	spec := m.config.GetString("spec")
	if spec == "" {
		if m.name == "default-docker" {
			spec = "/etc/docker/plugins/rexray.spec"
		} else {
			fname := cleanName(m.name)
			spec = fmt.Sprintf("/etc/docker/plugins/%s.spec", fname)
		}
	}

	m.ctx.WithField("path", spec).Debug("docker voldriver spec file")

	if !gotil.FileExists(spec) {
		if err := ioutil.WriteFile(spec, []byte(specPath), 0644); err != nil {
			return err
		}
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
	m.ctx.WithServiceName(m.ctx.ServiceName())

	mux.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, `{"Implements": ["VolumeDriver"]}`)
		m.ctx.Debug("/VolumeDriver.Create")
	})

	mux.HandleFunc("/VolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Create: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Create")
		store := apiutils.NewStoreWithVars(pr.Opts)
		vtype := store.GetStringPtr("type")
		if vtype == nil {
			vtype = store.GetStringPtr("volumetype")
		}
		_, err := m.lsc.Integration().Create(
			m.ctx,
			pr.Name,
			&apitypes.VolumeCreateOpts{
				AvailabilityZone: store.GetStringPtr("availabilityZone"),
				IOPS:             store.GetInt64Ptr("iops"),
				Size:             store.GetInt64Ptr("size"),
				Type:             vtype,
				Opts:             store,
			})

		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Create: error creating volume")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Remove", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Remove: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Remove")

		// TODO We need the service name
		err := m.lsc.Integration().Remove(m.ctx, pr.Name, apiutils.NewStore())
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Remove: error removing volume")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Path", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Path: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Path")

		mountPath, err := m.lsc.Integration().Path(
			m.ctx, "", pr.Name, apiutils.NewStore())
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Path: error returning path")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Mount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Mount: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Mount")

		mountPath, _, err := m.lsc.Integration().Mount(
			m.ctx, "", pr.Name, &apitypes.VolumeMountOpts{})
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Mount: error mounting volume")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, fmt.Sprintf("{\"Mountpoint\": \"%s\"}", mountPath))
	})

	mux.HandleFunc("/VolumeDriver.Unmount", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Unmount: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Unmount")

		err := m.lsc.Integration().Unmount(
			m.ctx, "", pr.Name, apiutils.NewStore())
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Unmount: error unmounting volume")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, `{}`)
	})

	mux.HandleFunc("/VolumeDriver.Get", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Path: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Get")

		volMapping, err := m.lsc.Integration().Inspect(
			m.ctx, pr.Name, apiutils.NewStore())
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Get: error getting volume")
			return
		}

		w.Header().Set(
			"Content-Type", "application/vnd.docker.plugins.v1.2+json")
		json.NewEncoder(w).Encode(map[string]apitypes.VolumeMapping{
			"Volume": volMapping,
		})
	})

	mux.HandleFunc("/VolumeDriver.List", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.List: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.List")

		volMappings, err := m.lsc.Integration().List(m.ctx, apiutils.NewStore())
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.List: error listing volumes")
			return
		}

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		json.NewEncoder(w).Encode(
			map[string][]apitypes.VolumeMapping{"Volumes": volMappings})
	})

	return mux
}
