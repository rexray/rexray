// +build !libstorage_storage_driver libstorage_storage_driver_azure

package storage

import (
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"hash"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/rubiojr/go-vhd/vhd"

	armCompute "github.com/Azure/azure-sdk-for-go/arm/compute"
	blobStorage "github.com/Azure/azure-sdk-for-go/storage"
	autorest "github.com/Azure/go-autorest/autorest/azure"

	"golang.org/x/crypto/pkcs12"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/azure"
	"github.com/codedellemc/libstorage/drivers/storage/azure/utils"
)

const (
	// Name of Blob service in URL
	blobServiceName = "blob"

	// Required by Azure blob name suffix
	vhdExtension = ".vhd"

	// Size 1GB
	size1GB int64 = 1024 * 1024 * 1024

	// Default new disk size
	defaultNewDiskSizeGB int32 = 128
)

type driver struct {
	name             string
	config           gofig.Config
	subscriptionID   string
	resourceGroup    string
	tenantID         string
	storageAccount   string
	storageAccessKey string
	container        string
	clientID         string
	clientSecret     string
	certPath         string
	useHTTPS         bool
}

func init() {
	registry.RegisterStorageDriver(azure.Name, newDriver)
}

func newDriver() types.StorageDriver {
	return &driver{name: azure.Name}
}

func (d *driver) Name() string {
	return d.name
}

// Init initializes the driver.
func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config
	d.tenantID = d.getTenantID()
	d.clientID = d.getClientID()
	if d.tenantID == "" || d.clientID == "" {
		context.Error(
			"tenantID or clientID are not set. Login will fail.")
	}
	d.clientSecret = d.getClientSecret()
	d.certPath = d.getCertPath()
	if d.clientSecret == "" && d.certPath == "" {
		context.Error(
			"clientSecret or certPath must be set for login.")
	}

	d.storageAccount = d.getStorageAccount()
	d.storageAccessKey = d.getStorageAccessKey()
	if d.storageAccount == "" || d.storageAccessKey == "" {
		context.Error(
			"storageAccount and storageAccessKey " +
				"are needed for correct work.")
	}
	d.container = d.getContainer()

	d.subscriptionID = d.getSubscriptionID()
	if d.subscriptionID == "" {
		context.Error("subscriptionID must be set for correct work.")
	}
	d.resourceGroup = d.getResourceGroup()
	if d.resourceGroup == "" {
		context.Warning("resourceGroup is not set." +
			" Some operations will fail.")
	}

	d.useHTTPS = d.getUseHTTPS()

	context.Info("storage driver initialized")

	return nil
}

const cacheKeyC = "cacheKey"

type azureSession struct {
	vmClient   *armCompute.VirtualMachinesClient
	blobClient *blobStorage.BlobStorageClient
}

var (
	sessions  = map[string]*azureSession{}
	sessionsL = &sync.Mutex{}
)

func writeHkeyB(h hash.Hash, ps []byte) {
	if ps == nil {
		return
	}
	h.Write(ps)
}

func writeHkey(h hash.Hash, ps *string) {
	writeHkeyB(h, []byte(*ps))
}

var (
	errLoginMsg           = "Failed to login to Azure"
	errAuthFailed         = goof.New(errLoginMsg)
	invalideRsaPrivateKey = goof.New(
		"PKCS#12 certificate must contain an RSA private key")
)

func decodePkcs12(
	pkcs []byte,
	password string) (*x509.Certificate, *rsa.PrivateKey, error) {

	privateKey, certificate, err := pkcs12.Decode(pkcs, password)
	if err != nil {
		return nil, nil, err
	}

	rsaPrivateKey, isRsaKey := privateKey.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil, nil, invalideRsaPrivateKey
	}

	return certificate, rsaPrivateKey, nil
}

func mustSession(ctx types.Context) *azureSession {
	return context.MustSession(ctx).(*azureSession)
}

