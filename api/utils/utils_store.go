package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/emccode/libstorage/api/types"
)

type keyValueStore struct {
	store map[string]interface{}
}

// NewStore initializes a new instance of the Store type.
func NewStore() types.Store {
	return newKeyValueStore(map[string]interface{}{})
}

// NewStoreWithData initializes a new instance of the Store type.
func NewStoreWithData(data map[string]interface{}) types.Store {
	return newKeyValueStore(data)
}

// NewStoreWithVars initializes a new instance of the Store type.
func NewStoreWithVars(vars map[string]string) types.Store {
	m := map[string]interface{}{}
	for k, v := range vars {
		m[k] = v
	}
	return newKeyValueStore(m)
}

func newKeyValueStore(m map[string]interface{}) types.Store {
	cm := map[string]interface{}{}
	for k, v := range m {
		cm[strings.ToLower(k)] = v
	}
	return &keyValueStore{cm}
}

func (s *keyValueStore) IsSet(k string) bool {
	_, ok := s.store[strings.ToLower(k)]
	return ok
}

func (s *keyValueStore) Get(k string) interface{} {
	return s.store[strings.ToLower(k)]
}

func (s *keyValueStore) GetStore(k string) types.Store {
	v := s.Get(k)
	switch tv := v.(type) {
	case types.Store:
		return tv
	default:
		return nil
	}
}

func (s *keyValueStore) GetString(k string) string {
	v := s.Get(k)
	switch tv := v.(type) {
	case string:
		return tv
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", tv)
	}
}

func (s *keyValueStore) GetStringPtr(k string) *string {
	v := s.Get(k)
	switch tv := v.(type) {
	case *string:
		return tv
	case string:
		return &tv
	case nil:
		return nil
	default:
		str := getStrFromPossiblePtr(v)
		return &str
	}
}

func (s *keyValueStore) GetBool(k string) bool {
	v := s.Get(k)
	switch tv := v.(type) {
	case bool:
		return tv
	case nil:
		return false
	default:
		b, _ := strconv.ParseBool(s.GetString(k))
		return b
	}
}

func (s *keyValueStore) GetBoolPtr(k string) *bool {
	v := s.Get(k)
	switch tv := v.(type) {
	case *bool:
		return tv
	case bool:
		return &tv
	case nil:
		return nil
	default:
		str := getStrFromPossiblePtr(v)
		b, _ := strconv.ParseBool(str)
		return &b
	}
}

func (s *keyValueStore) GetInt(k string) int {
	v := s.Get(k)
	switch tv := v.(type) {
	case int:
		return tv
	case nil:
		return 0
	default:
		if iv, err := strconv.ParseInt(s.GetString(k), 10, 64); err == nil {
			return int(iv)
		}
		return 0
	}
}

func (s *keyValueStore) GetIntPtr(k string) *int {
	v := s.Get(k)
	switch tv := v.(type) {
	case *int:
		return tv
	case int:
		return &tv
	case nil:
		return nil
	default:
		str := getStrFromPossiblePtr(v)
		var iivp *int
		if iv, err := strconv.ParseInt(str, 10, 64); err == nil {
			iiv := int(iv)
			iivp = &iiv
		}
		return iivp
	}
}

func (s *keyValueStore) GetInt64(k string) int64 {
	v := s.Get(k)
	switch tv := v.(type) {
	case int64:
		return tv
	case nil:
		return 0
	default:
		if iv, err := strconv.ParseInt(s.GetString(k), 10, 64); err == nil {
			return iv
		}
		return 0
	}
}

func (s *keyValueStore) GetInt64Ptr(k string) *int64 {
	v := s.Get(k)
	switch tv := v.(type) {
	case *int64:
		return tv
	case int64:
		return &tv
	case nil:
		return nil
	default:
		str := getStrFromPossiblePtr(v)
		var ivp *int64
		if iv, err := strconv.ParseInt(str, 10, 64); err == nil {
			ivp = &iv
		}
		return ivp
	}
}

func (s *keyValueStore) GetStringSlice(k string) []string {
	v := s.Get(k)
	switch tv := v.(type) {
	case []string:
		return tv
	default:
		return nil
	}
}

func (s *keyValueStore) GetIntSlice(k string) []int {
	v := s.Get(k)
	switch tv := v.(type) {
	case []int:
		return tv
	default:
		return nil
	}
}

func (s *keyValueStore) GetBoolSlice(k string) []bool {
	v := s.Get(k)
	switch tv := v.(type) {
	case []bool:
		return tv
	default:
		return nil
	}
}

func (s *keyValueStore) GetMap(k string) map[string]interface{} {
	v := s.Get(k)
	switch tv := v.(type) {
	case map[string]interface{}:
		return tv
	default:
		return nil
	}
}

func (s *keyValueStore) Set(k string, v interface{}) {
	s.store[strings.ToLower(k)] = v
}

func (s *keyValueStore) Keys() []string {
	keys := []string{}
	for k := range s.store {
		keys = append(keys, k)
	}
	return keys
}

func getStrFromPossiblePtr(i interface{}) string {
	rv := reflect.ValueOf(i)
	if rv.Kind() == reflect.Ptr {
		return fmt.Sprintf("%v", rv.Elem())
	}
	return fmt.Sprintf("%v", i)
}
