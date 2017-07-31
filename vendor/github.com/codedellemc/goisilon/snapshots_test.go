package goisilon

import (
	"fmt"
	"testing"
)

func TestSnapshotsGet(t *testing.T) {

	snapshotPath := "test_snapshots_get_volume"
	snapshotName1 := "test_snapshots_get_snapshot_0"
	snapshotName2 := "test_snapshots_get_snapshot_1"

	// create the test volume
	_, err := client.CreateVolume(defaultCtx, snapshotPath)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, snapshotPath)

	// identify all snapshots on the cluster
	snapshotMap := make(map[int64]string)
	snapshots, err := client.GetSnapshots(defaultCtx)
	if err != nil {
		panic(err)
	}
	for _, snapshot := range snapshots {
		snapshotMap[snapshot.Id] = snapshot.Name
	}
	initialSnapshotCount := len(snapshots)

	// Add the test snapshots
	testSnapshot1, err := client.CreateSnapshot(
		defaultCtx, snapshotPath, snapshotName1)
	if err != nil {
		panic(err)
	}
	testSnapshot2, err := client.CreateSnapshot(
		defaultCtx, snapshotPath, snapshotName2)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.RemoveSnapshot(defaultCtx, testSnapshot1.Id, snapshotName1)
	defer client.RemoveSnapshot(defaultCtx, testSnapshot2.Id, snapshotName2)

	// get the updated snapshot list
	snapshots, err = client.GetSnapshots(defaultCtx)
	if err != nil {
		panic(err)
	}

	// verify that the new snapshots are there as well as all the old snapshots.
	if len(snapshots) != initialSnapshotCount+2 {
		panic(fmt.Sprintf("Incorrect number of snapshots.  Expected: %d Actual: %d\n", initialSnapshotCount+2, len(snapshots)))
	}
	// remove the original snapshots and add the new ones.  in the end, we
	// should only have the snapshots we just created and nothing more.
	for _, snapshot := range snapshots {
		if _, found := snapshotMap[snapshot.Id]; found == true {
			// this snapshot existed prior to the test start
			delete(snapshotMap, snapshot.Id)
		} else {
			// this snapshot is new
			snapshotMap[snapshot.Id] = snapshot.Name
		}
	}
	if len(snapshotMap) != 2 {
		panic(fmt.Sprintf("Incorrect number of new snapshots.  Expected: 2 Actual: %d\n", len(snapshotMap)))
	}
	if _, found := snapshotMap[testSnapshot1.Id]; found == false {
		panic(fmt.Sprintf("testSnapshot1 was not in the snapshot list\n"))
	}
	if _, found := snapshotMap[testSnapshot2.Id]; found == false {
		panic(fmt.Sprintf("testSnapshot2 was not in the snapshot list\n"))
	}

}

