// Package goioc provides a basic type registry for Go.
package goioc

import (
	"reflect"
	"sync"
)

var (
	ctors    = map[string]func() interface{}{}
	ctorsRWL sync.RWMutex
)

// Register registers a new type with a name and function that returns
// a new instance of the type.
//
// If Register is called with a name that is already registered, the
// old registration will be overridden.
func Register(name string, ctor func() interface{}) {
	ctorsRWL.Lock()
	defer ctorsRWL.Unlock()
	ctors[name] = ctor
}

// New returns a new instance of the specified type name. If the
// specified type name is not found then a nil value is returned.
func New(name string) interface{} {
	ctorsRWL.RLock()
	defer ctorsRWL.RUnlock()
	ctor, ok := ctors[name]
	if !ok {
		return nil
	}
	return ctor()
}

// Implements returns a channel that receives instantiated instances
// of all the registered types that implement the provided interface.
//
// For example, imagine all objects that implement the io.Reader
// interface should be returned:
//
//         Implements((*io.Reader)(nil))
//
// The above call will return instantiated objects for all registered
// types that implement the io.Reader interface.
func Implements(ifaceObj interface{}) chan interface{} {
	c := make(chan interface{})
	go func() {
		ifaceType := reflect.TypeOf(ifaceObj).Elem()
		ctorsRWL.RLock()
		defer ctorsRWL.RUnlock()
		for _, ctor := range ctors {
			o := ctor()
			if reflect.TypeOf(o).Implements(ifaceType) {
				c <- o
			}
		}
		close(c)
	}()
	return c
}
