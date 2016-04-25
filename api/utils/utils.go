package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

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

// GetHeader is a case-insensitive way to retrieve a header's value.
func GetHeader(headers http.Header, name string) []string {
	for k, v := range headers {
		if strings.ToLower(k) == strings.ToLower(name) {
			return v
		}
	}
	return nil
}
