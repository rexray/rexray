package utils

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api/types"
)

var (
	instanceIDTypeName = GetTypePkgPathAndName(&types.InstanceID{})
	loggerTypeName     = GetTypePkgPathAndName(&log.Logger{})
)

// GetInstanceID gets a pointer to an InstanceID from the context.
func GetInstanceID(ctx context.Context) (*types.InstanceID, error) {
	obj := ctx.Value("instanceID")
	if obj == nil {
		return nil, NewContextKeyErr("instanceID")
	}
	typedObj, ok := obj.(*types.InstanceID)
	if !ok {
		return nil, NewContextTypeErr(
			"instanceID", instanceIDTypeName, GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// GetLogger gets a pointer to a Logger from the context.
func GetLogger(ctx context.Context) (*log.Logger, error) {
	obj := ctx.Value("logger")
	if obj == nil {
		return nil, NewContextKeyErr("logger")
	}
	typedObj, ok := obj.(*log.Logger)
	if !ok {
		return nil, NewContextTypeErr(
			"logger", loggerTypeName, GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}
