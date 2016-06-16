package module

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/emccode/libstorage/api/context"
	apiserver "github.com/emccode/libstorage/api/server"
	apitypes "github.com/emccode/libstorage/api/types"
	apiclient "github.com/emccode/libstorage/client"

	"github.com/emccode/rexray/util"
)

// Module is the interface to which types adhere in order to participate as
// daemon modules.
type Module interface {

	// Start starts the module.
	Start() error

	// Stop signals the module to stop.
	Stop() error

	// Name is the name of the module.
	Name() string

	// Address is the network address at which the module is available.
	Address() string

	// Description is a free-form field ot add descriptive information about
	// the module instance.
	Description() string
}

// Init initializes the module.
type Init func(ctx apitypes.Context, config *Config) (Module, error)

var (
	modTypes    map[string]*Type
	modTypesRwl sync.RWMutex

	modInstances    map[string]*Instance
	modInstancesRwl sync.RWMutex
)

// GetModOptVal gets a module's option value.
func GetModOptVal(opts map[string]string, key string) string {
	if opts == nil {
		return ""
	}
	return opts[key]
}

// Config is a struct used to configure a module.
type Config struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Address     string          `json:"address"`
	Config      gofig.Config    `json:"config,omitempty"`
	Client      apitypes.Client `json:"-"`
}

// Type is a struct that describes a module type
type Type struct {
	Name     string `json:"name"`
	InitFunc Init   `json:"-"`
}

// Instance is a struct that describes a module instance
type Instance struct {
	Type        *Type   `json:"-"`
	TypeName    string  `json:"typeName"`
	Inst        Module  `json:"-"`
	Name        string  `json:"name"`
	Config      *Config `json:"config,omitempty"`
	Description string  `json:"description"`
	IsStarted   bool    `json:"started"`
}

func init() {
	modTypes = map[string]*Type{}
	modInstances = map[string]*Instance{}
	initConfig()
}

func initConfig() {
	cfg := gofig.NewRegistration("Module")
	cfg.Yaml(`
rexray:
    modules:
        default-admin:
            type:     admin
            desc:     The default admin module.
            host:     unix:///var/run/rexray/server.sock
            disabled: false
        default-docker:
            type:     docker
            desc:     The default docker module.
            host:     unix:///run/docker/plugins/rexray.sock
            spec:     /etc/docker/plugins/rexray.spec
            disabled: false
`)
	cfg.Key(gofig.String, "", "10s", "", "rexray.module.startTimeout")
	gofig.Register(cfg)
}

// Types returns a channel that receives the registered module types.
func Types() <-chan *Type {

	c := make(chan *Type)

	go func() {
		modTypesRwl.RLock()
		defer modTypesRwl.RUnlock()

		for _, mt := range modTypes {
			c <- mt
		}

		close(c)
	}()

	return c
}

// Instances returns a channel that receives the instantiated module instances.
func Instances() <-chan *Instance {

	c := make(chan *Instance)

	go func() {
		modInstancesRwl.RLock()
		defer modInstancesRwl.RUnlock()

		for _, mi := range modInstances {
			c <- mi
		}

		close(c)
	}()

	return c
}

// InitializeDefaultModules initializes the default modules.
func InitializeDefaultModules(
	ctx apitypes.Context,
	config gofig.Config) (<-chan error, error) {

	modTypesRwl.RLock()
	defer modTypesRwl.RUnlock()

	config = config.Scope("rexray")

	if !config.IsSet(apitypes.ConfigIgVolOpsMountPath) {
		config.Set(
			apitypes.ConfigIgVolOpsMountPath, util.LibFilePath("volumes"))
	}

	var serverErrChan <-chan error

	if config.GetBool(apitypes.ConfigEmbedded) {

		var (
			err    error
			server apitypes.Server
			errs   <-chan error
		)

		host, isRunning := util.IsLocalServerActive(ctx, config)

		if isRunning {
			ctx = ctx.WithValue(context.HostKey, host)
			ctx.WithField("host", host).Debug(
				"not starting embeddded server; " +
					"local server already running")
		} else {
			if server, errs, err = apiserver.Serve(ctx, config); err != nil {
				return nil, err
			}
			go func() {
				if err = <-errs; err != nil {
					ctx.Error(err)
				}
			}()
			if host == "" {
				config.Set(apitypes.ConfigHost, server.Addrs()[0])
			}

			serverErrChan = errs
		}

	}

	modConfigs, err := getConfiguredModules(ctx, config)
	if err != nil {
		return nil, err
	}

	for _, mc := range modConfigs {

		lsc, err := apiclient.New(ctx, mc.Config)
		if err != nil {
			panic(err)
		}
		mc.Client = lsc
		mod, err := InitializeModule(ctx, mc)
		if err != nil {
			return nil, err
		}
		func() {
			modInstancesRwl.Lock()
			defer modInstancesRwl.Unlock()
			modInstances[mod.Name] = mod
		}()
	}

	return serverErrChan, nil
}

