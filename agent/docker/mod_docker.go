package docker

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/codedellemc/rexray/agent"
	"github.com/codedellemc/rexray/libstorage/api/context"
	"github.com/codedellemc/rexray/libstorage/api/registry"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	apiutils "github.com/codedellemc/rexray/libstorage/api/utils"
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

const configFormat = `
rexray:
  modules:
    default-docker:
      type:     docker
      desc:     The default Docker module.
      host:     unix://%[1]s
      spec:     %[2]s
      disabled: %[3]v
`

func init() {

	agent.RegisterModule(modName, newModule)
	registry.RegisterConfigReg(
		"Docker",
		func(ctx apitypes.Context, r gofig.ConfigRegistration) {
			var (
				disabled bool
				sockFile = getSockFile(ctx)
			)
			// If CSI_ENDPOINT is set to just "rexray.sock" then that means
			// the Docker module is disabled. This logic is only true for
			// the environment variable, not the config property.
			if v := os.Getenv("CSI_ENDPOINT"); v == sockFile {
				disabled = true
			} else if v := os.Getenv("DOCKER"); v != "" {
				if dd, err := strconv.ParseBool(v); err == nil {
					disabled = !dd
				}
			}
			r.SetYAML(fmt.Sprintf(
				configFormat,
				getSockFile(ctx),
				getSpecFile(ctx),
				disabled))
		})
}

func getSockFile(ctx apitypes.Context) string {
	return path.Join(
		context.MustPathConfig(ctx).Home,
		"/run/docker/plugins/rexray.sock")
}

func getSpecFile(ctx apitypes.Context) string {
	return path.Join(
		context.MustPathConfig(ctx).Home,
		"/etc/docker/plugins/rexray.spec")
}

func newModule(ctx apitypes.Context, c *agent.Config) (agent.Module, error) {

	host := strings.Trim(c.Address, " ")

	if host == "" {
		if c.Name == "default-docker" {
			host = fmt.Sprintf("unix://%s", getSockFile(ctx))
		} else {
			fname := cleanName(c.Name)
			host = fmt.Sprintf("unix://%s",
				path.Join(
					context.MustPathConfig(ctx).Home,
					"/run/docker/plugins",
					fmt.Sprintf("%s.sock", fname)))
		}
	}

	c.Address = host
	config := c.Config

	return &mod{
		ctx:    ctx,
		config: config,
		lsc:    c.Client,
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

type pluginRequest struct {
	Name string            `json:"Name,omitempty"`
	Opts map[string]string `json:"Opts,omitempty"`
}

func (m *mod) Start() error {

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

	if err := os.MkdirAll(
		path.Join(context.MustPathConfig(m.ctx).Home,
			"/etc/docker/plugins"),
		0755); err != nil {
		return err
	}

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
	// m.ctx.WithServiceName(m.ctx.ServiceName())

	mux.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, `{"Implements": ["VolumeDriver"]}`)
		m.ctx.Debug("/Plugin.Activate")
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
		encrypted := store.GetBoolPtr("encrypted")
		_, err := m.lsc.Integration().Create(
			m.ctx,
			pr.Name,
			&apitypes.VolumeCreateOpts{
				AvailabilityZone: store.GetStringPtr("availabilityZone"),
				IOPS:             store.GetInt64Ptr("iops"),
				Size:             store.GetInt64Ptr("size"),
				Type:             vtype,
				Encrypted:        encrypted,
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

		opts := &apitypes.VolumeRemoveOpts{Opts: apiutils.NewStore()}

		// TODO We need the service name
		err := m.lsc.Integration().Remove(m.ctx, pr.Name, opts)
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

		_, err := m.lsc.Integration().Unmount(
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
			m.ctx.WithError(err).Error("/VolumeDriver.Get: error decoding json")
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

	mux.HandleFunc("/VolumeDriver.Capabilities", func(w http.ResponseWriter, r *http.Request) {
		var pr pluginRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, fmt.Sprintf("{\"Error\":\"%s\"}", err.Error()), 500)
			m.ctx.WithError(err).Error("/VolumeDriver.Capabilities: error decoding json")
			return
		}

		m.ctx.WithField("pluginResponse", pr).Debug("/VolumeDriver.Capabilities")

		w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.2+json")
		fmt.Fprintln(w, `{"Capabilities": { "Scope": "global" }}`)
	})

	return mux
}
