package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/types"
)

func TestVolumeObject(t *testing.T) {

	v := &types.Volume{
		ID:   "vol-000",
		Name: "Volume 000",
		Size: 378,
		Attachments: []*types.VolumeAttachment{
			&types.VolumeAttachment{
				InstanceID: &types.InstanceID{
					ID:     "iid-000",
					Driver: "vfs",
				},
				VolumeID:   "vol-000",
				DeviceName: "/dev/xvd000",
			},
		},
		Fields: map[string]string{
			"priority": "2",
			"owner":    "root@example.com",
		},
	}

	d, err := ValidateVolume(v)
	if d == nil {
		assert.NoError(t, err, string(d))
	} else {
		assert.NoError(t, err)
	}
}

func TestVolumeSchema(t *testing.T) {
	s := VolumeSchema

	d := []byte(`{
    "id": "vol-000",
    "name": "Volume 000",
    "size": 378,
    "attachments": [
        {
            "instanceID": {
               "id": "iid-000",
               "driver": "vfs"
            },
            "volumeID": "vol-000",
            "deviceName": "/dev/xvd00"
        }
    ],
    "fields": {
        "priority": "2",
        "owner": "root@example.com"
    }
}`)

	err := Validate(nil, s, d)
	assert.NoError(t, err)

	d = []byte(`{
    "id2": "vol-000",
    "name": "Volume 000",
    "size": 378,
    "attachments": [
        {
            "instanceID": {
               "id": "iid-000",
               "driver": "vfs"
            },
            "volumeID": "vol-000",
            "deviceName": "/dev/xvd00"
        }
    ],
    "fields": {
        "priority": "2",
        "owner": "root@example.com"
    }
}`)

	err = Validate(nil, s, d)
	assert.Error(t, err)
	assert.EqualError(t, err, `"#" must have property "id"`)

	d = []byte(`{
    "id": "vol-000",
    "name": "Volume 000",
    "size": 378,
    "attachments": [
        {
            "instanceID": {
               "id": null,
               "driver": null
            },
            "volumeID": "vol-000",
            "deviceName": "/dev/xvd00"
        }
    ],
    "fields": {
        "priority": "2",
        "owner": "root@example.com"
    }
}`)

	err = Validate(nil, s, d)
	assert.Error(t, err)

	errTxt1 := `"#/attachments/[0]/instanceID/id": must be of type "string"`
	errTxt2 := `"#/attachments/[0]/instanceID/driver": must be of type "string"`
	assert.True(t, err.Error() == errTxt1 || err.Error() == errTxt2)

	d = []byte(`{
    "id": "vol-000",
    "name": "Volume 000",
    "size": 378,
    "attachments": [
        {
            "instanceID": {
               "id": "iid-000",
               "driver": "vfs"
            },
            "volumeID": "vol-000",
            "deviceName": "/dev/xvd00"
        }
    ],
    "fields": {
        "priority": 2,
        "owner": "root@example.com"
    }
}`)

	err = Validate(nil, s, d)
	assert.Error(t, err)
	assert.EqualError(t, err, `"#/fields/priority": must be of type "string"`)
}

func TestSnapshotObject(t *testing.T) {

	s := &types.Snapshot{
		ID:         "snap-000",
		Name:       "Snapshot 000",
		VolumeID:   "vol-000",
		VolumeSize: 10240,
		StartTime:  1455826676,
		Fields: map[string]string{
			"sparse": "true",
			"region": "US",
		},
	}

	d, err := ValidateSnapshot(s)
	if d == nil {
		assert.NoError(t, err, string(d))
	} else {
		assert.NoError(t, err)
	}
}

func TestSnapshotSchema(t *testing.T) {
	s := SnapshotSchema

	d := []byte(`{
    "id": "snap-000",
    "name": "Snapshot-000",
    "description": "A snapshot of Volume-000 (vol-000)",
    "startTime": 1455826676,
    "volumeID": "vol-000",
    "volumeSize": 10240,
    "fields": {
        "sparse": "true",
        "region": "US"
    }
}`)

	err := Validate(nil, s, d)
	assert.NoError(t, err)

	d = []byte(`{
    "id2": "snap-000",
    "name": "Snapshot-000",
    "description": "A snapshot of Volume-000 (vol-000)",
    "startTime": 1455826676,
    "volumeID": "vol-000",
    "volumeSize": 10240,
    "fields": {
        "sparse": "true",
        "region": "US"
    }
}`)

	err = Validate(nil, s, d)
	assert.Error(t, err)
	assert.EqualError(t, err, `"#" must have property "id"`)

	d = []byte(`{
    "id": "snap-000",
    "name": "Snapshot-000",
    "description": "A snapshot of Volume-000 (vol-000)",
    "startTime": 1455826676,
    "volumeID": "vol-000",
    "volumeSize": "10240",
    "fields": {
        "sparse": "true",
        "region": "US"
    }
}`)

	err = Validate(nil, s, d)
	assert.Error(t, err)
	assert.EqualError(t, err, `"#/volumeSize": must be of type "number"`)

	d = []byte(`{
    "id": "snap-000",
    "name": "Snapshot-000",
    "description": "A snapshot of Volume-000 (vol-000)",
    "startTime": 1455826676,
    "volumeID": "vol-000",
    "volumeSize": 10240,
    "fields": {
        "sparse": true,
        "region": "US"
    }
}`)

	err = Validate(nil, s, d)
	assert.Error(t, err)
	assert.EqualError(t, err, `"#/fields/sparse": must be of type "string"`)
}

func TestVolumeCreateRequestObject(t *testing.T) {

	availabilityZone := "US"
	iops := int64(1000)
	size := int64(10240)
	volType := "gold"

	v := &types.VolumeCreateRequest{
		Name:             "Volume 001",
		AvailabilityZone: &availabilityZone,
		IOPS:             &iops,
		Size:             &size,
		Type:             &volType,
		Opts: map[string]interface{}{
			"priority": 2,
			"owner":    "root@example.com",
			"customData": map[string]int{
				"1": 1,
				"2": 2,
			},
		},
	}

	d, err := ValidateVolumeCreateRequest(v)
	if d == nil {
		assert.NoError(t, err, string(d))
	} else {
		assert.NoError(t, err)
	}
}

func TestVolumeSnapshotRequestObject(t *testing.T) {
	snapshotName := "snapshot1"
	opts := map[string]interface{}{
		"priority": 2,
	}

	s := &types.VolumeSnapshotRequest{
		SnapshotName: snapshotName,
		Opts:         opts,
	}

	d, err := ValidateVolumeSnapshotRequest(s)
	if d == nil {
		assert.NoError(t, err, string(d))
	} else {
		assert.NoError(t, err)
	}
}
