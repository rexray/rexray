package errors

import (
	"fmt"
)

type Error struct {
	fields map[string]interface{}
}

type Fields map[string]interface{}

func (e *Error) Error() string {
	return e.fields["message"].(string)
}

func (e *Error) PlayGolf() bool {
	return true
}

func (e *Error) GolfExportedFields() map[string]interface{} {
	return e.fields
}

func New(message string) error {
	return &Error{Fields{"message": message}}
}

func Newf(format string, a ...interface{}) error {
	return &Error{Fields{"message": fmt.Sprintf(format, a)}}
}

func WithError(message string, inner error) error {
	return WithFieldsE(nil, message, inner)
}

func WithField(key string, val interface{}, message string) error {
	return WithFields(Fields{key: val}, message)
}

func WithFieldE(key string, val interface{}, message string, inner error) error {
	return WithFieldsE(Fields{key: val}, message, inner)
}

func WithFields(fields map[string]interface{}, message string) error {
	return WithFieldsE(fields, message, nil)
}

func WithFieldsE(fields map[string]interface{}, message string, inner error) error {

	if fields == nil {
		fields = Fields{}
	}

	if inner != nil {
		fields["inner"] = inner
	}

	fields["message"] = message

	return &Error{fields}
}
