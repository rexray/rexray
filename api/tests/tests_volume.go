package tests

func (t *testRunner) justBeforeEachVolumeSpec() {
	t.vol, t.err = t.client.Storage().VolumeCreate(
		t.ctx, t.volName, t.volCreateOpts())
	t.Θ(t.vol, t.err)
}

func (t *testRunner) afterEachVolumeSpec() {
	if t.vol == nil {
		return
	}
	t.Ε(t.client.Storage().VolumeRemove(t.ctx, t.vol.ID, t.volRemoveOpts()))
	t.vol = nil
}

func (t *testRunner) itVolumeSpecInspect() {
	t.Θ(t.client.Storage().VolumeInspect(t.ctx, t.vol.ID, t.volInsOpts()))
}

func (t *testRunner) itVolumeSpecDelete() {
	t.Ε(t.client.Storage().VolumeRemove(t.ctx, t.vol.ID, t.volRemoveOpts()))
	t.vol = nil
}

func (t *testRunner) itVolumeSpecAttach() {
	t.Α(t.client.Storage().VolumeAttach(t.ctx, t.vol.ID, t.volAttOpts()))
}

func (t *testRunner) justBeforeEachAttVolumeSpec() {
	t.vol, t.nextDev, t.err = t.client.Storage().VolumeAttach(
		t.ctx, t.vol.ID, t.volAttOpts())
	t.Α(t.vol, t.nextDev, t.err)
}

func (t *testRunner) itAttVolumeSpecInspect() {
	t.ΘΑ(t.client.Storage().VolumeInspect(t.ctx, t.vol.ID, t.volInsOpts(1)))
}

func (t *testRunner) itAttVolumeSpecInspectAvai() {
	t.Ξ(t.client.Storage().VolumeInspect(t.ctx, t.vol.ID, t.volInsOpts(17)))
}

func (t *testRunner) itAttVolumeSpecInspectAttOrAvai() {
	t.ΘΑ(t.client.Storage().VolumeInspect(t.ctx, t.vol.ID, t.volInsOpts(27)))
}

func (t *testRunner) itAttVolumeSpecDetach() {
	t.Θ(t.client.Storage().VolumeDetach(t.ctx, t.vol.ID, t.volDetOpts()))
}
