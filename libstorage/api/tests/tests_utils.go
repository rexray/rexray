package tests

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

const copyFileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

func (t *testRunner) copyFile(dst, src string, perm os.FileMode) (err error) {
	var (
		in  *os.File
		out *os.File
	)
	if in, err = os.Open(src); err != nil {
		return
	}
	defer func() {
		err = in.Close()
	}()
	if out, err = os.OpenFile(dst, copyFileFlags, perm); err != nil {
		return
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	t.ctx.WithFields(map[string]interface{}{
		"src": src,
		"dst": dst,
	}).Debug("copied file")
	return
}

func newTempDir(d *string) {
	var err error
	Ω(*d).ShouldNot(BeADirectory())
	*d, err = ioutil.TempDir("", "")
	Ω(err).ToNot(HaveOccurred())
	Ω(*d).Should(BeADirectory())
}

func (t *testRunner) volCreateOpts() *types.VolumeCreateOpts {
	return &types.VolumeCreateOpts{Opts: t.store}
}

func (t *testRunner) volRemoveOpts() *types.VolumeRemoveOpts {
	return &types.VolumeRemoveOpts{Opts: t.store}
}

func (t *testRunner) volInsOpts(
	mask ...types.VolumeAttachmentsTypes) *types.VolumeInspectOpts {
	opts := &types.VolumeInspectOpts{Opts: t.store}
	if len(mask) == 1 {
		opts.Attachments = mask[0]
	}
	return opts
}

func (t *testRunner) volListOpts() *types.VolumesOpts {
	return &types.VolumesOpts{Opts: t.store}
}

func (t *testRunner) volAttOpts() *types.VolumeAttachOpts {
	return &types.VolumeAttachOpts{NextDevice: &t.nextDev, Opts: t.store}
}

func (t *testRunner) volDetOpts() *types.VolumeDetachOpts {
	return &types.VolumeDetachOpts{Opts: t.store}
}

func (t *testRunner) locDevOpts() *types.LocalDevicesOpts {
	return &types.LocalDevicesOpts{Opts: t.store}
}

func (t *testRunner) Α(vol *types.Volume, tok string, err error) {
	t.Θ(vol, err)
	Ω(tok).Should(Equal(t.expectedNextDev))
	Ω(vol.Attachments).Should(HaveLen(1))
}

func (t *testRunner) ΘΑ(vol *types.Volume, err error) {
	t.Θ(vol, err)
	Ω(vol.Attachments).Should(HaveLen(1))
}

func (t *testRunner) Θ(vol *types.Volume, err error) {
	Ω(err).ShouldNot(HaveOccurred())
	Ω(vol).ShouldNot(BeNil())
	Ω(vol.ID).Should(Equal(t.expectedVolID))
	Ω(vol.Name).Should(Equal(t.expectedVolName))
}

func (t *testRunner) Ξ(result interface{}, err error) {
	Ω(err).Should(HaveOccurred())
	Ω(result).Should(BeNil())
}

func (t *testRunner) Ψ(result interface{}, err error) {
	Ω(err).ShouldNot(HaveOccurred())
	Ω(result).Should(BeEmpty())
}

func (t *testRunner) Ε(err error) {
	Ω(err).ShouldNot(HaveOccurred())
}