func (d *driver) Login(ctx types.Context) (interface{}, error) {
	sessionsL.Lock()
	defer sessionsL.Unlock()

	ctx.Debug("login to azure storage driver")
	var (
		hkey     = md5.New()
		ckey     string
		certData []byte
		spt      *autorest.ServicePrincipalToken
		err      error
	)

	if d.tenantID == "" {
		return nil, goof.New("Empty tenantID")
	}

	writeHkey(hkey, &d.subscriptionID)
	writeHkey(hkey, &d.tenantID)
	writeHkey(hkey, &d.storageAccount)
	writeHkey(hkey, &d.storageAccessKey)
	if d.clientID != "" && d.clientSecret != "" {
		ctx.Debug("login to azure using clientID and clientSecret")
		writeHkey(hkey, &d.clientID)
		writeHkey(hkey, &d.clientSecret)
	} else if d.certPath != "" {
		ctx.Debug("login to azure using clientCert")
		certData, err = ioutil.ReadFile(d.certPath)
		if err != nil {
			return nil, goof.WithError(
				"Failed to read provided certificate file",
				err)
		}
		writeHkeyB(hkey, certData)
	} else {
		ctx.Error("No login information provided")
		return nil, errAuthFailed
	}
	ckey = fmt.Sprintf("%x", hkey.Sum(nil))

	if session, ok := sessions[ckey]; ok {
		ctx.WithField(cacheKeyC, ckey).Debug(
			"using cached azure client")
		return session, nil
	}

	oauthConfig, err := autorest.PublicCloud.OAuthConfigForTenant(
		d.tenantID)
	if err != nil {
		return nil, goof.WithError(
			"Failed to create OAuthConfig for tenant", err)
	}

	if d.clientID != "" && d.clientSecret != "" {
		spt, err = autorest.NewServicePrincipalToken(
			*oauthConfig, d.clientID, d.clientSecret,
			autorest.PublicCloud.ResourceManagerEndpoint)
		if err != nil {
			return nil, goof.WithError(
				"Failed to create Service Principal Token"+
					" with client ID and secret", err)
		}
	} else {
		certificate, rsaPrivateKey, err := decodePkcs12(certData, "")
		if err != nil {
			return nil, goof.WithError(
				"Failed to decode certificate data", err)
		}

		spt, err = autorest.NewServicePrincipalTokenFromCertificate(
			*oauthConfig, d.clientID, certificate,
			rsaPrivateKey,
			autorest.PublicCloud.ResourceManagerEndpoint)
		if err != nil {
			return nil, goof.WithError(
				"Failed to create Service Principal Token"+
					" with certificate ", err)
		}
	}

	newVMC := armCompute.NewVirtualMachinesClient(d.subscriptionID)
	newVMC.Authorizer = spt
	newVMC.PollingDelay = 5 * time.Second
	bc, err := blobStorage.NewBasicClient(
		d.storageAccount,
		d.storageAccessKey)
	if err != nil {
		return nil, goof.WithError(
			"Failed to create BlobStorage client", err)
	}
	newBC := bc.GetBlobService()
	session := azureSession{
		blobClient: &newBC,
		vmClient:   &newVMC,
	}
	sessions[ckey] = &session

	ctx.WithField(cacheKeyC, ckey).Info(
		"login to azure storage driver created and cached")

	return &session, nil
}

// NextDeviceInfo returns the information about the driver's next
// available device workflow.
func (d *driver) NextDeviceInfo(
	ctx types.Context) (*types.NextDeviceInfo, error) {
	return utils.NextDeviceInfo, nil
}

// Type returns the type of storage the driver provides.
func (d *driver) Type(ctx types.Context) (types.StorageType, error) {
	//Example: Block storage
	return types.Block, nil
}

// InstanceInspect returns an instance.
func (d *driver) InstanceInspect(
	ctx types.Context,
	opts types.Store) (*types.Instance, error) {

	iid := context.MustInstanceID(ctx)
	return &types.Instance{
		Name:         iid.ID,
		InstanceID:   iid,
		ProviderName: iid.Driver,
	}, nil
}

// Volumes returns all volumes or a filtered list of volumes.
func (d *driver) Volumes(
	ctx types.Context,
	opts *types.VolumesOpts) ([]*types.Volume, error) {

	list, err := mustSession(ctx).blobClient.ListBlobs(d.container,
		blobStorage.ListBlobsParameters{Include: "metadata"})
	if err != nil {
		return nil, goof.WithError("error listing blobs", err)
	}
	// Convert retrieved volumes to libStorage types.Volume
	vols, convErr := d.toTypesVolume(ctx, &list.Blobs, opts.Attachments)
	if convErr != nil {
		return nil, goof.WithError(
			"error converting to types.Volume", convErr)
	}
	return vols, nil
}

// VolumeInspect inspects a single volume.
func (d *driver) VolumeInspect(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeInspectOpts) (*types.Volume, error) {

	return d.getVolume(ctx, volumeID, opts.Attachments)
}

