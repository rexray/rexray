package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkFormatAuthHeaderVal(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if "Basic YWt1dHo6cGFzc3dvcmQ=" !=
				fmtAuthHeaderVal("akutz", "password") {
				b.FailNow()
			}
		}
	})
}

func TestFormatAuthHeaderVal(t *testing.T) {
	assert.Equal(
		t, "Basic YWt1dHo6cGFzc3dvcmQ=",
		fmtAuthHeaderVal("akutz", "password"))
}

func assertLen(t *testing.T, obj interface{}, expLen int) {
	if !assert.Len(t, obj, expLen) {
		t.FailNow()
	}
}

func assertError(t *testing.T, err error) {
	if !assert.Error(t, err) {
		t.FailNow()
	}
}

func assertNoError(t *testing.T, err error) {
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func assertNil(t *testing.T, i interface{}) {
	if !assert.Nil(t, i) {
		t.FailNow()
	}
}

func assertNotNil(t *testing.T, i interface{}) {
	if !assert.NotNil(t, i) {
		t.FailNow()
	}
}
