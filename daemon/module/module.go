package module

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
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

// Init initializes the module with its unique id and the address on which the
// module will be hosted.
type Init func(id int32, address string) (Module, error)

var (
	nextModTypeId     int32
	nextModInstanceId int32

	modTypes    map[int32]*ModuleType
	modTypesRwl sync.RWMutex

	modInstances    map[int32]*ModuleInstance
	modInstancesRwl sync.RWMutex
)

type ModuleType struct {
	Id               int32    `json:"id"`
	Name             string   `json:"name"`
	IgnoreFailOnInit bool     `json:"-"`
	InitFunc         Init     `json:"-"`
	Addresses        []string `json:"addresses"`
}

type ModuleInstance struct {
	Id          int32       `json:"id"`
	Type        *ModuleType `json:"-"`
	TypeId      int32       `json:"typeId"`
	Inst        Module      `json:"-"`
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Description string      `json:"description"`
	IsStarted   bool        `json:"started"`
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
		if mt.Addresses != nil {
			for _, addr := range mt.Addresses {
				_, initErr := InitializeModule(id, addr)
				if initErr != nil {
					if mt.IgnoreFailOnInit {
						log.Printf("%v", initErr)
					} else {
						return initErr
					}
				}
			}
		}
	}

	return nil
}

func InitializeModule(modTypeId int32, address string) (*ModuleInstance, error) {

	modInstancesRwl.Lock()
	defer modInstancesRwl.Unlock()

	mt, modTypeExists := modTypes[modTypeId]
	if !modTypeExists {
		return nil, errors.New(fmt.Sprintf("Unknown module type id %d", modTypeId))
	}

	modInstId := atomic.AddInt32(&nextModInstanceId, 1)
	mod, initErr := mt.InitFunc(modInstId, address)
	if initErr != nil {
		atomic.AddInt32(&nextModInstanceId, -1)
		return nil, errors.New(fmt.Sprintf(
			"Error initializing module type %d, %s at %s ERR: %v\n",
			mt.Id, mt.Name, address, initErr))
	}

	modInst := &ModuleInstance{
		Id:          modInstId,
		Type:        mt,
		TypeId:      mt.Id,
		Inst:        mod,
		Name:        mod.Name(),
		Address:     mod.Address(),
		Description: mod.Description(),
	}
	modInstances[modInstId] = modInst

	log.Printf("Initialized module type %d, %d-%s for %s\n",
		modInst.TypeId, modInst.Id, modInst.Name, modInst.Address)

	return modInst, nil
}

func RegisterModule(
	name string,
	ignoreFailOnInit bool,
	initFunc Init,
	addresses []string) int32 {
	modTypesRwl.Lock()
	defer modTypesRwl.Unlock()

	modTypeId := atomic.AddInt32(&nextModTypeId, 1)
	modTypes[modTypeId] = &ModuleType{
		Id:               modTypeId,
		Name:             name,
		IgnoreFailOnInit: ignoreFailOnInit,
		InitFunc:         initFunc,
		Addresses:        addresses,
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
		return nil, errors.New(fmt.Sprintf(
			"Unknown module instance id %d", modInstId))
	}

	return mod, nil
}

func StartModule(modInstId int32) error {

	modInstancesRwl.RLock()
	defer modInstancesRwl.RUnlock()

	mod, modExists := modInstances[modInstId]

	if !modExists {
		return errors.New(fmt.Sprintf(
			"Unknown module instance id %d", modInstId))
	}

	started := make(chan bool)
	timeout := make(chan bool)
	startError := make(chan error)

	go func() {

		defer func() {
			r := recover()

			errMsg := fmt.Sprintf(
				"Error starting module type %d, %d-%s at %s",
				mod.TypeId, mod.Id, mod.Name, mod.Address)

			if r == nil {
				startError <- errors.New(errMsg)
				return
			}

			switch x := r.(type) {
			case string:
				startError <- errors.New(
					fmt.Sprintf("%s - ERR: %s", errMsg, x))
			case error:
				startError <- errors.New(
					fmt.Sprintf("%s - ERR: %v", errMsg, x))
			default:
				startError <- errors.New(
					fmt.Sprintf("%s - ERR: %v", errMsg, x))
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
		log.Printf("Started module type %d, %d-%s at %s\n",
			mod.TypeId, mod.Id, mod.Name, mod.Address)
	case <-timeout:
		//log.Printf("Timed out starting module type %d, %d-%s at %s\n",
		//	mtid, miid, name, addr)
	case sErr := <-startError:
		return sErr
	}

	return nil
}
