package gocsi

import (
	"sync"
	"time"

	"github.com/codedellemc/gocsi/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"golang.org/x/net/context"
)

// IdempotencyProvider is the interface that works with a server-side,
// gRPC interceptor to provide serial access and idempotency for CSI's
// volume resources.
type IdempotencyProvider interface {
	// GetVolumeName should return the name of the volume specified
	// by the provided volume ID. If the volume does not exist then
	// an empty string should be returned.
	GetVolumeName(
		ctx context.Context,
		id *csi.VolumeID) (string, error)

	// GetVolumeInfo should return information about the volume
	// specified by the provided volume name. If the volume does not
	// exist then a nil value should be returned.
	GetVolumeInfo(
		ctx context.Context,
		name string) (*csi.VolumeInfo, error)

	// IsControllerPublished should return publication info about
	// the volume specified by the provided volume name or ID.
	IsControllerPublished(
		ctx context.Context,
		id *csi.VolumeID) (*csi.PublishVolumeInfo, error)

	// IsNodePublished should return a flag indicating whether or
	// not the volume exists and is published on the current host.
	IsNodePublished(
		ctx context.Context,
		id *csi.VolumeID,
		pubVolInfo *csi.PublishVolumeInfo,
		targetPath string) (bool, error)
}

// NewIdempotentInterceptor returns a new server-side, gRPC interceptor
// that can be used in conjunction with an IdempotencyProvider to
// provide serialized, idempotent access to the following CSI RPCs:
//
//  * CreateVolume
//  * DeleteVolume
//  * ControllerPublishVolume
//  * ControllerUnpublishVolume
//  * NodePublishVolume
//  * NodeUnpublishVolume
func NewIdempotentInterceptor(
	p IdempotencyProvider,
	timeout time.Duration) grpc.UnaryServerInterceptor {

	i := &idempotencyInterceptor{
		p:        p,
		timeout:  timeout,
		volLocks: map[string]*volLockInfo{},
	}

	return i.handle
}

type volLockInfo struct {
	MutexWithTryLock
	methodInErr map[string]bool
}

type idempotencyInterceptor struct {
	sync.Mutex
	p        IdempotencyProvider
	timeout  time.Duration
	volLocks map[string]*volLockInfo
}

func (i *idempotencyInterceptor) lockWithName(name string) *volLockInfo {
	i.Lock()
	defer i.Unlock()
	lock := i.volLocks[name]
	if lock == nil {
		lock = &volLockInfo{
			MutexWithTryLock: NewMutexWithTryLock(),
			methodInErr:      map[string]bool{},
		}
		i.volLocks[name] = lock
	}
	return lock
}

