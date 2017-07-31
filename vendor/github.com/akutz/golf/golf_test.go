package golf

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	isDebug = true
	os.Exit(m.Run())
}

func assertMapLen(
	t *testing.T,
	fields map[string]interface{},
	expectedLen int) {

	lf := len(fields)
	if lf != expectedLen {
		t.Fatalf("len(fields) == %d instead of %d", lf, expectedLen)
	}
}

func assertKeyEquals(
	t *testing.T,
	fields map[string]interface{},
	key string,
	expectedVal interface{}) {

	v, ok := fields[key]
	if !ok {
		t.Fatalf("missing %s", key)
	}
	if expectedVal == nil {
		if !isNil(v) {
			t.Fatalf("%s != nil, == %v", key, expectedVal, v)
		}
	} else if v != expectedVal {
		t.Fatalf("%s != %v, == %v", key, expectedVal, v)
	}
}

func assertKeyMissing(
	t *testing.T,
	fields map[string]interface{},
	key string) {

	_, ok := fields[key]
	if ok {
		t.Fatalf("exists %s", key)
	}
}

func assertKeyNotEquals(
	t *testing.T,
	fields map[string]interface{},
	key string,
	expectedVal interface{}) {

	v, ok := fields[key]
	if !ok {
		t.Fatalf("missing %s", key)
	}
	if v == expectedVal {
		t.Fatalf("%s == %v", key, expectedVal)
	}
}
