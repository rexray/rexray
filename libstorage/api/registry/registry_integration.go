package registry

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	apiutils "github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

type idm struct {
	types.IntegrationDriver
	sync.RWMutex
	ctx        types.Context
	config     gofig.Config
	used       map[string]int
	retryCount int
	retryWait  time.Duration
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

	d.ctx = ctx
	d.config = config
	d.used = map[string]int{}
	d.retryCount = config.GetInt(types.ConfigIgVolOpsMountRetryCount)
	if v := config.GetString(types.ConfigIgVolOpsMountRetryWait); v != "" {
		var err error
		d.retryWait, err = time.ParseDuration(v)
		if err != nil {
			return err
		}
	}

	d.initPathCache(ctx)

	ctx.WithFields(log.Fields{
		types.ConfigIgVolOpsPathCacheEnabled:  d.pathCacheEnabled(),
		types.ConfigIgVolOpsPathCacheAsync:    d.pathCacheAsync(),
		types.ConfigIgVolOpsUnmountIgnoreUsed: d.ignoreUsedCount(),
		types.ConfigIgVolOpsMountPreempt:      d.preempt(),
		types.ConfigIgVolOpsCreateDisable:     d.disableCreate(),
		types.ConfigIgVolOpsRemoveDisable:     d.disableRemove(),
	}).Info("libStorage integration driver successfully initialized")

	return nil
}

var initPathCacheMap = map[string]interface{}{"attachments": true}

func (d *idm) initPathCache(ctx types.Context) {
	if !d.pathCacheEnabled() {
		ctx.Info("path cache initializion disabled")
		return
	}

	if name, ok := context.ServiceName(ctx); !ok || name == "" {
		ctx.Info("path cache initializion disabled; no service name in ctx")
		return
	}

	f := func(async bool) {
		ctx.WithField("async", async).Info("initializing the path cache")
		_, err := d.List(ctx, apiutils.NewStoreWithData(initPathCacheMap))
		if err != nil {
			ctx.WithField("async", async).WithError(err).Error(
				"error initializing the path cache")
		} else {
			ctx.WithField("async", async).Debug("initialized the path cache")
		}
	}

	if d.pathCacheAsync() {
		go f(true)
	} else {
		f(false)
	}
}

func (d *idm) List(
	ctx types.Context,
	opts types.Store) ([]types.VolumeMapping, error) {

	fields := log.Fields{
		"opts": opts}
	ctx.WithFields(fields).Debug("listing volumes")

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

	if !d.pathCacheEnabled() {
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

	fields := log.Fields{
		"volumeName": volumeName,
		"opts":       opts}
	ctx.WithFields(fields).Debug("inspecting volume")

	return d.IntegrationDriver.Inspect(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Mount(
	ctx types.Context,
	volumeID, volumeName string,
	opts *types.VolumeMountOpts) (string, *types.Volume, error) {

	opts.Preempt = d.preempt()

	fields := log.Fields{
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"opts":       opts}
	ctx.WithFields(fields).Debug("mounting volume")

	ctx = ctx.Join(d.ctx)

	mp, vol, err := d.IntegrationDriver.Mount(
		ctx, volumeID, volumeName, opts)
	if err != nil {
		for x := 0; x < d.retryCount; x++ {
			time.Sleep(d.retryWait)
			mp, vol, err = d.IntegrationDriver.Mount(
				ctx, volumeID, volumeName, opts)
			if err == nil {
				break
			}
		}
		if err != nil {
			return "", nil, err
		}
	}

	// if the volume has attachments assign the new mount point to the
	// MountPoint field of the first attachment element
	if len(vol.Attachments) > 0 {
		vol.Attachments[0].MountPoint = mp
	}

	d.incCount(volumeName)
	return mp, vol, err
}

func (d *idm) Unmount(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {

	fields := log.Fields{
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"opts":       opts}
	ctx.WithFields(fields).Debug("unmounting volume")

	if d.ignoreUsedCount() ||
		d.resetCount(volumeName) ||
		!d.isCounted(volumeName) {

		d.initCount(volumeName)
		return d.IntegrationDriver.Unmount(
			ctx.Join(d.ctx), volumeID, volumeName, opts)
	}

	d.decCount(volumeName)
	return nil, nil
}

func (d *idm) Path(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (string, error) {

	fields := log.Fields{
		"volumeName": volumeName,
		"volumeID":   volumeID,
		"opts":       opts}
	ctx.WithFields(fields).Debug("getting path to volume")

	if !d.pathCacheEnabled() {
		return d.IntegrationDriver.Path(
			ctx.Join(d.ctx), volumeID, volumeName, opts)
	}

	if !d.isCounted(volumeName) {
		ctx.WithFields(fields).Debug("skipping path lookup")
		return "", nil
	}

	return d.IntegrationDriver.Path(ctx.Join(d.ctx), volumeID, volumeName, opts)
}

func (d *idm) Create(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	fields := log.Fields{
		"volumeName": volumeName,
		"opts":       opts}
	ctx.WithFields(fields).Debug("creating volume")

	if d.disableCreate() {
		ctx.Debug("disableRemove skipped creation")
		return nil, nil
	}
	return d.IntegrationDriver.Create(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Remove(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeRemoveOpts) error {

	fields := log.Fields{
		"volumeName": volumeName,
		"opts":       opts}
	ctx.WithFields(fields).Debug("removing volume")

	if d.disableRemove() {
		ctx.Debug("disableRemove skipped deletion")
		return nil
	}
	return d.IntegrationDriver.Remove(ctx.Join(d.ctx), volumeName, opts)
}

func (d *idm) Attach(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeAttachOpts) (string, error) {

	fields := log.Fields{
		"volumeName": volumeName,
		"opts":       opts}
	ctx.WithFields(fields).Debug("attaching volume")

	return d.IntegrationDriver.Attach(ctx.Join(d.ctx), volumeName, opts)

}

func (d *idm) Detach(
	ctx types.Context,
	volumeName string,
	opts *types.VolumeDetachOpts) error {

	fields := log.Fields{
		"volumeName": volumeName,
		"opts":       opts}
	ctx.WithFields(fields).Debug("detaching volume")

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
		c = 1
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
	return d.config.GetBool(types.ConfigIgVolOpsMountPreempt)
}

func (d *idm) disableCreate() bool {
	return d.config.GetBool(types.ConfigIgVolOpsCreateDisable)
}

func (d *idm) disableRemove() bool {
	return d.config.GetBool(types.ConfigIgVolOpsRemoveDisable)
}

func (d *idm) ignoreUsedCount() bool {
	return d.config.GetBool(types.ConfigIgVolOpsUnmountIgnoreUsed)
}

func (d *idm) pathCacheEnabled() bool {
	return d.config.GetBool(types.ConfigIgVolOpsPathCacheEnabled)
}

func (d *idm) pathCacheAsync() bool {
	return d.config.GetBool(types.ConfigIgVolOpsPathCacheAsync)
}