// VolumeCreate creates a new volume.
func (d *driver) VolumeCreate(ctx types.Context, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {

	if opts.Encrypted != nil && *opts.Encrypted {
		return nil, types.ErrNotImplemented
	}

	id, ok := context.InstanceID(ctx)
	if !ok || id == nil {
		return nil, goof.New(
			"Can't create volume outside of Azure instance")
	}
	vmName := id.ID

	if !strings.HasSuffix(volumeName, vhdExtension) {
		ctx.Warning("Disk name doesn't end with '.vhd'." +
			" This extension will be added automatically.")
		volumeName = volumeName + vhdExtension
	}

	size := int64(defaultNewDiskSizeGB)
	if opts.Size != nil && *opts.Size != 0 {
		size = *opts.Size
	}
	size *= size1GB

	fields := map[string]interface{}{
		"provider":      d.Name(),
		"vmName":        vmName,
		"volumeName":    volumeName,
		"size_in_bytes": size,
	}

	_, err := d.getVM(ctx, vmName)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"VM could not be obtained.", err)
	}

	blobClient := mustSession(ctx).blobClient
	err = d.createDiskBlob(volumeName, size, blobClient)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"failed to create volume for VM", err)
	}

	var volume *types.Volume
	volume, err = d.getVolume(ctx, volumeName, types.VolAttNone)
	if err != nil {
		d.deleteDiskBlob(volumeName, blobClient)
		return nil, goof.WithFieldsE(fields,
			"failed to get just created/attached volume", err)
	}

	return volume, nil
}

// VolumeCreateFromSnapshot creates a new volume from an existing snapshot.
func (d *driver) VolumeCreateFromSnapshot(
	ctx types.Context,
	snapshotID, volumeName string,
	opts *types.VolumeCreateOpts) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeCopy copies an existing volume.
func (d *driver) VolumeCopy(
	ctx types.Context,
	volumeID, volumeName string,
	opts types.Store) (*types.Volume, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeSnapshot snapshots a volume.
func (d *driver) VolumeSnapshot(
	ctx types.Context,
	volumeID, snapshotName string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// VolumeRemove removes a volume.
func (d *driver) VolumeRemove(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeRemoveOpts) error {

	//TODO check if volume is attached? if so fail

	_, err := mustSession(ctx).blobClient.DeleteBlobIfExists(
		d.container, volumeID, nil)
	if err != nil {
		fields := map[string]interface{}{
			"provider": d.Name(),
			"volumeID": volumeID,
		}
		return goof.WithFieldsE(fields, "error removing volume", err)
	}
	return nil
}

var (
	errMissingNextDevice  = goof.New("missing next device")
	errVolAlreadyAttached = goof.New("volume already attached to a host")
)

// VolumeAttach attaches a volume and provides a token clients can use
// to validate that device has appeared locally.
func (d *driver) VolumeAttach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeAttachOpts) (*types.Volume, string, error) {

	id, ok := context.InstanceID(ctx)
	if !ok || id == nil {
		return nil, "", goof.New(
			"Can't attach volume outside of Azure instance")
	}
	vmName := id.ID

	fields := map[string]interface{}{
		"provider":   d.Name(),
		"vmName":     vmName,
		"volumeID":   volumeID,
		"nextDevice": *opts.NextDevice,
	}

	volume, err := d.getVolume(ctx, volumeID,
		types.VolumeAttachmentsRequested)
	if err != nil {
		return nil, "", goof.WithFieldsE(fields,
			"failed to get volume for attach", err)
	}
	// Check if volume is already attached
	// TODO: maybe add the check that new instance is the same as current
	if len(volume.Attachments) > 0 {
		// Detach already attached volume if forced
		if !opts.Force {
			return nil, "", errVolAlreadyAttached
		}
		_, err = d.VolumeDetach(
			ctx,
			volumeID,
			&types.VolumeDetachOpts{
				Force: opts.Force,
				Opts:  opts.Opts,
			})
		if err != nil {
			return nil, "", goof.WithFieldsE(fields,
				"failed to detach volume first", err)
		}
	}

	if opts.NextDevice == nil {
		return nil, "", errMissingNextDevice
	}

	vm, err := d.getVM(ctx, vmName)
	if err != nil {
		return nil, "", goof.WithFieldsE(fields,
			"VM could not be obtained.", err)
	}

	err = d.attachDisk(ctx, volumeID, volume.Size*size1GB, vm)
	if err != nil {
		return nil, "", goof.WithFieldsE(fields,
			"failed to attach volume.", err)
	}

	volume, err = d.getVolume(ctx, volumeID,
		types.VolumeAttachmentsRequested)
	if err != nil {
		return nil, "", goof.WithFieldsE(fields,
			"failed to get just created/attached volume", err)
	}
	if len(volume.Attachments) == 0 {
		return nil, "", goof.WithFieldsE(fields,
			"volume is not attached to VM", err)
	}

	return volume, *opts.NextDevice, nil
}

var errVolAlreadyDetached = goof.New("volume already detached")

// VolumeDetach detaches a volume.
func (d *driver) VolumeDetach(
	ctx types.Context,
	volumeID string,
	opts *types.VolumeDetachOpts) (*types.Volume, error) {

	id, ok := context.InstanceID(ctx)
	if !ok || id == nil {
		return nil, goof.New(
			"Can't detach volume outside of Azure instance")
	}
	vmName := id.ID

	fields := map[string]interface{}{
		"provider": d.Name(),
		"vmName":   vmName,
		"volumeID": volumeID,
	}

	volume, err := d.getVolume(ctx, volumeID,
		types.VolumeAttachmentsRequested)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"failed to get volume", err)
	}
	if len(volume.Attachments) == 0 {
		return nil, errVolAlreadyDetached
	}

	vm, err := d.getVM(ctx, vmName)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"failed to detach volume to VM", err)
	}

	disks := *vm.StorageProfile.DataDisks
	for i, disk := range disks {
		// Disk is paged blob in Azure. So blob name is case sensitive.
		if disk.Name != nil && *disk.Name == volumeID {
			// found the disk
			disks = append(disks[:i], disks[i+1:]...)
			break
		}
	}
	newVM := armCompute.VirtualMachine{
		Location: vm.Location,
		VirtualMachineProperties: &armCompute.VirtualMachineProperties{
			StorageProfile: &armCompute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}

	_, err = mustSession(ctx).vmClient.CreateOrUpdate(d.resourceGroup,
		vmName, newVM, nil)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"failed to detach volume", err)
	}

	volume, err = d.getVolume(ctx, volumeID,
		types.VolumeAttachmentsRequested)
	if err != nil {
		return nil, goof.WithFieldsE(fields,
			"failed to get volume", err)
	}
	if len(volume.Attachments) != 0 {
		return nil, goof.WithFieldsE(fields,
			"volume is not detached", err)
	}
	return volume, nil
}

