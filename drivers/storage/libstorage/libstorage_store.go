package libstorage

import "github.com/emccode/libstorage/api/types"

type lss struct {
	types.Store
}

func (s *lss) GetServiceInfo(service string) *types.ServiceInfo {
	if obj, ok := s.Get(service).(*types.ServiceInfo); ok {
		return obj
	}
	return nil
}

func (s *lss) GetExecutorInfo(lsx string) *types.ExecutorInfo {
	if obj, ok := s.Get(lsx).(*types.ExecutorInfo); ok {
		return obj
	}
	return nil
}

func (s *lss) GetInstanceID(driverName string) *types.InstanceID {
	return s.Store.GetInstanceID(driverName)
}
