# The LibStorage Model

Friendly, portable, powerful

---

## InstanceID
An instance ID identifies a host to a remote storage platform.

### Go
```go
type InstanceID struct {
	// ID is the instance ID
	ID string `json:"id"`

	// Metadata is any extra information about the instance ID.
	Metadata interface{} `json:"metadata"`
}
```

### JSON
```json
{
    "id":       "",
    "metadata": {}
}
```

## Instance
The instance provides information about a storage object.

### Go
```go
type Instance struct {
    // The ID of the instance to which the object is connected.
	InstanceID *InstanceID `json:"instanceID"`

    // The name of the instance.
	Name string `json:"name"`

	// The name of the provider that owns the object.
	ProviderName string `json:"providerName"`

	// The region from which the object originates.
	Region string `json:"region"`
}
```

### JSON
```json
{
    "instanceID":   {
        "id":       "",
        "metadata": {}
    },
    "name":         "",
    "providerName": "",
    "region":       ""
}
```

## BlockDevice
A block device provides information about a block-storage device.

### Go
```go
type BlockDevice struct {

    // The name of the device.
	DeviceName string `json:"deviceName"`

    // The ID of the instance to which the device is connected.
	InstanceID *InstanceID `json:"instanceID"`

    // The name of the network on which the device resides.
	NetworkName string `json:"networkName"`

	// The name of the provider that owns the block device.
	ProviderName string `json:"providerName"`

    // The region from which the device originates.
	Region string `json:"region"`

    // The device status.
	Status string `json:"status"`

	// The ID of the volume for which the device is mounted.
	VolumeID string `json:"volumeID"`
}
```

### JSON
```json
{
    "deviceName":   "",
    "instanceID":   {
        "id":       "",
        "metadata": {}
    },
    "networkName":  "",
    "providerName": "",
    "region":       "",
    "status":       "",
    "volumeID":     ""
}
```

## Snapshot
A snapshot provides information about a storage-layer snapshot.

### Go
```go
type Snapshot struct {
    // A description of the snapshot.
	Description string `json:"description"`

	// The name of the snapshot.
	Name string `json:"name"`

	// The snapshot's ID.
	SnapshotID string `json:"snapshotID"`

	// The time at which the request to create the snapshot was submitted.
	StartTime string `json:"startTime"`

	// The status of the snapshot.
	Status string `json:"status"`

    // The ID of the volume to which the snapshot belongs.
	VolumeID string `json:"volumeID"`

    // The size of the volume to which the snapshot belongs.
	VolumeSize string `json:"volumeSize"`
}
```

### JSON
```json
{
    "description":  "",
    "name":         "",
    "snapshotID":   "",
    "startTime":    "",
    "status":       "",
    "volumeID":     "",
    "volumeSize":   ""
}
```

## Volume
A volume provides information about a storage volume.

### Go
```go
type Volume struct {
    // The volume's attachments.
	Attachments []*VolumeAttachment `json:"attachments"`

    // The availability zone for which the volume is available.
	AvailabilityZone string `json:"availabilityZone"`

    // The volume IOPs.
	IOPS int64 `json:"iops"`

	// The name of the volume.
	Name string `json:"name"`

    // The name of the network on which the volume resides.
	NetworkName string `json:"networkName"`

    // The size of the volume.
	Size string `json:"size"`

	// The volume status.
	Status string `json:"status"`

    // The volume ID.
	VolumeID string `json:"volumeID"`

    // The volume type.
	VolumeType string `json:"volumeType"`
}
```

### JSON
```json
{
    "attachments": [
        {
            "deviceName":   "",
            "instanceID":   {
                "id":       "",
                "metadata": {}
            },
            "status":       "",
            "volumeID":     ""
        }
    ],
    "availabilityZone":     "",
    "iops":                 0,
    "name":                 "",
    "networkName":          "",
    "size":                 "",
    "volumeID":             "",
    "volumeType":           ""
}
```

## VolumeAttachment
A volume attachment provides information about an object attached to a storage
volume.

### Go
```go
type VolumeAttachment struct {
    // The name of the device on which the volume to which the object is
	// attached is mounted.
	DeviceName string

    // The ID of the instance on which the volume to which the attachment
	// belongs is mounted.
	InstanceID *InstanceID

    // The status of the attachment.
	Status string

	// The ID of the volume to which the attachment belongs.
	VolumeID string
}
```

### JSON
```json
{
    "deviceName":   "",
    "instanceID":   {
        "id":       "",
        "metadata": {}
    },
    "status":       "",
    "volumeID":     ""
}
```