// InitializeModule initializes a module.
func InitializeModule(
	ctx apitypes.Context, modConfig *Config) (*Instance, error) {

	modInstancesRwl.Lock()
	defer modInstancesRwl.Unlock()

	typeName := strings.ToLower(modConfig.Type)

	lf := log.Fields{
		"typeName": typeName,
		"address":  modConfig.Address,
	}

	mt, modTypeExists := modTypes[typeName]
	if !modTypeExists {
		return nil, goof.WithFields(lf, "unknown module type")
	}

	mod, initErr := mt.InitFunc(ctx, modConfig)
	if initErr != nil {
		return nil, initErr
	}

	modName := mod.Name()

	modInst := &Instance{
		Type:        mt,
		TypeName:    typeName,
		Inst:        mod,
		Name:        modName,
		Config:      modConfig,
		Description: mod.Description(),
	}
	modInstances[modName] = modInst

	lf["name"] = modName
	ctx.WithFields(lf).Info("initialized module instance")

	return modInst, nil
}

// RegisterModule registers a module type.
func RegisterModule(name string, initFunc Init) {

	modTypesRwl.Lock()
	defer modTypesRwl.Unlock()

	name = strings.ToLower(name)
	modTypes[name] = &Type{
		Name:     name,
		InitFunc: initFunc,
	}
}

// StartDefaultModules starts the default modules.
func StartDefaultModules(ctx apitypes.Context, config gofig.Config) error {
	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	for name := range modInstances {
		startErr := StartModule(ctx, config, name)
		if startErr != nil {
			return startErr
		}
	}

	return nil
}

// GetModuleInstance gets the module instance with the provided name.
func GetModuleInstance(name string) (*Instance, error) {
	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	name = strings.ToLower(name)
	mod, modExists := modInstances[name]

	if !modExists {
		return nil,
			goof.WithField("name", name, "unknown module instance")
	}

	return mod, nil
}

// StartModule starts the module with the provided instance name.
func StartModule(ctx apitypes.Context, config gofig.Config, name string) error {

	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	name = strings.ToLower(name)
	lf := map[string]interface{}{"name": name}

	mod, modExists := modInstances[name]

	if !modExists {
		return goof.WithFields(lf, "unknown module instance")
	}

	lf["typeName"] = mod.Type.Name
	lf["address"] = mod.Config.Address

	started := make(chan bool)
	timeout := make(chan bool)
	startError := make(chan error)

	go func() {

		sErr := mod.Inst.Start()
		if sErr != nil {
			startError <- sErr
		} else {
			started <- true
		}
	}()

	go func() {
		time.Sleep(startTimeout(config))
		timeout <- true
	}()

	select {
	case <-started:
		mod.IsStarted = true
		ctx.WithFields(lf).Info("started module")
	case <-timeout:
		return goof.New("timed out while monitoring module start")
	case sErr := <-startError:
		return sErr
	}

	return nil
}

func getConfiguredModules(
	ctx apitypes.Context, c gofig.Config) ([]*Config, error) {

	mods := c.Get("rexray.modules")
	modMap, ok := mods.(map[string]interface{})
	if !ok {
		return nil, goof.New("invalid format rexray.modules")
	}
	ctx.WithField("count", len(modMap)).Debug("got modules map")

	modConfigs := []*Config{}

	for name := range modMap {
		name = strings.ToLower(name)

		ctx.WithField("name", name).Debug("processing module config")

		scope := fmt.Sprintf("rexray.modules.%s", name)
		ctx.WithField("scope", scope).Debug("getting scoped config for module")
		sc := c.Scope(scope)

		disabled := sc.GetBool("disabled")
		if disabled {
			ctx.WithField("name", name).Debug("ignoring disabled module config")
			continue
		}

		mc := &Config{
			Name:        name,
			Type:        strings.ToLower(sc.GetString("type")),
			Description: sc.GetString("desc"),
			Address:     sc.GetString("host"),
			Config:      sc,
		}

		ctx.WithFields(log.Fields{
			"name": mc.Name,
			"type": mc.Type,
			"desc": mc.Description,
			"addr": mc.Address,
		}).Info("created new mod config")

		modConfigs = append(modConfigs, mc)
	}

	return modConfigs, nil
}

func startTimeout(config gofig.Config) time.Duration {
	dur, err := time.ParseDuration(
		config.GetString("rexray.module.startTimeout"))
	if err != nil {
		return time.Duration(10) * time.Second
	}
	return dur
}
