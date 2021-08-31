package libstorage

import "github.com/AVENTER-UG/rexray/libstorage/api/types"

type lss struct {
	types.Store
}

func (s *lss) GetServiceInfo(service string) *types.ServiceInfo {
	if obj, ok := s.Get(service).(*types.ServiceInfo); ok {
		return obj
	}
	return nil
}

func (s *lss) GetInstanceID(service string) *types.InstanceID {
	return s.Store.GetInstanceID(service)
}

func (s *lss) GetLSXSupported(driverName string) types.LSXSupportedOp {
	if obj, ok := s.Store.Get(driverName).(types.LSXSupportedOp); ok {
		return obj
	}
	return 0
}