// Snapshots returns all volumes or a filtered list of snapshots.
func (d *driver) Snapshots(
	ctx types.Context,
	opts types.Store) ([]*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotInspect inspects a single snapshot.
func (d *driver) SnapshotInspect(
	ctx types.Context,
	snapshotID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotCopy copies an existing snapshot.
func (d *driver) SnapshotCopy(
	ctx types.Context,
	snapshotID, snapshotName, destinationID string,
	opts types.Store) (*types.Snapshot, error) {
	// TODO Snapshots are not implemented yet
	return nil, types.ErrNotImplemented
}

// SnapshotRemove removes a snapshot.
func (d *driver) SnapshotRemove(
	ctx types.Context,
	snapshotID string,
	opts types.Store) error {
	// TODO Snapshots are not implemented yet
	return types.ErrNotImplemented
}

// Get volume or snapshot name without config tag
func (d *driver) getPrintableName(name string) string {
	return strings.TrimPrefix(name, d.tag()+azure.TagDelimiter)
}

// Prefix volume or snapshot name with config tag
func (d *driver) getFullName(name string) string {
	if d.tag() != "" {
		return d.tag() + azure.TagDelimiter + name
	}
	return name
}

// Retrieve config arguments
func (d *driver) getSubscriptionID() string {
	return d.config.GetString(azure.ConfigAzureSubscriptionIDKey)
}

func (d *driver) getResourceGroup() string {
	return d.config.GetString(azure.ConfigAzureResourceGroupKey)
}

func (d *driver) getTenantID() string {
	return d.config.GetString(azure.ConfigAzureTenantIDKey)
}

func (d *driver) getStorageAccount() string {
	return d.config.GetString(azure.ConfigAzureStorageAccountKey)
}

func (d *driver) getStorageAccessKey() string {
	return d.config.GetString(azure.ConfigAzureStorageAccessKeyKey)
}

func (d *driver) getContainer() string {
	result := d.config.GetString(azure.ConfigAzureContainerKey)
	if result == "" {
		result = "vhds"
	}
	return result
}

func (d *driver) getClientID() string {
	return d.config.GetString(azure.ConfigAzureClientIDKey)
}

func (d *driver) getClientSecret() string {
	return d.config.GetString(azure.ConfigAzureClientSecretKey)
}

func (d *driver) getCertPath() string {
	return d.config.GetString(azure.ConfigAzureCertPathKey)
}

func (d *driver) getUseHTTPS() bool {
	return d.config.GetBool(azure.ConfigAzureUseHTTPSKey)
}

func (d *driver) tag() string {
	return d.config.GetString(azure.ConfigAzureTagKey)
}

// TODO rexrayTag
/*func (d *driver) rexrayTag() string {
  return d.config.GetString("azure.rexrayTag")
}*/

// var errGetLocDevs = goof.New("error getting local devices from context")

func (d *driver) toTypesVolume(
	ctx types.Context,
	blobs *[]blobStorage.Blob,
	attachments types.VolumeAttachmentsTypes) ([]*types.Volume, error) {

	/* TODO:
	var (
		ld *types.LocalDevices
		ldOK bool
	)

	if attachments.Devices() {
		// Get local devices map from context
		if ld, ldOK = context.LocalDevices(ctx); !ldOK {
			return nil, errGetLocDevs
		}
	}
	*/

	var volumesSD []*types.Volume
	for _, blob := range *blobs {
		volumeSD, err := d.toTypeVolume(ctx, &blob, attachments)
		if err != nil {
			ctx.WithError(err).Warning("Failed to convert volume")
		} else if volumeSD != nil {
			volumesSD = append(volumesSD, volumeSD)
		}
	}
	return volumesSD, nil
}

func (d *driver) toTypeVolume(
	ctx types.Context,
	blob *blobStorage.Blob,
	attachments types.VolumeAttachmentsTypes) (*types.Volume, error) {

	// Metadata can have these fileds:
	// microsoftazurecompute_resourcegroupname:trex
	// microsoftazurecompute_vmname:ttt
	// microsoftazurecompute_disktype:DataDisk (or OSDisk)
	// microsoftazurecompute_diskid:7d9df1c9-7b4f-41d4-a2e3-6870dfa573ba
	// microsoftazurecompute_diskname:ttt-20161221-130722
	// microsoftazurecompute_disksizeingb:50

	btype := blob.Metadata["microsoftazurecompute_disktype"]
	if btype == "" && !strings.HasSuffix(blob.Name, vhdExtension) {
		return nil, nil
	}
	attachState := types.VolumeAvailable
	bstate := "detached"
	if blob.Metadata["microsoftazurecompute_vmname"] != "" {
		bstate = "attached"
		attachState = types.VolumeAttached
	}
	var attachmentsSD []*types.VolumeAttachment
	if attachments.Requested() && attachState == types.VolumeAttached {
		attachedIID := types.InstanceID{
			ID:     blob.Metadata["microsoftazurecompute_vmname"],
			Driver: d.name,
		}
		if attachments.Mine() {
			id, ok := context.InstanceID(ctx)
			if !ok || id == nil {
				return nil, goof.New("Can't get isntance" +
					" ID to filter volume")
			}

			if id.ID == attachedIID.ID {
				attachmentsSD = append(
					attachmentsSD,
					&types.VolumeAttachment{
						InstanceID: &attachedIID,
						VolumeID:   blob.Name,
					})
			}
		} else {
			attachmentsSD = append(
				attachmentsSD,
				&types.VolumeAttachment{
					InstanceID: &attachedIID,
					VolumeID:   blob.Name,
				})
		}
	}

	volumeSD := &types.Volume{
		Name:            blob.Name,
		ID:              blob.Name,
		Status:          bstate,
		Type:            btype,
		Size:            blob.Properties.ContentLength / size1GB,
		AttachmentState: attachState,
		Attachments:     attachmentsSD,
		// TODO:
		//AvailabilityZone: *volume.AvailabilityZone,
		//Encrypted:        *volume.Encrypted,
	}

	// Some volume types have no IOPS, so we get nil in volume.Iops
	//if volume.Iops != nil {
	//	volumeSD.IOPS = *volume.Iops
	//}

	return volumeSD, nil
}

func (d *driver) diskURI(name string) string {
	scheme := "http"
	if d.useHTTPS {
		scheme = "https"
	}
	host := fmt.Sprintf("%s://%s.%s.%s", scheme, d.storageAccount,
		blobServiceName, autorest.PublicCloud.StorageEndpointSuffix)
	uri := fmt.Sprintf("%s/%s/%s", host, d.container, name)
	return uri
}

func (d *driver) getVM(ctx types.Context, name string) (
	*armCompute.VirtualMachine, error) {

	vm, err := mustSession(ctx).vmClient.Get(d.resourceGroup, name, "")
	if err != nil {
		fields := map[string]interface{}{
			"provider": d.Name(),
			"vmName":   name,
		}
		return nil, goof.WithFieldsE(
			fields, "failed to get virtual machine", err)
	}

	return &vm, nil
}

func (d *driver) getVolume(
	ctx types.Context,
	volumeID string,
	attachments types.VolumeAttachmentsTypes) (*types.Volume, error) {

	list, err := mustSession(ctx).blobClient.ListBlobs(d.container,
		blobStorage.ListBlobsParameters{
			Prefix:  volumeID,
			Include: "metadata"})
	if err != nil {
		return nil, goof.WithError("error listing blobs", err)
	}
	if len(list.Blobs) == 0 {
		return nil, goof.New("error to get volume")
	}
	// Convert retrieved volumes to libStorage types.Volume
	return d.toTypeVolume(ctx, &list.Blobs[0], attachments)
}

func (d *driver) createDiskBlob(
	name string,
	size int64,
	blobClient *blobStorage.BlobStorageClient) error {

	// create new blob
	vhdSize := size + vhd.VHD_HEADER_SIZE
	err := blobClient.PutPageBlob(d.container, name, vhdSize, nil)
	if err != nil {
		return goof.WithError("PageBlob could not be created.", err)
	}

	// add VHD signature
	h := vhd.CreateFixedHeader(uint64(size), &vhd.VHDOptions{})
	b := new(bytes.Buffer)
	err = binary.Write(b, binary.BigEndian, h)
	if err != nil {
		d.deleteDiskBlob(name, blobClient)
		return goof.WithError("Vhd header could not be created.", err)
	}
	header := b.Bytes()
	err = blobClient.PutPage(d.container, name, size, vhdSize-1,
		blobStorage.PageWriteTypeUpdate,
		header[:vhd.VHD_HEADER_SIZE], nil)
	if err != nil {
		d.deleteDiskBlob(name, blobClient)
		return goof.WithError(
			"Vhd header could not be updated in the blob.", err)
	}

	return nil
}

func (d *driver) deleteDiskBlob(
	blobName string,
	blobClient *blobStorage.BlobStorageClient) error {

	return blobClient.DeleteBlob(d.container, blobName, nil)
}

func (d *driver) getNextDiskLun(
	vm *armCompute.VirtualMachine) (int32, error) {

	// 64 is a max number of LUNs per VM
	used := make([]bool, 64)
	disks := *vm.StorageProfile.DataDisks
	for _, disk := range disks {
		if disk.Lun != nil {
			used[*disk.Lun] = true
		}
	}
	for k, v := range used {
		if !v {
			return int32(k), nil
		}
	}
	return -1, goof.New("Free Lun could not be found.")
}

func (d *driver) attachDisk(
	ctx types.Context,
	volumeName string,
	size int64,
	vm *armCompute.VirtualMachine) error {

	lun, err := d.getNextDiskLun(vm)
	if err != nil {
		return goof.WithError(
			"Could not find find an empty Lun to attach disk to.",
			err)
	}

	uri := d.diskURI(volumeName)
	disks := *vm.StorageProfile.DataDisks
	sizeGB := int32(size / size1GB)
	disks = append(disks,
		armCompute.DataDisk{
			Name:         &volumeName,
			Vhd:          &armCompute.VirtualHardDisk{URI: &uri},
			Lun:          &lun,
			CreateOption: armCompute.Attach,
			DiskSizeGB:   &sizeGB,
			// TODO:
			// Caching:      cachingMode,
		})
	newVM := armCompute.VirtualMachine{
		Location: vm.Location,
		VirtualMachineProperties: &armCompute.VirtualMachineProperties{
			StorageProfile: &armCompute.StorageProfile{
				DataDisks: &disks,
			},
		},
	}

	_, err = mustSession(ctx).vmClient.CreateOrUpdate(d.resourceGroup,
		*vm.Name, newVM, nil)
	if err != nil {
		detail := err.Error()
		if strings.Contains(detail,
			"Code=\"AcquireDiskLeaseFailed\"") {

			// if lease cannot be acquired, immediately detach
			// the disk and return the original error
			ctx.Info("failed to acquire disk lease, try detach")
			_, _ = d.VolumeDetach(ctx, volumeName, nil)
		}
		return goof.WithError("failed to attach volume to VM", err)
	}

	return nil
}
