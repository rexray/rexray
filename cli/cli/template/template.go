package template

import (
	"io"
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"join":  strings.Join,
	"sort":  sortSeq,
	"where": where,
	"json":  jsonify,
	"jsonp": jsonpify,
}

// Template is an interface description for "text/template".
type Template interface {
	Execute(wr io.Writer, data interface{}) error
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

// NewTemplate returns a new template.
func NewTemplate(
	name, format string,
	funcs template.FuncMap) (Template, error) {

	return newTemplate(name, format, funcs)
}

// MustTemplate returns a new template and panics if the provided template
// format is invalid.
func MustTemplate(name, format string, funcs template.FuncMap) Template {
	return template.Must(newTemplate(name, format, funcs))
}

func newTemplate(
	name, format string, funcs template.FuncMap) (*template.Template, error) {

	if funcs == nil {
		funcs = template.FuncMap{}
	}
	for k, v := range funcMap {
		funcs[k] = v
	}
	return template.New(name).Funcs(funcs).Parse(format)
}
