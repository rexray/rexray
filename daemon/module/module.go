package module

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/core/config"
	"github.com/emccode/rexray/core/errors"
)

// Module is the interface to which types adhere in order to participate as
// daemon modules.
type Module interface {

	// Id gets the module's unique identifier.
	ID() int32

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
type Init func(id int32, config *Config) (Module, error)

var (
	nextModTypeID     int32
	nextModInstanceID int32

	modTypes    map[int32]*Type
	modTypesRwl sync.RWMutex

	modInstances    map[int32]*Instance
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
	Address string         `json:"address"`
	Config  *config.Config `json:"config,omitempty"`
}

// Type is a struct that describes a module type
type Type struct {
	ID               int32     `json:"id"`
	Name             string    `json:"name"`
	IgnoreFailOnInit bool      `json:"-"`
	InitFunc         Init      `json:"-"`
	DefaultConfigs   []*Config `json:"defaultConfigs"`
}

// Instance is a struct that describes a module instance
type Instance struct {
	ID          int32   `json:"id"`
	Type        *Type   `json:"-"`
	TypeID      int32   `json:"typeId"`
	Inst        Module  `json:"-"`
	Name        string  `json:"name"`
	Config      *Config `json:"config,omitempty"`
	Description string  `json:"description"`
	IsStarted   bool    `json:"started"`
}

func init() {
	nextModTypeID = 0
	nextModInstanceID = 0
	modTypes = make(map[int32]*Type)
	modInstances = make(map[int32]*Instance)
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
func InitializeDefaultModules() error {
	modTypesRwl.Lock()
	defer modTypesRwl.Unlock()

	for id, mt := range modTypes {
		if mt.DefaultConfigs != nil {
			for _, mc := range mt.DefaultConfigs {
				_, initErr := InitializeModule(id, mc)
				if initErr != nil {
					if mt.IgnoreFailOnInit {
						log.WithField("error", initErr).Warn(
							"ignoring initialization failure")
					} else {
						return initErr
					}
				}
			}
		}
	}

	return nil
}

// InitializeModule initializes a module.
func InitializeModule(
	modTypeID int32,
	modConfig *Config) (*Instance, error) {

	modInstancesRwl.Lock()
	defer modInstancesRwl.Unlock()

	lf := log.Fields{
		"typeId":  modTypeID,
		"address": modConfig.Address,
	}

	mt, modTypeExists := modTypes[modTypeID]
	if !modTypeExists {
		return nil, errors.WithFields(lf, "unknown module type")
	}

	lf["typeName"] = mt.Name
	lf["ignoreFailOnInit"] = mt.IgnoreFailOnInit

	modInstID := atomic.AddInt32(&nextModInstanceID, 1)
	mod, initErr := mt.InitFunc(modInstID, modConfig)
	if initErr != nil {
		atomic.AddInt32(&nextModInstanceID, -1)
		return nil, initErr
	}

	modInst := &Instance{
		ID:          modInstID,
		Type:        mt,
		TypeID:      mt.ID,
		Inst:        mod,
		Name:        mod.Name(),
		Config:      modConfig,
		Description: mod.Description(),
	}
	modInstances[modInstID] = modInst

	lf["id"] = modInstID
	log.WithFields(lf).Info("initialized module instance")

	return modInst, nil
}

// RegisterModule registers a module.
func RegisterModule(
	name string,
	ignoreFailOnInit bool,
	initFunc Init,
	defaultConfigs []*Config) int32 {
	modTypesRwl.Lock()
	defer modTypesRwl.Unlock()

	modTypeID := atomic.AddInt32(&nextModTypeID, 1)
	modTypes[modTypeID] = &Type{
		ID:               modTypeID,
		Name:             name,
		IgnoreFailOnInit: ignoreFailOnInit,
		InitFunc:         initFunc,
		DefaultConfigs:   defaultConfigs,
	}

	return modTypeID
}

// StartDefaultModules starts the default modules.
func StartDefaultModules() error {
	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	for id := range modInstances {
		startErr := StartModule(id)
		if startErr != nil {
			return startErr
		}
	}

	return nil
}

// GetModuleInstance gets the module instance with the provided instance ID.
func GetModuleInstance(modInstID int32) (*Instance, error) {
	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	mod, modExists := modInstances[modInstID]

	if !modExists {
		return nil,
			errors.WithField("id", modInstID, "unknown module instance")
	}

	return mod, nil
}

// StartModule starts the module with the provided instance ID.
func StartModule(modInstID int32) error {

	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	lf := map[string]interface{}{"id": modInstID}

	mod, modExists := modInstances[modInstID]

	if !modExists {
		return errors.WithFields(lf, "unknown module instance")
	}

	lf["id"] = mod.ID
	lf["typeId"] = mod.Type.ID
	lf["typeName"] = mod.Type.Name
	lf["address"] = mod.Config.Address

	started := make(chan bool)
	timeout := make(chan bool)
	startError := make(chan error)

	go func() {

		defer func() {
			r := recover()
			m := "error starting module"

			errMsg := fmt.Sprintf(
				"Error starting module type %d, %d-%s at %s",
				mod.TypeID, mod.ID, mod.Name, mod.Config.Address)

			if r == nil {
				startError <- errors.New(errMsg)
				return
			}

			switch x := r.(type) {
			case string:
				lf["inner"] = x
				startError <- errors.WithFields(lf, m)
			case error:
				startError <- errors.WithFieldsE(lf, m, x)
			default:
				startError <- errors.WithFields(lf, m)
			}
		}()

		sErr := mod.Inst.Start()
		if sErr != nil {
			startError <- sErr
		} else {
			started <- true
		}
	}()

	go func() {
		time.Sleep(3 * time.Second)
		timeout <- true
	}()

	select {
	case <-started:
		mod.IsStarted = true
		log.WithFields(lf).Info("started module")
	case <-timeout:
		log.WithFields(lf).Debug("timed out while monitoring module start")
	case sErr := <-startError:
		return sErr
	}

	return nil
}
