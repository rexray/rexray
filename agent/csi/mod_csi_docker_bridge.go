package csi

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"
	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/gocsi/mount"

	apictx "github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	dvol "github.com/docker/go-plugins-helpers/volume"
)

const (
	dockerMountPath = "rexray.docker.mount.path"
	rrCSINFSVolumes = "rexray.csi.nfs.volumes"
)

func init() {
	registry.RegisterConfigReg(
		"Docker Bridge",
		func(ctx apitypes.Context, r gofig.ConfigRegistration) {
			pathConfig := apictx.MustPathConfig(ctx)
			r.Key(gofig.String,
				"", "", "",
				rrCSINFSVolumes,
				"csiNFSVolumes",
				"X_CSI_NFS_VOLUMES")

			r.Key(gofig.String, "",
				path.Join(pathConfig.Lib, "docker", "volumes"),
				"", dockerMountPath)
		})
}

var (
	nfsVols     string
	nfsVolsL    sync.Mutex
	nfsVolsOnce sync.Once
)

type dockerBridge struct {
	ctx    apitypes.Context
	config gofig.Config
	cs     *csiService

	fsType  string
	mntPath string

	byName    map[string]csi.VolumeInfo
	byNameRWL sync.RWMutex
}

func newDockerBridge(
	ctx apitypes.Context,
	config gofig.Config,
	cs *csiService) (*dockerBridge, error) {

	oldMntPath := config.GetString(apitypes.ConfigIgVolOpsMountPath)
	oldDatName := config.GetString(apitypes.ConfigIgVolOpsMountRootPath)
	newMntPath := config.GetString(dockerMountPath)

	if err := os.MkdirAll(newMntPath, 0755); err != nil {
		err := fmt.Errorf(
			"newDockerBridge: create new mount dir failed: %v", err)
		ctx.WithField("newMntPath", newMntPath).Error(err)
		return nil, err
	}

	// Migrate volumes with data directories from the previous mount
	// area to the new mount area.
	if err := bindMountOldDataDirs(
		ctx, oldMntPath, oldDatName, newMntPath); err != nil {
		err := fmt.Errorf(
			"newDockerBridge: bindMountOldDataDirs failed: %v", err)
		ctx.WithFields(map[string]interface{}{
			"oldMntPath": oldMntPath,
			"oldDatName": oldDatName,
			"newMntPath": newMntPath,
		}).Error(err)
		return nil, err
	}

	byName := map[string]csi.VolumeInfo{}
	nfsVolsOnce.Do(func() {
		nfsVols = path.Join(apictx.MustPathConfig(ctx).Lib, "csi-nfs-vol.map")
	})

	// Check to see if there are any CSI-NFS volume mappings.
	if isCSINFS(cs.serviceType) {
		if err := initNFSVolMap(ctx, config, nfsVols, byName); err != nil {
			return nil, err
		}
	}

	endpoint := &dockerBridge{
		ctx:     ctx,
		config:  config,
		cs:      cs,
		fsType:  config.GetString(apitypes.ConfigIgVolOpsCreateDefaultFsType),
		mntPath: newMntPath,
		byName:  byName,
	}

	ctx.Info("docker-csi-bridge: created Docker bridge endpoint")
	return endpoint, nil
}

