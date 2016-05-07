package registry

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/types"
)

type idm struct {
	types.IntegrationDriver
	sync.RWMutex
	ctx    types.Context
	config gofig.Config
	used   map[string]int
}

// NewIntegrationDriverManager returns a new integration driver manager.
func NewIntegrationDriverManager(
	d types.IntegrationDriver) types.IntegrationDriver {
	return &idm{IntegrationDriver: d, used: map[string]int{}}
}

func (d *idm) Name() string {
	return d.IntegrationDriver.Name()
}

func (d *idm) Init(ctx types.Context, config gofig.Config) error {
	if err := d.IntegrationDriver.Init(ctx, config); err != nil {
		return err
	}

	d.config = config
	d.ctx = ctx
	d.used = map[string]int{}
	ctx.WithField("pathCache", "").Debug("checking volume path cache setting")
	return nil
}

func (d *idm) List(
	ctx types.Context,
	opts types.Store) ([]types.VolumeMapping, error) {

	volMaps, err := d.IntegrationDriver.List(ctx.Join(d.ctx), opts)
	if err != nil {
		return nil, err
	}

	volMapsWithNames := []types.VolumeMapping{}
	for _, vm := range volMaps {
		if vm.VolumeName() != "" {
			volMapsWithNames = append(volMapsWithNames, vm)
		}
	}

	if !d.pathCache() {
		return volMapsWithNames, nil
	}

	for _, vm := range volMapsWithNames {
		vmn := vm.VolumeName()
		if !d.isCounted(vmn) && vm.MountPoint() != "" {
			d.initCount(vmn)
		}
	}

	return volMapsWithNames, nil
}

func (d *idm) Inspect(
	ctx types.Context,
	volumeName string,
	opts types.Store) (types.VolumeMapping, error) {
	return d.IntegrationDriver.Inspect(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Mount(
	ctx types.Context,
	volumeID, volumeName string,
	opts *types.VolumeMountOpts) (string, *types.Volume, error) {

	mp, vol, err := d.IntegrationDriver.Mount(
		ctx.Join(d.ctx), volumeID, volumeName, opts)
	if err != nil {
		return "", nil, err
	}
	d.incCount(volumeName)
	return mp, vol, err
}

func (d *idm) Unmount(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) error {

	if d.ignoreUsedCount() ||
		d.resetCount(volumeName) ||
		!d.isCounted(volumeName) {

		d.initCount(volumeName)
		return d.IntegrationDriver.Unmount(
			ctx.Join(d.ctx), volumeID, volumeName, opts)
	}

	d.decCount(volumeName)
	return nil
}

func (d *idm) Path(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (string, error) {

	fields := log.Fields{
		"driverName": d.Name(),
		"volumeName": volumeName,
		"volumeID":   volumeID}

	if !d.pathCache() {
		return d.IntegrationDriver.Path(
			ctx.Join(d.ctx), volumeID, volumeName, opts)
	}

	if !d.isCounted(volumeName) {
		d.ctx.WithFields(fields).Debug("skipping path lookup")
		return "", nil
	}

	return d.IntegrationDriver.Path(ctx, volumeID, volumeName, opts)
}

func (d *idm) Create(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	if d.disableCreate() {
		return nil, nil
	}
	return d.IntegrationDriver.Create(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Remove(
	ctx types.Context,
	volumeName string,
	opts types.Store) error {
	if d.disableRemove() {
		return nil
	}
	return d.IntegrationDriver.Remove(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Attach(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeAttachOpts) (string, error) {
	return d.IntegrationDriver.Attach(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Detach(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeDetachOpts) error {
	return d.IntegrationDriver.Detach(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) initCount(volumeName string) {
	d.Lock()
	defer d.Unlock()
	d.used[volumeName] = 0
	d.ctx.WithFields(log.Fields{
		"volumeName": volumeName,
		"count":      0,
	}).Debug("init count")
}

func (d *idm) resetCount(volumeName string) bool {
	d.Lock()
	defer d.Unlock()
	c, ok := d.used[volumeName]
	if ok && c < 2 {
		d.ctx.WithFields(log.Fields{
			"volumeName": volumeName,
			"count":      c,
		}).Info("count reset")
		d.used[volumeName] = 0
		return true
	}
	return false
}

func (d *idm) addCount(volumeName string, delta int) {
	d.Lock()
	defer d.Unlock()
	c, ok := d.used[volumeName]
	if ok {
		c = c + delta
	} else {
		c = 0
	}
	d.used[volumeName] = c
	d.ctx.WithFields(log.Fields{
		"volumeName": volumeName,
		"count":      c,
	}).Debug("set count")
}

func (d *idm) isCounted(volumeName string) bool {
	d.RLock()
	defer d.RUnlock()
	_, ok := d.used[volumeName]
	return ok
}

func (d *idm) incCount(volumeName string) {
	d.addCount(volumeName, 1)
}

func (d *idm) decCount(volumeName string) {
	d.addCount(volumeName, -1)
}

func (d *idm) preempt() bool {
	return d.config.GetBool(types.ConfigVolMountPreempt)
}

func (d *idm) disableCreate() bool {
	return d.config.GetBool(types.ConfigVolCreateDisable)
}

func (d *idm) disableRemove() bool {
	return d.config.GetBool(types.ConfigVolRemoveDisable)
}

func (d *idm) ignoreUsedCount() bool {
	return d.config.GetBool(types.ConfigVolUnmountIgnoreUsed)
}

func (d *idm) pathCache() bool {
	return d.config.GetBool(types.ConfigVolPathCache)
}
