package v1

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/thecodeteam/goisilon/api"
)

// GetIsiSnapshots queries a list of all snapshots on the cluster
func GetIsiSnapshots(
	ctx context.Context,
	client api.Client) (resp *getIsiSnapshotsResp, err error) {
	// PAPI call: GET https://1.2.3.4:8080/platform/1/snapshot/snapshots
	err = client.Get(ctx, snapshotsPath, "", nil, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetIsiSnapshot queries an individual snapshot on the cluster
func GetIsiSnapshot(
	ctx context.Context,
	client api.Client,
	id int64) (*IsiSnapshot, error) {
	// PAPI call: GET https://1.2.3.4:8080/platform/1/snapshot/snapshots/123
	snapshotUrl := fmt.Sprintf("%s/%d", snapshotsPath, id)
	var resp *getIsiSnapshotsResp
	err := client.Get(ctx, snapshotUrl, "", nil, nil, &resp)
	if err != nil {
		return nil, err
	}
	// PAPI returns the snapshot data in a JSON list with the same structure as
	// when querying all snapshots.  Since this is for a single Id, we just
	// want the first (and should be only) entry in the list.
	return resp.SnapshotList[0], nil
}

// CreateIsiSnapshot makes a new snapshot on the cluster
func CreateIsiSnapshot(
	ctx context.Context,
	client api.Client,
	path, name string) (resp *IsiSnapshot, err error) {
	// PAPI call: POST https://1.2.3.4:8080/platform/1/snapshot/snapshots
	//            Content-Type: application/json
	//            {path: "/path/to/volume"
	//             name: "snapshot_name"  <--- optional
	//            }
	if path == "" {
		return nil, errors.New("no path set")
	}

	data := &SnapshotPath{Path: path}
	if name != "" {
		data.Name = name
	}

	err = client.Post(ctx, snapshotsPath, "", nil, nil, data, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CopyIsiSnaphost copies all files/directories in a snapshot to a new directory
func CopyIsiSnapshot(
	ctx context.Context,
	client api.Client,
	sourceSnapshotName, sourceVolume, destinationName string) (resp *IsiVolume, err error) {
	// PAPI calls: PUT https://1.2.3.4:8080/namespace/path/to/volumes/destination_volume_name
	//             x-isi-ifs-copy-source: /path/to/snapshot/volumes/source_volume_name

	headers := map[string]string{
		"x-isi-ifs-copy-source": path.Join(
			"/",
			realVolumeSnapshotPath(client, sourceSnapshotName),
			sourceVolume),
	}

	// copy the volume
	err = client.Put(ctx, realNamespacePath(client), destinationName, nil, headers, nil, &resp)

	return resp, err
}

// RemoveIsiSnapshot deletes a snapshot from the cluster
func RemoveIsiSnapshot(
	ctx context.Context,
	client api.Client,
	id int64) error {
	// PAPI call: DELETE https://1.2.3.4:8080/platform/1/snapshot/snapshots/123
	snapshotUrl := fmt.Sprintf("%s/%d", snapshotsPath, id)
	err := client.Delete(ctx, snapshotUrl, "", nil, nil, nil)

	return err
}
