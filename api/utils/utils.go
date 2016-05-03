package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	_ "github.com/akutz/golf"
)

// GetTypePkgPathAndName gets ths type and package path of the provided
// instance.
func GetTypePkgPathAndName(i interface{}) string {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}
	pkgPath := t.PkgPath()
	typeName := t.Name()
	if pkgPath == "" {
		return typeName
	}
	return fmt.Sprintf("%s.%s", pkgPath, typeName)
}

// GetTempSockFile returns a new sock file in a temp space.
func GetTempSockFile() string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	name := f.Name()
	os.RemoveAll(name)
	return fmt.Sprintf("%s.sock", name)
}