func (i *idempotencyInterceptor) handle(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	switch treq := req.(type) {
	case *csi.ControllerPublishVolumeRequest:
		return i.controllerPublishVolume(ctx, treq, info, handler)
	case *csi.ControllerUnpublishVolumeRequest:
		return i.controllerUnpublishVolume(ctx, treq, info, handler)
	case *csi.CreateVolumeRequest:
		return i.createVolume(ctx, treq, info, handler)
	case *csi.DeleteVolumeRequest:
		return i.deleteVolume(ctx, treq, info, handler)
	case *csi.NodePublishVolumeRequest:
		return i.nodePublishVolume(ctx, treq, info, handler)
	case *csi.NodeUnpublishVolumeRequest:
		return i.nodeUnpublishVolume(ctx, treq, info, handler)
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) controllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	name, err := i.p.GetVolumeName(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_VOLUME_DOES_NOT_EXIST,
			""), nil
	}

	lock := i.lockWithName(name)
	if !lock.TryLock(i.timeout) {
		return ErrControllerPublishVolume(
			csi.Error_ControllerPublishVolumeError_OPERATION_PENDING_FOR_VOLUME,
			""), nil
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil ||
			res.(*csi.ControllerPublishVolumeResponse).GetError() != nil {
			lock.methodInErr[info.FullMethod] = true
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if inErr, ok := lock.methodInErr[info.FullMethod]; ok && inErr {
		return handler(ctx, req)
	}

	pubInfo, err := i.p.IsControllerPublished(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if pubInfo != nil {
		log.WithField("name", name).Info("idempotent controller publish")
		return &csi.ControllerPublishVolumeResponse{
			Reply: &csi.ControllerPublishVolumeResponse_Result_{
				Result: &csi.ControllerPublishVolumeResponse_Result{
					PublishVolumeInfo: pubInfo,
				},
			},
		}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) controllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	name, err := i.p.GetVolumeName(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_VOLUME_DOES_NOT_EXIST,
			""), nil
	}

	lock := i.lockWithName(name)
	if !lock.TryLock(i.timeout) {
		return ErrControllerUnpublishVolume(
			csi.Error_ControllerUnpublishVolumeError_OPERATION_PENDING_FOR_VOLUME,
			""), nil
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil ||
			res.(*csi.ControllerUnpublishVolumeResponse).GetError() != nil {
			lock.methodInErr[info.FullMethod] = true
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if inErr, ok := lock.methodInErr[info.FullMethod]; ok && inErr {
		return handler(ctx, req)
	}

	pubInfo, err := i.p.IsControllerPublished(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if pubInfo == nil {
		log.WithField("name", name).Info("idempotent controller publish")
		return &csi.ControllerUnpublishVolumeResponse{
			Reply: &csi.ControllerUnpublishVolumeResponse_Result_{
				Result: &csi.ControllerUnpublishVolumeResponse_Result{},
			},
		}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) createVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	lock := i.lockWithName(req.Name)
	if !lock.TryLock(i.timeout) {
		return ErrCreateVolume(
			csi.Error_CreateVolumeError_OPERATION_PENDING_FOR_VOLUME,
			""), nil
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil ||
			res.(*csi.CreateVolumeResponse).GetError() != nil {
			lock.methodInErr[info.FullMethod] = true
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if inErr, ok := lock.methodInErr[info.FullMethod]; ok && inErr {
		return handler(ctx, req)
	}

	volInfo, err := i.p.GetVolumeInfo(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if volInfo != nil {
		log.WithField("name", req.Name).Info("idempotent create")
		return &csi.CreateVolumeResponse{
			Reply: &csi.CreateVolumeResponse_Result_{
				Result: &csi.CreateVolumeResponse_Result{
					VolumeInfo: volInfo,
				},
			},
		}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) deleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	name, err := i.p.GetVolumeName(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return ErrDeleteVolume(
			csi.Error_DeleteVolumeError_VOLUME_DOES_NOT_EXIST,
			""), nil
	}

	lock := i.lockWithName(name)
	if !lock.TryLock(i.timeout) {
		return ErrDeleteVolume(
			csi.Error_DeleteVolumeError_OPERATION_PENDING_FOR_VOLUME,
			""), nil
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil ||
			res.(*csi.DeleteVolumeResponse).GetError() != nil {
			lock.methodInErr[info.FullMethod] = true
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if inErr, ok := lock.methodInErr[info.FullMethod]; ok && inErr {
		return handler(ctx, req)
	}

	volInfo, err := i.p.GetVolumeInfo(ctx, name)
	if err != nil {
		return nil, err
	}

	if volInfo == nil {
		log.WithField("name", name).Info("idempotent delete")
		return &csi.DeleteVolumeResponse{
			Reply: &csi.DeleteVolumeResponse_Result_{
				Result: &csi.DeleteVolumeResponse_Result{},
			},
		}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) nodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	name, err := i.p.GetVolumeName(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_VOLUME_DOES_NOT_EXIST,
			""), nil
	}

	lock := i.lockWithName(name)
	if !lock.TryLock(i.timeout) {
		return ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_OPERATION_PENDING_FOR_VOLUME,
			""), nil
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil ||
			res.(*csi.NodePublishVolumeResponse).GetError() != nil {
			lock.methodInErr[info.FullMethod] = true
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if inErr, ok := lock.methodInErr[info.FullMethod]; ok && inErr {
		return handler(ctx, req)
	}

	ok, err := i.p.IsNodePublished(
		ctx, req.VolumeId, req.PublishVolumeInfo, req.TargetPath)
	if err != nil {
		return nil, err
	}
	if ok {
		log.WithField("name", name).Info("idempotent node publish")
		return &csi.NodePublishVolumeResponse{
			Reply: &csi.NodePublishVolumeResponse_Result_{
				Result: &csi.NodePublishVolumeResponse_Result{},
			},
		}, nil
	}

	return handler(ctx, req)
}

func (i *idempotencyInterceptor) nodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (res interface{}, resErr error) {

	name, err := i.p.GetVolumeName(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST,
			""), nil
	}

	lock := i.lockWithName(name)
	if !lock.TryLock(i.timeout) {
		return ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_OPERATION_PENDING_FOR_VOLUME,
			""), nil
	}

	// At the end of this function check for a response error or if
	// the response itself contains an error. If either is true then
	// mark the current method as in error.
	//
	// If neither is true then check to see if the method has been
	// marked in error in the past and remove that mark to reclaim
	// memory.
	defer func() {
		if resErr != nil ||
			res.(*csi.NodeUnpublishVolumeResponse).GetError() != nil {
			lock.methodInErr[info.FullMethod] = true
		} else if _, ok := lock.methodInErr[info.FullMethod]; ok {
			delete(lock.methodInErr, info.FullMethod)
		}
	}()
	defer lock.Unlock()

	// If the method has been marked in error then it means a previous
	// call to this function returned an error. In these cases a
	// subsequent call should bypass idempotency.
	if inErr, ok := lock.methodInErr[info.FullMethod]; ok && inErr {
		return handler(ctx, req)
	}

	ok, err := i.p.IsNodePublished(ctx, req.VolumeId, nil, req.TargetPath)
	if err != nil {
		return nil, err
	}
	if !ok {
		log.WithField("name", name).Info("idempotent node unpublish")
		return &csi.NodeUnpublishVolumeResponse{
			Reply: &csi.NodeUnpublishVolumeResponse_Result_{
				Result: &csi.NodeUnpublishVolumeResponse_Result{},
			},
		}, nil
	}

	return handler(ctx, req)
}
