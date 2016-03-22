package utils

import (
	"fmt"
	"reflect"
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
