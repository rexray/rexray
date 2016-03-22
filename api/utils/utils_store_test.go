package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	assert.False(t, s.IsSet("hello"))
	assert.Nil(t, s.Get("hello"))
}

func TestGetStringPtr(t *testing.T) {
	s := NewStore()
	v := "hello"
	fv := fmt.Sprintf("%v", v)
	s.Set("myVal", fv)
	pv := s.GetStringPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &fv)
	pv = s.GetStringPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", v)
	pv = s.GetStringPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &v)
	pv = s.GetStringPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)
}

func TestGetBoolPtr(t *testing.T) {
	s := NewStore()
	v := true
	fv := fmt.Sprintf("%v", v)
	s.Set("myVal", fv)
	pv := s.GetBoolPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &fv)
	pv = s.GetBoolPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", v)
	pv = s.GetBoolPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &v)
	pv = s.GetBoolPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)
}

func TestGetIntPtr(t *testing.T) {
	s := NewStore()
	v := 5
	fv := fmt.Sprintf("%v", v)
	s.Set("myVal", fv)
	pv := s.GetIntPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &fv)
	pv = s.GetIntPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", v)
	pv = s.GetIntPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &v)
	pv = s.GetIntPtr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)
}

func TestGetInt64Ptr(t *testing.T) {
	s := NewStore()
	v := 5
	fv := fmt.Sprintf("%v", v)
	s.Set("myVal", fv)
	pv := s.GetInt64Ptr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &fv)
	pv = s.GetInt64Ptr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", v)
	pv = s.GetInt64Ptr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)

	s.Set("myVal", &v)
	pv = s.GetInt64Ptr("myVal")
	assert.NotNil(t, pv)
	assert.EqualValues(t, v, *pv)
}