func TestSnapshotsGetByPath(t *testing.T) {

	snapshotPath1 := "test_snapshots_get_by_path_volume_1"
	snapshotPath2 := "test_snapshots_get_by_path_volume_2"
	snapshotName1 := "test_snapshots_get_by_path_snapshot_1"
	snapshotName2 := "test_snapshots_get_by_path_snapshot_2"
	snapshotName3 := "test_snapshots_get_by_path_snapshot_3"

	// create the two test volumes
	_, err := client.CreateVolume(defaultCtx, snapshotPath1)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, snapshotPath1)
	_, err = client.CreateVolume(defaultCtx, snapshotPath2)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, snapshotPath2)

	// identify all snapshots on the cluster
	snapshotMap := make(map[int64]string)
	snapshots, err := client.GetSnapshotsByPath(defaultCtx, snapshotPath1)
	if err != nil {
		panic(err)
	}
	for _, snapshot := range snapshots {
		snapshotMap[snapshot.Id] = snapshot.Name
	}
	initialSnapshotCount := len(snapshots)

	// Add the test snapshots
	testSnapshot1, err := client.CreateSnapshot(
		defaultCtx, snapshotPath1, snapshotName1)
	if err != nil {
		panic(err)
	}
	testSnapshot2, err := client.CreateSnapshot(
		defaultCtx, snapshotPath2, snapshotName2)
	if err != nil {
		panic(err)
	}
	testSnapshot3, err := client.CreateSnapshot(
		defaultCtx, snapshotPath1, snapshotName3)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.RemoveSnapshot(defaultCtx, testSnapshot1.Id, snapshotName1)
	defer client.RemoveSnapshot(defaultCtx, testSnapshot2.Id, snapshotName2)
	defer client.RemoveSnapshot(defaultCtx, testSnapshot3.Id, snapshotName3)

	// get the updated snapshot list
	snapshots, err = client.GetSnapshotsByPath(defaultCtx, snapshotPath1)
	if err != nil {
		panic(err)
	}

	// verify that the new snapshots in the given path are there as well
	// as all the old snapshots in that path.
	if len(snapshots) != initialSnapshotCount+2 {
		panic(fmt.Sprintf("Incorrect number of snapshots for path (%s).  Expected: %d Actual: %d\n", snapshotPath1, initialSnapshotCount+2, len(snapshots)))
	}
	// remove the original snapshots and add the new ones.  in the end, we
	// should only have the snapshots we just created and nothing more.
	for _, snapshot := range snapshots {
		if _, found := snapshotMap[snapshot.Id]; found == true {
			// this snapshot existed prior to the test start
			delete(snapshotMap, snapshot.Id)
		} else {
			// this snapshot is new
			snapshotMap[snapshot.Id] = snapshot.Name
		}
	}
	if len(snapshotMap) != 2 {
		panic(fmt.Sprintf("Incorrect number of new snapshots.  Expected: 2 Actual: %d\n", len(snapshotMap)))
	}
	if _, found := snapshotMap[testSnapshot1.Id]; found == false {
		panic(fmt.Sprintf("testSnapshot1 was not in the snapshot list\n"))
	}
	if _, found := snapshotMap[testSnapshot3.Id]; found == false {
		panic(fmt.Sprintf("testSnapshot3 was not in the snapshot list\n"))
	}
}

func TestSnapshotCreate(t *testing.T) {

	snapshotPath := "test_snapshot_create_volume"
	snapshotName := "test_snapshot_create_snapshot"

	// create the test volume
	_, err := client.CreateVolume(defaultCtx, snapshotPath)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, snapshotPath)

	// make sure the snapshot doesn't exist yet
	snapshot, err := client.GetSnapshot(defaultCtx, -1, snapshotName)
	if err == nil && snapshot != nil {
		panic(fmt.Sprintf("Snapshot (%s) already exists.\n", snapshotName))
	}

	// Add the test snapshot
	testSnapshot, err := client.CreateSnapshot(
		defaultCtx, snapshotPath, snapshotName)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.RemoveSnapshot(defaultCtx, testSnapshot.Id, snapshotName)

	// get the updated snapshot list
	snapshot, err = client.GetSnapshot(
		defaultCtx, testSnapshot.Id, snapshotName)
	if err != nil {
		panic(err)
	}
	if snapshot == nil {
		panic(fmt.Sprintf("Snapshot (%s) was not created.\n", snapshotName))
	}
	if snapshot.Name != snapshotName {
		panic(fmt.Sprintf("Snapshot name not set properly.  Expected: (%s) Actual: (%s)\n", snapshotName, snapshot.Name))
	}
	if snapshot.Path != client.API.VolumePath(snapshotPath) {
		panic(fmt.Sprintf("Snapshot path not set properly.  Expected: (%s) Actual: (%s)\n", snapshotPath, snapshot.Path))
	}
}

