package module

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/config"
	"github.com/emccode/rexray/errors"
)

// Module is the interface to which types adhere in order to participate as
// daemon modules.
type Module interface {

	// Id gets the module's unique identifier.
	Id() int32

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
type Init func(id int32, config *ModuleConfig) (Module, error)

var (
	nextModTypeId     int32
	nextModInstanceId int32

	modTypes    map[int32]*ModuleType
	modTypesRwl sync.RWMutex

	modInstances    map[int32]*ModuleInstance
	modInstancesRwl sync.RWMutex
)

func GetModOptVal(opts map[string]string, key string) string {
	if opts == nil {
		return ""
	} else {
		return opts[key]
	}
}

type ModuleConfig struct {
	Address string         `json:"address"`
	Config  *config.Config `json:"config,omitempty"`
}

type ModuleType struct {
	Id               int32           `json:"id"`
	Name             string          `json:"name"`
	IgnoreFailOnInit bool            `json:"-"`
	InitFunc         Init            `json:"-"`
	DefaultConfigs   []*ModuleConfig `json:"defaultConfigs"`
}

type ModuleInstance struct {
	Id          int32         `json:"id"`
	Type        *ModuleType   `json:"-"`
	TypeId      int32         `json:"typeId"`
	Inst        Module        `json:"-"`
	Name        string        `json:"name"`
	Config      *ModuleConfig `json:"config,omitempty"`
	Description string        `json:"description"`
	IsStarted   bool          `json:"started"`
}

func init() {
	nextModTypeId = 0
	nextModInstanceId = 0
	modTypes = make(map[int32]*ModuleType)
	modInstances = make(map[int32]*ModuleInstance)
}

func ModuleTypes() <-chan *ModuleType {

	c := make(chan *ModuleType)

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

func ModuleInstances() <-chan *ModuleInstance {

	c := make(chan *ModuleInstance)

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

func InitializeModule(
	modTypeId int32,
	modConfig *ModuleConfig) (*ModuleInstance, error) {

	modInstancesRwl.Lock()
	defer modInstancesRwl.Unlock()

	lf := log.Fields{
		"typeId":  modTypeId,
		"address": modConfig.Address,
	}

	mt, modTypeExists := modTypes[modTypeId]
	if !modTypeExists {
		return nil, errors.WithFields(lf, "unknown module type")
	}

	lf["typeName"] = mt.Name
	lf["ignoreFailOnInit"] = mt.IgnoreFailOnInit

	modInstId := atomic.AddInt32(&nextModInstanceId, 1)
	mod, initErr := mt.InitFunc(modInstId, modConfig)
	if initErr != nil {
		atomic.AddInt32(&nextModInstanceId, -1)
		return nil, initErr
	}

	modInst := &ModuleInstance{
		Id:          modInstId,
		Type:        mt,
		TypeId:      mt.Id,
		Inst:        mod,
		Name:        mod.Name(),
		Config:      modConfig,
		Description: mod.Description(),
	}
	modInstances[modInstId] = modInst

	lf["id"] = modInstId
	log.WithFields(lf).Info("initialized module instance")

	return modInst, nil
}

func RegisterModule(
	name string,
	ignoreFailOnInit bool,
	initFunc Init,
	defaultConfigs []*ModuleConfig) int32 {
	modTypesRwl.Lock()
	defer modTypesRwl.Unlock()

	modTypeId := atomic.AddInt32(&nextModTypeId, 1)
	modTypes[modTypeId] = &ModuleType{
		Id:               modTypeId,
		Name:             name,
		IgnoreFailOnInit: ignoreFailOnInit,
		InitFunc:         initFunc,
		DefaultConfigs:   defaultConfigs,
	}

	return modTypeId
}

func StartDefaultModules() error {
	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	for id, _ := range modInstances {
		startErr := StartModule(id)
		if startErr != nil {
			return startErr
		}
	}

	return nil
}

func GetModuleInstance(modInstId int32) (*ModuleInstance, error) {
	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	mod, modExists := modInstances[modInstId]

	if !modExists {
		return nil,
			errors.WithField("id", modInstId, "unknown module instance")
	}

	return mod, nil
}

func StartModule(modInstId int32) error {

	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	lf := map[string]interface{}{"id": modInstId}

	mod, modExists := modInstances[modInstId]

	if !modExists {
		return errors.WithFields(lf, "unknown module instance")
	}

	lf["id"] = mod.Id
	lf["typeId"] = mod.Type.Id
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
				mod.TypeId, mod.Id, mod.Name, mod.Config.Address)

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