func initNFSVolMap(
	ctx apitypes.Context,
	config gofig.Config,
	nfsVolsFilePath string,
	byName map[string]csi.VolumeInfo) (failed error) {

	r := strings.NewReplacer("/", "-", `\`, "-")

	splitNameHostExport := func(p string) {

		nameHostExport := strings.SplitN(p, "=", 2)
		if len(nameHostExport) != 2 {
			return
		}

		hostExport := strings.SplitN(nameHostExport[1], ":", 2)
		if len(hostExport) != 2 {
			return
		}

		var (
			nfsVolName = r.Replace(nameHostExport[0])
			nfsHost    = hostExport[0]
			nfsExport  = hostExport[1]
		)

		byName[nfsVolName] = csi.VolumeInfo{
			Id: &csi.VolumeID{
				Values: map[string]string{
					"host":   nfsHost,
					"export": nfsExport,
				},
			},
			Metadata: &csi.VolumeMetadata{
				Values: map[string]string{
					"name": nfsVolName,
				},
			},
		}

		ctx.WithFields(map[string]interface{}{
			"nfsHost":    nfsHost,
			"nfsExport":  nfsExport,
			"nfsVolName": nfsVolName,
		}).Debug("getNFSVolMap: cached prepopulated nfs volume")
	}

	// Check to see if there is a mappings file and use it, otherwise
	// look at the cofiguration property.
	if gotil.FileExists(nfsVolsFilePath) {
		f, err := os.Open(nfsVolsFilePath)
		if err != nil {
			return fmt.Errorf("getNFSVolMap: failed to open map file: %v", err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			splitNameHostExport(scanner.Text())
		}
	} else {
		// Check the config property.
		list := config.GetStringSlice(rrCSINFSVolumes)
		for _, l := range list {
			splitNameHostExport(l)
		}
	}

	return nil
}

func bindMountOldDataDirs(
	ctx apitypes.Context,
	oldMntPath, oldDatName, newMntPath string) error {

	lf := map[string]interface{}{
		"oldMntPath": oldMntPath,
		"oldDatName": oldDatName,
		"newMntPath": newMntPath,
	}

	// Create the glob pattern for the directories to return.
	globPatt := path.Join(oldMntPath, "*", oldDatName)

	// Get the directories to bind mount into the new location.
	fileNames, err := filepath.Glob(globPatt)
	if err != nil {
		ctx.WithFields(lf).Errorf("bindMountOldDataDirs: glob failed: %v", err)
		return err
	}

	for _, p := range fileNames {
		lfp := ctx.WithField("oldVolDatDir", p).WithFields(lf)
		lfp.Debug("bindMountOldDataDirs: processing")

		st, err := os.Stat(p)
		if err != nil {
			lfp.Errorf("bindMountOldDataDirs: stat old path failed: %v", err)
			return err
		}
		if !st.IsDir() {
			lfp.Debug("bindMountOldDataDirs: skip non-dir")
			continue
		}

		// Get the volume name and new mount path.
		volName := path.Base(path.Dir(p))
		newVolMntPath := path.Join(newMntPath, volName)
		lfn := ctx.WithField("newVolMntPath", newVolMntPath).WithFields(lf)

		if _, err := os.Stat(newVolMntPath); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(newVolMntPath, 0755); err != nil {
					lfn.Errorf("bindMountOldDataDirs: mkdir success: %v", err)
					continue
				}
				lfn.Debug("bindMountOldDataDirs: mkdir")
				if err := mount.BindMount(p, newVolMntPath); err != nil {
					lfn.Errorf("bindMountOldDataDirs: mount failed: %v", err)
					continue
				}
				lfn.Debug("bindMountOldDataDirs: mount success")
				continue
			}
			lfn.Errorf("bindMountOldDataDirs: stat new path failed: %v", err)
			return err
		}
	}
	return nil
}

func isCSINFS(serviceType string) bool {
	return strings.EqualFold(serviceType, "csi-nfs")
}

func (d *dockerBridge) isCSINFS() bool {
	return isCSINFS(d.cs.serviceType)
}

// cacheListResult caches the name-to-id mapping for a list of
// csi.VolumeInfo objects. This function replaces the existing list
// as the result of a ListVolumes RPC represents the most up-to-date
// view of the underlying storage platform
func (d *dockerBridge) cacheListResult(vols []*csi.VolumeInfo) {
	d.byNameRWL.Lock()
	defer d.byNameRWL.Unlock()

	// If this is not CSI-NFS then replace the existing volume info objects.
	if !d.isCSINFS() {
		d.byName = map[string]csi.VolumeInfo{}
	}

	for _, vi := range vols {
		if vi.Id == nil {
			continue
		}
		name := d.getName(*vi)
		if name == "" {
			d.ctx.Debugf("docker-csi-bridge: failed to cache id/name: %v", vi)
			continue
		}
		if _, ok := d.byName[name]; ok {
			// volume with same name exists already!
			d.ctx.Warnf("docker-csi-bridge: volume with name:%s already exists, skipping id/name: %v", name, vi)
			continue
		}
		d.byName[name] = *vi
	}
}

func (d *dockerBridge) getName(vi csi.VolumeInfo) string {
	if vi.Metadata != nil {
		if v := vi.Metadata.Values[mdKeyName]; v != "" {
			return v
		}
	}
	if strings.EqualFold(d.cs.serviceType, "libstorage") {
		return ""
	}
	return vi.Id.Values[idKeyID]
}

func (d *dockerBridge) getVolumeInfo(name string) (csi.VolumeInfo, bool) {
	d.byNameRWL.RLock()
	defer d.byNameRWL.RUnlock()
	vol, ok := d.byName[name]
	return vol, ok
}

func (d *dockerBridge) setVolumeInfo(name string, volInfo csi.VolumeInfo) {
	d.byNameRWL.Lock()
	defer d.byNameRWL.Unlock()
	d.byName[name] = volInfo
}

func (d *dockerBridge) delVolumeInfo(name string) {
	d.byNameRWL.Lock()
	defer d.byNameRWL.Unlock()
	delete(d.byName, name)
}

var (
	createParamCapabilities []*csi.VolumeCapability

	csiVersion = &csi.Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
)

const (
	idKeyID   = "id"
	mdKeyName = "name"

	errCodeCreateVolAlreadyExits = int32(
		csi.Error_CreateVolumeError_VOLUME_ALREADY_EXISTS)
	errCodeDeleteVolDoesNotExit = int32(
		csi.Error_DeleteVolumeError_VOLUME_DOES_NOT_EXIST)
	errCodeCtrlPubVolDoesNotExit = int32(
		csi.Error_ControllerPublishVolumeError_VOLUME_DOES_NOT_EXIST)
	errCodeCtrlUnpubVolDoesNotExit = int32(
		csi.Error_ControllerUnpublishVolumeError_VOLUME_DOES_NOT_EXIST)
	errCodeNodePubVolDoesNotExit = int32(
		csi.Error_NodePublishVolumeError_VOLUME_DOES_NOT_EXIST)
	errCodeNodeUnpubVolDoesNotExit = int32(
		csi.Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST)
)

func errIsVolAlreadyExists(err error) error {

	terr, ok := err.(*gocsi.Error)
	if !ok {
		return err
	}

	if terr.FullMethod == gocsi.FMCreateVolume &&
		terr.Code == errCodeCreateVolAlreadyExits {
		return nil
	}

	return err
}

func errIsVolDoesNotExist(err error) error {

	terr, ok := err.(*gocsi.Error)
	if !ok {
		return err
	}

	var exp int32 = -1

	switch terr.FullMethod {
	case gocsi.FMControllerPublishVolume:
		exp = errCodeCtrlPubVolDoesNotExit
	case gocsi.FMControllerUnpublishVolume:
		exp = errCodeCtrlUnpubVolDoesNotExit
	case gocsi.FMDeleteVolume:
		exp = errCodeDeleteVolDoesNotExit
	case gocsi.FMNodePublishVolume:
		exp = errCodeNodePubVolDoesNotExit
	case gocsi.FMNodeUnpublishVolume:
		exp = errCodeNodeUnpubVolDoesNotExit
	}

	if terr.Code == exp {
		return nil
	}

	return err
}

func errIsVolAttToNode(err error) error {

	terr, ok := err.(*gocsi.Error)
	if !ok {
		return err
	}

	var exp int32 = -1

	switch terr.FullMethod {
	case gocsi.FMNodePublishVolume:
		exp = errCodeNodePubVolDoesNotExit
	case gocsi.FMNodeUnpublishVolume:
		exp = errCodeNodeUnpubVolDoesNotExit
	}

	if terr.Code == exp {
		return nil
	}

	return err
}

func (d *dockerBridge) Create(req *dvol.CreateRequest) error {

	// If the service is CSI-NFS then create is handled differently.
	if d.isCSINFS() {
		var (
			nfsHost    string
			nfsExport  string
			nfsVolName = req.Name
		)

		for k, v := range req.Options {
			if strings.EqualFold(k, "host") {
				nfsHost = v
			} else if strings.EqualFold(k, "export") {
				nfsExport = v
			}
		}

		// Cache the mapping.
		nfsVolsL.Lock()
		defer nfsVolsL.Unlock()

		f, err := os.OpenFile(nfsVols, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = fmt.Fprintf(f, "%s=%s:%s\n", nfsVolName, nfsHost, nfsExport)
		if err != nil {
			return err
		}

		d.byNameRWL.Lock()
		defer d.byNameRWL.Unlock()

		d.byName[nfsVolName] = csi.VolumeInfo{
			Id: &csi.VolumeID{
				Values: map[string]string{
					"host":   nfsHost,
					"export": nfsExport,
				},
			},
			Metadata: &csi.VolumeMetadata{
				Values: map[string]string{
					"name": nfsVolName,
				},
			},
		}

		return nil
	}

	// Create a new gRPC, CSI client.
	c, err := d.cs.dial(d.ctx)
	if err != nil {
		d.ctx.Errorf("docker-csi-bridge: Create: client failure: %v", err)
		return err
	}
	defer c.Close()

	// Create a new CSI Controller client.
	cc := csi.NewControllerClient(c)

	// Check to see if the create option "size" is set.
	var (
		sizeGiB   int64
		sizeBytes uint64
	)
	for k, v := range req.Options {
		if strings.EqualFold(k, "size") {
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			sizeGiB = int64(i)

			// Translate size from GiB to bytes.
			if sizeGiB > 0 {
				sizeBytes = uint64(sizeGiB * 1024 * 1024 * 1024)
			}

			break
		}
	}

	// Call the CSI CreateVolume RPC.
	vol, err := gocsi.CreateVolume(
		d.ctx, cc, csiVersion,
		req.Name,
		sizeBytes, sizeBytes,
		createParamCapabilities,
		req.Options)

	// If there is an error, check to see if it is VOLUME_ALREADY_EXISTS.
	// If it is then the function below will return a nil value, otherwise
	// the original error is returned.
	if err != nil {
		return errIsVolAlreadyExists(err)
	}

	// Cache the volume.
	d.setVolumeInfo(req.Name, *vol)

	return nil
}

func (d *dockerBridge) List() (*dvol.ListResponse, error) {

	volMap := map[string]csi.VolumeInfo{}

	// If the service is CSI-NFS then grab volumes from cache, as that's
	// the only place they will be
	if d.isCSINFS() {
		d.byNameRWL.RLock()
		defer d.byNameRWL.RUnlock()

		for name, vi := range d.byName {
			d.ctx.WithField("volume", vi.Id.Values).WithField(
				"name", name).Debug(
				"docker-csi-bridge: List: found volume from cache")
			volMap[name] = vi
		}
		return d.buildListResponse(volMap)
	}

	// Create a new gRPC, CSI client.
	c, err := d.cs.dial(d.ctx)
	if err != nil {
		d.ctx.Errorf("docker-csi-bridge: List: client failure: %v", err)
		return nil, err
	}
	defer c.Close()

	// Create a new CSI Controller client.
	cc := csi.NewControllerClient(c)

	// Check if the CSI plugin supports ListVolumes
	caps, err := gocsi.ControllerGetCapabilities(d.ctx, cc, csiVersion)
	if err != nil {
		return nil, err
	}

	if !controllerListVolsSupported(caps) {
		return d.buildListResponse(volMap)
	}

	vols, _, err := gocsi.ListVolumes(d.ctx, cc, csiVersion, 0, "")
	if err != nil {
		d.ctx.Errorf("docker-csi-bridge: List: list volumes failed: %v", err)
		return nil, err
	}

	// Cache the list results in order to keep the name-to-id mappings
	// as up-to-date as possible.
	go d.cacheListResult(vols)

	for _, vi := range vols {
		if vi.Id == nil || len(vi.Id.Values) == 0 {
			d.ctx.Warn("docker-csi-bridge: List: skipped volume w missing id")
			continue
		}

		name := d.getName(*vi)
		if name == "" {
			d.ctx.WithField("volume", vi).Warn(
				"docker-csi-bridge: List: skipped volume w missing id and name")
			continue
		}

		if _, ok := volMap[name]; ok {
			d.ctx.WithField("volume", vi).Warn(
				"docker-csi-bridge: List: skipped volume with duplicate name")
			continue
		}

		//d.ctx.WithField("volume", vi.Id.Values).WithField(
		//	"name", name).Debug(
		//	"docker-csi-bridge: List: found new volume")
		volMap[name] = *vi
	}

	return d.buildListResponse(volMap)
}

func (d *dockerBridge) buildListResponse(
	volMap map[string]csi.VolumeInfo) (*dvol.ListResponse, error) {

	var (
		i   int
		res = &dvol.ListResponse{
			Volumes: make([]*dvol.Volume, len(volMap)),
		}
	)

	for name, vi := range volMap {
		vol, err := d.toDockerVolume(name, vi)
		if err != nil {
			return nil, fmt.Errorf(
				"docker-csi-bridge: buildListResponse: %s: %v", name, err)
		}
		res.Volumes[i] = vol
		i++
	}

	return res, nil
}

func (d *dockerBridge) Get(req *dvol.GetRequest) (*dvol.GetResponse, error) {

	volInfo, ok := d.getVolumeInfo(req.Name)
	if !ok {
		return nil, fmt.Errorf(
			"docker-csi-bridge: Get: unknown volume: %s", req.Name)
	}

	vol, err := d.toDockerVolume(req.Name, volInfo)
	if err != nil {
		return nil, fmt.Errorf(
			"docker-csi-bridge: Get: %s: %v", req.Name, err)
	}

	return &dvol.GetResponse{Volume: vol}, nil
}

// Remove the volume with the following steps:
//
// * Get volume from cache
// * DeleteVolume
// * Remove volume from cache
func (d *dockerBridge) Remove(req *dvol.RemoveRequest) (failed error) {

	// If the service is CSI-NFS then remove is handled differently.
	if d.isCSINFS() {

		// Make sure the volume is removed from the cache if this function
		// completes successfully.
		defer func() {
			if failed == nil {
				d.delVolumeInfo(req.Name)
			}
		}()

		// Remove the volume from the mappings file.
		nfsVolsL.Lock()
		defer nfsVolsL.Unlock()

		var lines []string

		if err := func() error {
			f, err := os.Open(nfsVols)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				l := scanner.Text()
				nameHostExport := strings.SplitN(l, "=", 2)
				if len(nameHostExport) != 2 {
					continue
				}
				nfsVolName := nameHostExport[0]
				if strings.EqualFold(req.Name, nfsVolName) {
					continue
				}
				lines = append(lines, l)
			}

			return nil
		}(); err != nil {
			return err
		}

		f, err := os.Create(nfsVols)
		if err != nil {
			return err
		}
		defer f.Close()

		for _, l := range lines {
			if _, err := fmt.Fprintln(f, l); err != nil {
				return err
			}
		}

		return nil
	}

	vol, ok := d.getVolumeInfo(req.Name)
	if !ok {
		return fmt.Errorf(
			"docker-csi-bridge: Remove: unknown volume: %s", req.Name)
	}

	// Make sure the volume is removed from the cache if this function
	// completes successfully.
	defer func() {
		if failed == nil {
			d.delVolumeInfo(req.Name)
		}
	}()

	// Create a new gRPC, CSI client.
	c, err := d.cs.dial(d.ctx)
	if err != nil {
		d.ctx.Errorf("docker-csi-bridge: Remove: client failure: %v", err)
		return err
	}
	defer c.Close()

	// Delete the volume using the Controller.
	if err := gocsi.DeleteVolume(
		d.ctx,
		csi.NewControllerClient(c),
		csiVersion, vol.Id, vol.Metadata); err != nil {

		// If there is an error, check to see if it is VOLUME_DOES_NOT_EXIST.
		// If it is then the function below will return a nil value, otherwise
		// the original error is returned.
		return errIsVolDoesNotExist(err)
	}

	return nil
}

func (d *dockerBridge) Path(req *dvol.PathRequest) (*dvol.PathResponse, error) {

	if _, ok := d.getVolumeInfo(req.Name); !ok {
		return nil, fmt.Errorf(
			"docker-csi-bridge: Path: unknown volume: %s", req.Name)
	}

	targetPath, ok, err := d.getTargetPath(req.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(
			"docker-csi-bridge: Path: volume not mounted: %s", req.Name)
	}

	return &dvol.PathResponse{Mountpoint: targetPath}, nil
}

var (
	refCounter  = map[string]int{}
	refCounterL sync.Mutex
)

func incRefCountFor(targetPath string) {
	refCounterL.Lock()
	defer refCounterL.Unlock()

	refCounter[targetPath] = refCounter[targetPath] + 1
}

func decRefCountFor(targetPath string) int {
	refCounterL.Lock()
	defer refCounterL.Unlock()

	v, ok := refCounter[targetPath]
	if ok && v > 0 {
		v--
		refCounter[targetPath] = v
	}
	return v
}

// Mount the volume with the following steps:
//
// * Get volume from cache
// * If volume does not exist, attempt to create it
// * Check to see if volume is already mounted
// * GetNodeID
// * ControllerPublishVolume
// * NodePublishVolume
// * Update cache with volume's new state
func (d *dockerBridge) Mount(
	req *dvol.MountRequest) (res *dvol.MountResponse, failed error) {

	// Create a new gRPC, CSI client.
	c, err := d.cs.dial(d.ctx)
	if err != nil {
		d.ctx.Errorf("docker-csi-bridge: Mount: client failure: %v", err)
		return nil, err
	}
	defer c.Close()

	// Create a new CSI Controller client.
	cc := csi.NewControllerClient(c)

	// Get the volume from the cache.
	vol, ok := d.getVolumeInfo(req.Name)

	// If the volume is not cached then create it.
	if !ok {
		if d.isCSINFS() {
			return nil, fmt.Errorf(
				"docker-csi-bridge: Mount: cannot implicitly create NFS volume: %s",
				req.Name)
		}
		newVol, err := gocsi.CreateVolume(
			d.ctx, cc, csiVersion,
			req.Name,
			0, 0,
			createParamCapabilities,
			nil)

		// If there's an error and it's not VOLUME_ALREADY_EXISTS then
		// fail this mount attempt.
		if errIsVolAlreadyExists(err) != nil {
			return nil, err
		}

		vol = *newVol
	}

	// Define the targetPath.
	targetPath, targetPathExists, err := d.getTargetPath(req.Name)
	if err != nil {
		return nil, err
	}

	// If this function exits without an error then increment
	// the ref cache for the target path.
	defer func() {
		if failed == nil {
			incRefCountFor(targetPath)
		}
	}()

	// Create the target directory.
	if !targetPathExists {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			d.ctx.WithField("targetPath", targetPath).Errorf(
				"docker-csi-bridge: Mount: create target path failed: %v", err)
		}
		d.ctx.WithField("targetPath", targetPath).Debug(
			"docker-csi-bridge: Mount: created target path")
	}

	// At this point it's known the volume is not mounted, so proceed
	// to do so:
	//
	// * GetNodeID
	// * ControllerPublishVolume
	// * NodePublishVolume
	// * Update cache with volume's new state

	// Create a new CSI Node client.
	nc := csi.NewNodeClient(c)

	// Check if the CSI plugin supports ControllerPublishVolume
	caps, err := gocsi.ControllerGetCapabilities(d.ctx, cc, csiVersion)
	if err != nil {
		return nil, err
	}

	var (
		volCap  *csi.VolumeCapability
		pubInfo *csi.PublishVolumeInfo
	)

	// Create a new volume capability for publishing the volume
	// via the Controller and Node.
	if d.isCSINFS() {
		volCap = newVolumeCapability("nfs")
	} else {
		volCap = newVolumeCapability(d.fsType)
	}

	if controllerPublishSupported(caps) {
		// Next, publish the volume at the Controller level. To do that this
		// Node's ID is required.
		nodeID, err := gocsi.GetNodeID(d.ctx, nc, csiVersion)
		if err != nil {
			return nil, err
		}

		// Publish the volume via the Controller.
		pubInfo, err = gocsi.ControllerPublishVolume(
			d.ctx, cc, csiVersion,
			vol.Id, vol.Metadata, nodeID,
			volCap, false)
		if err != nil {
			return nil, err
		}
	}

	// The target path of the volume is determined based on the
	// volume's name and the Docker-CSI bridge root path for
	// mounting volumes.
	targetPath = path.Join(d.mntPath, req.Name)

	// Publish the volume via the Node.
	if err := gocsi.NodePublishVolume(
		d.ctx, nc, csiVersion,
		vol.Id, vol.Metadata,
		pubInfo, targetPath,
		volCap, false); err != nil {

		return nil, err
	}

	return &dvol.MountResponse{Mountpoint: targetPath}, nil
}

func newVolumeCapability(
	fsType string, flags ...string) *csi.VolumeCapability {

	return &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{
				FsType:     fsType,
				MountFlags: flags,
			},
		},
	}
}

// Unmount the volume with the following steps:
//
// * Get volume from cache
// * Check to see if volume is already unmounted
// * GetNodeID
// * NodeUnpublishVolume
// * ControllerUnpublishVolume
// * Update cache with volume's new state
func (d *dockerBridge) Unmount(req *dvol.UnmountRequest) (failed error) {

	vol, ok := d.getVolumeInfo(req.Name)
	if !ok {
		return fmt.Errorf(
			"docker-csi-bridge: Unmount: unknown volume: %s", req.Name)
	}

	// Get the target path(s) to unpublish
	targetPath, _, _ := d.getTargetPath(req.Name)

	// Create a new gRPC, CSI client.
	c, err := d.cs.dial(d.ctx)
	if err != nil {
		d.ctx.Errorf("docker-csi-bridge: Unmount: client failure: %v", err)
		return err
	}
	defer c.Close()

	// Create a new CSI Node client.
	nc := csi.NewNodeClient(c)

	// First, unpublish the volume from this Node.
	if err := gocsi.NodeUnpublishVolume(
		d.ctx,
		nc,
		csiVersion,
		vol.Id,
		vol.Metadata,
		targetPath); err != nil {

		// If there is an error, check to see if it is VOLUME_DOES_NOT_EXIST.
		// If it is then the function below will return a nil value, otherwise
		// the original error is returned.
		return errIsVolDoesNotExist(err)
	}

	// Only progress further if there are no more Docker containers
	// using this target path.
	if v := decRefCountFor(targetPath); v > 0 {
		return nil
	}

	// If the function completes successfully the remove the target path.
	defer func() {
		if failed == nil {
			os.RemoveAll(targetPath)
		}
	}()

	// Create a new CSI Controller client.
	cc := csi.NewControllerClient(c)

	// Check if the CSI plugin supports ControllerPublishVolume
	caps, err := gocsi.ControllerGetCapabilities(d.ctx, cc, csiVersion)
	if err != nil {
		return err
	}

	// Next, unpublish the volume at the Controller level. To do that this
	// Node's ID is required.
	if controllerPublishSupported(caps) {
		nodeID, err := gocsi.GetNodeID(d.ctx, nc, csiVersion)
		if err != nil {
			return err
		}

		// Unpublish the volume at the Controller level.
		if err := gocsi.ControllerUnpublishVolume(
			d.ctx, cc, csiVersion, vol.Id, vol.Metadata, nodeID); err != nil {

			// If there is an error, check to see if it is VOLUME_DOES_NOT_EXIST.
			// If it is then the function below will return a nil value, otherwise
			// the original error is returned.
			return errIsVolDoesNotExist(err)
		}
	}

	return nil
}

func (d *dockerBridge) Capabilities() *dvol.CapabilitiesResponse {
	res := &dvol.CapabilitiesResponse{}
	res.Capabilities.Scope = "global"
	return res
}

func (d *dockerBridge) getTargetPath(volName string) (string, bool, error) {
	targetPath := path.Join(d.mntPath, volName)
	_, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return targetPath, false, nil
		}
		return targetPath, false, err
	}
	return targetPath, true, nil
}

func controllerCapSupported(
	caps []*csi.ControllerServiceCapability,
	cap csi.ControllerServiceCapability_RPC_Type) bool {

	for _, c := range caps {
		if c.GetRpc().GetType() == cap {
			return true
		}
	}
	return false
}

func controllerPublishSupported(caps []*csi.ControllerServiceCapability) bool {
	return controllerCapSupported(
		caps,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
}

func controllerListVolsSupported(caps []*csi.ControllerServiceCapability) bool {
	return controllerCapSupported(
		caps,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES)
}

func (d *dockerBridge) toDockerVolume(
	name string, volInfo csi.VolumeInfo) (*dvol.Volume, error) {

	vol := &dvol.Volume{Name: name}

	// Add the CSI VolumeInfo metadata to the Docker volume's
	// Status map.
	if volInfo.Metadata != nil && len(volInfo.Metadata.Values) > 0 {
		vol.Status = map[string]interface{}{}
		for k, v := range volInfo.Metadata.Values {
			vol.Status[k] = v
		}
	}

	// Include the volume's target path if it's mounted.
	targetPath, ok, err := d.getTargetPath(name)
	if err != nil {
		return nil, err
	}
	if ok {
		vol.Mountpoint = targetPath
	}

	return vol, nil
}