func TestSnapshotRemove(t *testing.T) {

	snapshotPath := "test_snapshot_remove_volume"
	snapshotName := "test_snapshot_remove_snapshot"

	// create the test volume
	_, err := client.CreateVolume(defaultCtx, snapshotPath)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, snapshotPath)

	// make sure the snapshot exists
	client.CreateSnapshot(defaultCtx, snapshotPath, snapshotName)
	snapshot, err := client.GetSnapshot(defaultCtx, -1, snapshotName)
	if err != nil {
		panic(err)
	}
	if snapshot == nil {
		panic(fmt.Sprintf("Test not setup properly.  No test snapshot (%s).", snapshotName))
	}

	// remove the snapshot
	err = client.RemoveSnapshot(defaultCtx, snapshot.Id, snapshotName)
	if err != nil {
		panic(err)
	}

	// make sure the snapshot was removed
	snapshot, err = client.GetSnapshot(defaultCtx, snapshot.Id, snapshotName)
	if err != nil {
		panic(err)
	}
	if snapshot != nil {
		panic(fmt.Sprintf("Snapshot (%s) was not removed.\n%+v\n", snapshotName, snapshot))
	}
}

func TestSnapshotCopy(t *testing.T) {

	sourceSnapshotPath := "test_snapshot_copy_src_volume"
	sourceSnapshotName := "test_snapshot_copy_src_snapshot"
	destinationVolume := "test_snapshot_copy_dst_volume"
	subdirectoryName := "test_snapshot_copy_sub_dir"
	sourceSubDirectory := fmt.Sprintf("%s/%s", sourceSnapshotPath, subdirectoryName)
	destinationSubDirectory := fmt.Sprintf("%s/%s", destinationVolume, subdirectoryName)

	// create the test volume
	_, err := client.CreateVolume(defaultCtx, sourceSnapshotPath)
	if err != nil {
		panic(err)
	}
	//	defer client.DeleteVolume(snapshotPath)
	// create a subdirectory in the test tvolume
	_, err = client.CreateVolume(defaultCtx, sourceSubDirectory)
	if err != nil {
		panic(err)
	}

	// make sure the snapshot doesn't exist yet
	snapshot, err := client.GetSnapshot(defaultCtx, -1, sourceSnapshotName)
	if err == nil && snapshot != nil {
		panic(fmt.Sprintf("Snapshot (%s) already exists.\n", sourceSnapshotName))
	}

	// Add the test snapshot
	testSnapshot, err := client.CreateSnapshot(
		defaultCtx, sourceSnapshotPath, sourceSnapshotName)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.RemoveSnapshot(defaultCtx, testSnapshot.Id, sourceSnapshotName)
	// remove the sub directory
	err = client.DeleteVolume(defaultCtx, sourceSubDirectory)
	if err != nil {
		panic(err)
	}

	// copy the snapshot to the destination volume
	copiedVolume, err := client.CopySnapshot(
		defaultCtx, testSnapshot.Id, testSnapshot.Name, destinationVolume)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, destinationVolume)

	if copiedVolume.Name != destinationVolume {
		panic(fmt.Sprintf("Copied volume has incorrect name.  Expected: (%s) Acutal: (%s)", destinationVolume, copiedVolume.Name))
	}

	// make sure the destination volume was created
	volume, err := client.GetVolume(defaultCtx, "", destinationVolume)
	if err != nil || volume == nil {
		panic(fmt.Sprintf("Destination volume: (%s) was not created.\n", destinationVolume))
	}
	// make sure the sub directory was also created
	subDirectory, err := client.GetVolume(defaultCtx, "", destinationSubDirectory)
	if err != nil {
		panic(fmt.Sprintf("Destination sub directory: (%s) was not created.\n", subdirectoryName))
	}
	if subDirectory.Name != destinationSubDirectory {
		panic(fmt.Sprintf("Sub directory has incorrect name.  Expected: (%s) Acutal: (%s)", destinationSubDirectory, subDirectory.Name))
	}
}
