package tests

import (
	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	apiclient "github.com/AVENTER-UG/rexray/libstorage/client"
)

func (t *testRunner) justBeforeEachClientSpec() {
	t.client, t.err = apiclient.New(t.ctx, t.config)
	Ω(t.err).ToNot(HaveOccurred())
	Ω(t.client).ShouldNot(BeNil())
	Ω(t.client.API().ServerName()).To(Equal(t.server.Name()))
}

func (t *testRunner) itClientSpecListRootResources() {
	roots, err := t.client.API().Root(t.ctx)
	Ω(err).ToNot(HaveOccurred())
	Ω(roots).To(HaveLen(5))
}

func (t *testRunner) itClientSpecListVolumes() {
	vols, err := t.client.API().Volumes(t.ctx, 0)
	Ω(err).ToNot(HaveOccurred())
	Ω(vols).Should(HaveLen(1))
	Ω(vols[t.driverName]).Should(BeEmpty())
}

func (t *testRunner) beforeEachClientServiceSpec() {
	t.ctx = t.ctx.WithValue(context.ServiceKey, t.driverName)
}

func (t *testRunner) itClientSvcSpecListVolumes() {
	t.ctx = t.ctx.WithValue(context.ServiceKey, t.driverName)
	t.Ψ(t.client.Storage().Volumes(t.ctx, t.volListOpts()))
}

func (t *testRunner) itClientSvcSpecInspectMissingVolume() {
	t.ctx = t.ctx.WithValue(context.ServiceKey, t.driverName)
	t.Ξ(t.client.Storage().VolumeInspect(t.ctx, t.volName, t.volInsOpts()))
}

func (t *testRunner) itClientSvcSpecCreateVolume() {
	t.Θ(t.client.Storage().VolumeCreate(t.ctx, t.volName, t.volCreateOpts()))
	t.Ε(t.client.Storage().VolumeRemove(t.ctx, t.volID, t.volRemoveOpts()))
}

func (t *testRunner) itPreservesLocalDevicesEmptyVals() {
	ld, err := t.client.Executor().LocalDevices(t.ctx, t.locDevOpts())
	Ω(err).ShouldNot(HaveOccurred())
	Ω(ld).ShouldNot(BeNil())
	Ω(ld.DeviceMap).ShouldNot(BeNil())
	Ω(ld.DeviceMap).ShouldNot(HaveLen(0))
	preserved := false
	for _, v := range ld.DeviceMap {
		if v == "" {
			preserved = true
			break
		}
	}
	Ω(preserved).Should(BeTrue())
}
