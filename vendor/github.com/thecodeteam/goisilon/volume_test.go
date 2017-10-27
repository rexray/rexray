package goisilon

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	apiv2 "github.com/thecodeteam/goisilon/api/v2"
)

func TestVolumeList(*testing.T) {
	volumeName1 := "test_get_volumes_name1"
	volumeName2 := "test_get_volumes_name2"

	// identify all volumes on the cluster
	volumeMap := make(map[string]bool)
	volumes, err := client.GetVolumes(defaultCtx)
	if err != nil {
		panic(err)
	}
	for _, volume := range volumes {
		volumeMap[volume.Name] = true
	}
	initialVolumeCount := len(volumes)

	// Add the test volumes
	testVolume1, err := client.CreateVolume(defaultCtx, volumeName1)
	if err != nil {
		panic(err)
	}
	testVolume2, err := client.CreateVolume(defaultCtx, volumeName2)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.DeleteVolume(defaultCtx, volumeName1)
	defer client.DeleteVolume(defaultCtx, volumeName2)

	// get the updated volume list
	volumes, err = client.GetVolumes(defaultCtx)
	if err != nil {
		panic(err)
	}

	// verify that the new volumes are there as well as all the old volumes.
	if len(volumes) != initialVolumeCount+2 {
		panic(fmt.Sprintf("Incorrect number of volumes.  Expected: %d Actual: %d\n", initialVolumeCount+2, len(volumes)))
	}
	// remove the original volumes and add the new ones.  in the end, we
	// should only have the volumes we just created and nothing more.
	for _, volume := range volumes {
		if _, found := volumeMap[volume.Name]; found == true {
			// this volume existed prior to the test start
			delete(volumeMap, volume.Name)
		} else {
			// this volume is new
			volumeMap[volume.Name] = true
		}
	}
	if len(volumeMap) != 2 {
		panic(fmt.Sprintf("Incorrect number of new volumes.  Expected: 2 Actual: %d\n", len(volumeMap)))
	}
	if _, found := volumeMap[testVolume1.Name]; found == false {
		panic(fmt.Sprintf("testVolume1 was not in the volume list\n"))
	}
	if _, found := volumeMap[testVolume2.Name]; found == false {
		panic(fmt.Sprintf("testVolume2 was not in the volume list\n"))
	}

}

func TestVolumeGetCreate(*testing.T) {
	volumeName := "test_get_create_volume_name"

	// make sure the volume doesn't exist yet
	volume, err := client.GetVolume(defaultCtx, volumeName, volumeName)
	if err == nil && volume != nil {
		panic(fmt.Sprintf("Volume (%s) already exists.\n", volumeName))
	}

	// Add the test volume
	testVolume, err := client.CreateVolume(defaultCtx, volumeName)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.DeleteVolume(defaultCtx, testVolume.Name)

	// get the new volume
	volume, err = client.GetVolume(defaultCtx, volumeName, volumeName)
	if err != nil {
		panic(err)
	}
	if volume == nil {
		panic(fmt.Sprintf("Volume (%s) was not created.\n", volumeName))
	}
	if volume.Name != volumeName {
		panic(fmt.Sprintf("Volume name not set properly.  Expected: (%s) Actual: (%s)\n", volumeName, volume.Name))
	}
}

func TestVolumeDelete(*testing.T) {
	volumeName := "test_remove_volume_name"

	// make sure the volume exists
	client.CreateVolume(defaultCtx, volumeName)
	volume, err := client.GetVolume(defaultCtx, volumeName, volumeName)
	if err != nil {
		panic(err)
	}
	if volume == nil {
		panic(fmt.Sprintf("Test not setup properly.  No test volume (%s).", volumeName))
	}

	// remove the volume
	err = client.DeleteVolume(defaultCtx, volumeName)
	if err != nil {
		panic(err)
	}

	// make sure the volume was removed
	volume, err = client.GetVolume(defaultCtx, volumeName, volumeName)
	if err == nil {
		panic(fmt.Sprintf("Attempting to get a removed volume should return an error but returned nil"))
	}
	if volume != nil {
		panic(fmt.Sprintf("Volume (%s) was not removed.\n%+v\n", volumeName, volume))
	}
}

func TestVolumeCopy(*testing.T) {
	sourceVolumeName := "test_copy_source_volume_name"
	destinationVolumeName := "test_copy_destination_volume_name"
	subDirectoryName := "test_sub_directory"
	sourceSubDirectoryPath := fmt.Sprintf("%s/%s", sourceVolumeName, subDirectoryName)
	destinationSubDirectoryPath := fmt.Sprintf("%s/%s", destinationVolumeName, subDirectoryName)

	// make sure the destination volume doesn't exist yet
	destinationVolume, err := client.GetVolume(
		defaultCtx, destinationVolumeName, destinationVolumeName)
	if err == nil && destinationVolume != nil {
		panic(fmt.Sprintf("Volume (%s) already exists.\n", destinationVolumeName))
	}

	// Add the test volume
	sourceTestVolume, err := client.CreateVolume(defaultCtx, sourceVolumeName)
	if err != nil {
		panic(err)
	}
	// make sure we clean up when we're done
	defer client.DeleteVolume(defaultCtx, sourceTestVolume.Name)
	// add a sub directory to the source volume
	_, err = client.CreateVolume(defaultCtx, sourceSubDirectoryPath)
	if err != nil {
		panic(err)
	}

	// copy the source volume to the test volume
	destinationTestVolume, err := client.CopyVolume(
		defaultCtx, sourceVolumeName, destinationVolumeName)
	if err != nil {
		panic(err)
	}
	defer client.DeleteVolume(defaultCtx, destinationTestVolume.Name)
	// verify the copied volume is the same as the source volume
	if destinationTestVolume == nil {
		panic(fmt.Sprintf("Destination volume (%s) was not created.\n", destinationVolumeName))
	}
	if destinationTestVolume.Name != destinationVolumeName {
		panic(fmt.Sprintf("Destination volume name not set properly.  Expected: (%s) Actual: (%s)\n", destinationVolumeName, destinationTestVolume.Name))
	}
	// make sure the destination volume contains the sub-directory
	subTestVolume, err := client.GetVolume(
		defaultCtx, "", destinationSubDirectoryPath)
	if err != nil {
		panic(err)
	}
	// verify the copied subdirectory is the same as int the source volume
	if subTestVolume == nil {
		panic(fmt.Sprintf("Destination sub directory (%s) was not created.\n", subDirectoryName))
	}
	if subTestVolume.Name != destinationSubDirectoryPath {
		panic(fmt.Sprintf("Destination sub directory name not set properly.  Expected: (%s) Actual: (%s)\n", destinationSubDirectoryPath, subTestVolume.Name))
	}

}

func TestVolumeExport(*testing.T) {
	// TODO: Make this more robust
	_, err := client.ExportVolume(defaultCtx, "testing")
	if err != nil {
		panic(err)
	}

}

func TestVolumeUnexport(*testing.T) {
	// TODO: Make this more robust
	err := client.UnexportVolume(defaultCtx, "testing")
	if err != nil {
		panic(err)
	}

}

func TestVolumePath(*testing.T) {
	// TODO: Make this more robust
	fmt.Println(client.API.VolumePath("testing"))
}

func TestVolumeGetExportMap(t *testing.T) {
	// TODO: Make this more robust
	volExMap, err := client.GetVolumeExportMap(defaultCtx, false)
	assertNoError(t, err)
	for v := range volExMap {
		t.Logf("volName=%s, volPath=%s", v.Name, client.API.VolumePath(v.Name))
	}
}

func TestVolumeQueryChildren(t *testing.T) {

	var (
		ctx = defaultCtx
		//context.WithValue(defaultCtx, log.LevelKey(), log.InfoLevel)

		err      error
		volume   Volume
		children VolumeChildrenMap

		volumeName   = "test_volume_query_children"
		dirPath0     = "dA"
		dirPath0a    = "dA/dAA"
		dirPath1     = "dA/dAA/dAAA"
		dirPath2     = "dB/dBB"
		dirPath3     = "dC"
		dirPath4     = "dC/dCC"
		fileName0    = "an_empty_file"
		fileName1    = fileName0
		fileName2    = fileName0
		fileName3    = fileName0
		volDirPath0  = path.Join(volumeName, dirPath0)
		volDirPath0a = path.Join(volumeName, dirPath0a)
		volDirPath1  = path.Join(volumeName, dirPath1)
		volDirPath2  = path.Join(volumeName, dirPath2)
		volDirPath3  = path.Join(volumeName, dirPath3)
		volDirPath4  = path.Join(volumeName, dirPath4)
		volFilePath0 = path.Join(volumeName, fileName0)
		volFilePath1 = path.Join(volDirPath2, fileName1)
		volFilePath2 = path.Join(volDirPath3, fileName2)
		volFilePath3 = path.Join(volDirPath4, fileName3)
		dirPath0Key  = path.Join(client.API.VolumePath(volumeName), dirPath0)
		dirPath0aKey = path.Join(client.API.VolumePath(volumeName), dirPath0a)
		dirPath1Key  = path.Join(client.API.VolumePath(volumeName), dirPath1)
		dirPath2Key  = path.Join(client.API.VolumePath(volumeName), dirPath2)
		dirPath3Key  = path.Join(client.API.VolumePath(volumeName), dirPath3)
		dirPath4Key  = path.Join(client.API.VolumePath(volumeName), dirPath4)
		filePath0Key = path.Join(client.API.VolumePath(volumeName), fileName0)
		filePath1Key = path.Join(dirPath2Key, fileName1)
		filePath2Key = path.Join(dirPath3Key, fileName2)
		filePath3Key = path.Join(dirPath4Key, fileName3)

		newUserName  = client.API.User()
		newGroupName = newUserName
		badUserID    = "999"
		badGroupID   = "999"
		badUserName  = "Unknown User"
		badGroupName = "Unknown Group"

		volChildCount = 9

		newDirMode  = apiv2.FileMode(0755)
		newFileMode = apiv2.FileMode(0644)
		badDirMode  = apiv2.FileMode(0700)
		badFileMode = apiv2.FileMode(0400)

		newDirACL = &apiv2.ACL{
			Action:        &apiv2.PActionTypeReplace,
			Authoritative: &apiv2.PAuthoritativeTypeMode,
			Owner: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   newUserName,
					Type: apiv2.PersonaIDTypeUser,
				},
			},
			Group: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   newGroupName,
					Type: apiv2.PersonaIDTypeGroup,
				},
			},
			Mode: &newDirMode,
		}

		newFileACL = &apiv2.ACL{
			Action:        &apiv2.PActionTypeReplace,
			Authoritative: &apiv2.PAuthoritativeTypeMode,
			Owner: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   newUserName,
					Type: apiv2.PersonaIDTypeUser,
				},
			},
			Group: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   newGroupName,
					Type: apiv2.PersonaIDTypeGroup,
				},
			},
			Mode: &newFileMode,
		}

		badDirACL = &apiv2.ACL{
			Action:        &apiv2.PActionTypeReplace,
			Authoritative: &apiv2.PAuthoritativeTypeMode,
			Owner: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   badUserID,
					Type: apiv2.PersonaIDTypeUID,
				},
			},
			Group: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   badGroupID,
					Type: apiv2.PersonaIDTypeGID,
				},
			},
			Mode: &badDirMode,
		}

		badFileACL = &apiv2.ACL{
			Action:        &apiv2.PActionTypeReplace,
			Authoritative: &apiv2.PAuthoritativeTypeMode,
			Owner: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   badUserID,
					Type: apiv2.PersonaIDTypeUID,
				},
			},
			Group: &apiv2.Persona{
				ID: &apiv2.PersonaID{
					ID:   badGroupID,
					Type: apiv2.PersonaIDTypeGID,
				},
			},
			Mode: &badFileMode,
		}
	)

	defer client.ForceDeleteVolume(ctx, volumeName)

	setACLsWithPaths := func(
		ctx context.Context, acl *apiv2.ACL, paths ...string) {
		for _, p := range paths {
			assertNoError(
				t,
				apiv2.ACLUpdate(ctx, client.API, p, acl))
		}
	}

	assertNewFileACL := func(cm map[string]*apiv2.ContainerChild, k string) {
		if !assert.Equal(t, newFileMode, *cm[k].Mode) ||
			!assert.Equal(t, newUserName, *cm[k].Owner) ||
			!assert.Equal(t, newGroupName, *cm[k].Group) {
			t.FailNow()
		}
	}

	assertBadFileACL := func(cm map[string]*apiv2.ContainerChild, k string) {
		if !assert.Equal(t, badFileMode, *cm[k].Mode) ||
			!assert.Equal(t, badUserName, *cm[k].Owner) ||
			!assert.Equal(t, badGroupName, *cm[k].Group) {
			t.FailNow()
		}
	}

	assertNewDirACL := func(cm map[string]*apiv2.ContainerChild, k string) {
		if !assert.Equal(t, newDirMode, *cm[k].Mode) ||
			!assert.Equal(t, newUserName, *cm[k].Owner) ||
			!assert.Equal(t, newGroupName, *cm[k].Group) {
			t.FailNow()
		}
	}

	assertBadDirACL := func(cm map[string]*apiv2.ContainerChild, k string) {
		if !assert.Equal(t, badDirMode, *cm[k].Mode) ||
			!assert.Equal(t, badUserName, *cm[k].Owner) ||
			!assert.Equal(t, badGroupName, *cm[k].Group) {
			t.FailNow()
		}
	}

	createObjs := func(ctx context.Context, createType int) {
		// make sure the volume exists
		client.CreateVolume(ctx, volumeName)
		volume, err = client.GetVolume(ctx, volumeName, volumeName)
		assertNoError(t, err)
		assertNotNil(t, volume)

		// assert the volume has no children
		children, err = client.QueryVolumeChildren(ctx, volumeName)
		assertNoError(t, err)
		assertLen(t, children, 0)

		switch createType {
		case 0:
			// create dirPath1
			assertError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath1,
				os.FileMode(newDirMode),
				false,
				false))

			// create dirPath1 again, recursively
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath1,
				os.FileMode(newDirMode),
				false,
				true))

			// create the second directory path
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath2,
				os.FileMode(newDirMode),
				true,
				true))

			// create file0
			assertNoError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volumeName,
				fileName0,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				false))

			// create file1
			assertNoError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volDirPath2,
				fileName1,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				false))

			// try and create file1 again; should fail
			assertError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volDirPath2,
				fileName1,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				false))

			// try and create file1 again; should work
			assertNoError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volDirPath2,
				fileName1,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				true))

			// create the third directory path
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath4,
				os.FileMode(newDirMode),
				true,
				true))

			setACLsWithPaths(
				ctx, newDirACL,
				volDirPath0, volDirPath1, volDirPath2, volDirPath3, volDirPath4)
			setACLsWithPaths(ctx, newFileACL, volFilePath0, volFilePath1)
			children, err = client.QueryVolumeChildren(ctx, volumeName)
			assertNoError(t, err)
			assertLen(t, children, volChildCount)
			assertNewDirACL(children, dirPath0Key)
			assertNewDirACL(children, dirPath1Key)
			assertNewDirACL(children, dirPath2Key)
			assertNewDirACL(children, dirPath3Key)
			assertNewDirACL(children, dirPath4Key)
			assertNewFileACL(children, filePath0Key)
			assertNewFileACL(children, filePath1Key)
		case 1:
			// test a single root dir with good perms that has an empty sub-dir
			// with bad perms
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath4,
				os.FileMode(newDirMode),
				true,
				true))
			setACLsWithPaths(ctx, newDirACL, volDirPath3, volDirPath4)
			children, err = client.QueryVolumeChildren(ctx, volumeName)
			assertNoError(t, err)
			assertLen(t, children, 2)
			assertNewDirACL(children, dirPath3Key)
			assertNewDirACL(children, dirPath4Key)
		case 2:
			// test a single, empty root dir with bad perms
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath3,
				os.FileMode(newDirMode),
				true,
				true))
			setACLsWithPaths(ctx, newDirACL, volDirPath3)
			children, err = client.QueryVolumeChildren(ctx, volumeName)
			assertNoError(t, err)
			assertLen(t, children, 1)
			assertNewDirACL(children, dirPath3Key)
		case 3:
			// test a single, root dir with bad perms that has a single file in
			// it where the file has good perms
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath3,
				os.FileMode(newDirMode),
				true,
				true))
			assertNoError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volDirPath3,
				fileName2,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				true))
			setACLsWithPaths(ctx, newFileACL, volFilePath2)
			setACLsWithPaths(ctx, newDirACL, volDirPath3)
			children, err = client.QueryVolumeChildren(ctx, volumeName)
			assertNoError(t, err)
			assertLen(t, children, 2)
			assertNewDirACL(children, dirPath3Key)
			assertNewFileACL(children, filePath2Key)
		case 4:
			// test a single, root dir with bad perms that has a single sub-dir
			// with good perms, and the sub-dir contains a file with good perms
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath4,
				os.FileMode(newDirMode),
				true,
				true))
			assertNoError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volDirPath4,
				fileName3,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				true))
			setACLsWithPaths(ctx, newFileACL, volFilePath3)
			setACLsWithPaths(ctx, newDirACL, volDirPath3, volDirPath4)
			children, err = client.QueryVolumeChildren(ctx, volumeName)
			assertNoError(t, err)
			assertLen(t, children, 3)
			assertNewDirACL(children, dirPath3Key)
			assertNewDirACL(children, dirPath4Key)
			assertNewFileACL(children, filePath3Key)
		case 5:
			// test a single, root dir with good perms that has a single sub-dir
			// with bad perms, and the sub-dir contains a file with good perms
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath4,
				os.FileMode(newDirMode),
				true,
				true))
			assertNoError(t, apiv2.ContainerCreateFile(
				ctx,
				client.API,
				volDirPath4,
				fileName3,
				0,
				newFileMode,
				&bufReadCloser{&bytes.Buffer{}},
				true))
			setACLsWithPaths(ctx, newFileACL, volFilePath3)
			setACLsWithPaths(ctx, newDirACL, volDirPath3, volDirPath4)
			children, err = client.QueryVolumeChildren(ctx, volumeName)
			assertNoError(t, err)
			assertLen(t, children, 3)
			assertNewDirACL(children, dirPath3Key)
			assertNewDirACL(children, dirPath4Key)
			assertNewFileACL(children, filePath3Key)
		case 6:
			// test /dA/dAA/dAAA where dA has bad perms; the volume delete
			// should fail
			assertNoError(t, client.CreateVolumeDir(
				ctx,
				volumeName,
				dirPath1,
				os.FileMode(newDirMode),
				true,
				true))
		}
	}

	// assert that ForceDeleteVolume works
	createObjs(ctx, 0)
	setACLsWithPaths(ctx, badFileACL, volFilePath0, volFilePath1)
	setACLsWithPaths(ctx,
		badDirACL, volDirPath4, volDirPath3, volDirPath1, volDirPath0)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, volChildCount)
	assertBadDirACL(children, dirPath0Key)
	assertBadDirACL(children, dirPath1Key)
	assertBadDirACL(children, dirPath3Key)
	assertBadDirACL(children, dirPath4Key)
	assertBadFileACL(children, filePath0Key)
	assertBadFileACL(children, filePath1Key)
	// force delete the volume
	assertNoError(t, client.ForceDeleteVolume(ctx, volumeName))

	// assert that an initial delete will result in the removal of files
	// and directories not in conflict
	createObjs(ctx, 0)
	setACLsWithPaths(ctx, badFileACL, volFilePath0, volFilePath1)
	setACLsWithPaths(ctx,
		badDirACL, volDirPath4, volDirPath3, volDirPath1, volDirPath0)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, volChildCount)
	assertBadDirACL(children, dirPath0Key)
	assertBadDirACL(children, dirPath1Key)
	assertBadDirACL(children, dirPath3Key)
	assertBadDirACL(children, dirPath4Key)
	assertBadFileACL(children, filePath0Key)
	assertBadFileACL(children, filePath1Key)
	// attempt to delete the volume; should fail, but the following paths
	// will have been removed:
	//
	// - /dB
	// - /dB/dBB
	// - /dB/dBB/an_empty_file
	// - /an_empty_file
	assertError(t, client.DeleteVolume(ctx, volumeName))
	setACLsWithPaths(
		ctx, newDirACL,
		volDirPath0, volDirPath1, volDirPath3, volDirPath4)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, volChildCount-4)
	assertNewDirACL(children, dirPath0Key)
	assertNewDirACL(children, dirPath1Key)
	assertNewDirACL(children, dirPath3Key)
	assertNewDirACL(children, dirPath4Key)
	// attempt to delete the volume; should succeed
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert that a file with bad perms deep in the hierarchy won't prevent
	// a delete
	createObjs(ctx, 0)
	setACLsWithPaths(ctx, badFileACL, volFilePath1)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, volChildCount)
	assertBadFileACL(children, filePath1Key)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert that a root file with bad perms will not prevent a delete
	createObjs(ctx, 0)
	setACLsWithPaths(ctx, badFileACL, volFilePath0)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, volChildCount)
	assertBadFileACL(children, filePath0Key)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert that a root-directory with an empty sub-dir with bad perms will
	// not prevent a delete
	createObjs(ctx, 1)
	setACLsWithPaths(ctx, badDirACL, volDirPath4)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 2)
	assertBadDirACL(children, dirPath4Key)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert that an empty root-directory with bad perms will not prevent a
	// delete
	createObjs(ctx, 2)
	setACLsWithPaths(ctx, badDirACL, volDirPath3)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 1)
	assertBadDirACL(children, dirPath3Key)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert that a root-directory with bad perms that contains a file with
	// good perms will prevent a delete
	createObjs(ctx, 3)
	setACLsWithPaths(ctx, badDirACL, volDirPath3)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 2)
	assertBadDirACL(children, dirPath3Key)
	assertNewFileACL(children, filePath2Key)
	assertError(t, client.DeleteVolume(ctx, volumeName))
	setACLsWithPaths(ctx, newDirACL, volDirPath3)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert the previous scenario will be handled by a force delete
	createObjs(ctx, 3)
	setACLsWithPaths(ctx, badDirACL, volDirPath3)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 2)
	assertBadDirACL(children, dirPath3Key)
	assertNewFileACL(children, filePath2Key)
	assertNoError(t, client.ForceDeleteVolume(ctx, volumeName))

	// assert that a root-directory with bad perms that contains a sub-dir
	// with good perms that contains a file with good perms prevents a
	// delete
	createObjs(ctx, 4)
	setACLsWithPaths(ctx, badDirACL, volDirPath3)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertBadDirACL(children, dirPath3Key)
	assertNewDirACL(children, dirPath4Key)
	assertNewFileACL(children, filePath3Key)
	assertError(t, client.DeleteVolume(ctx, volumeName))
	setACLsWithPaths(ctx, newDirACL, volDirPath3)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert the previous scenario will be handled by a force delete
	createObjs(ctx, 4)
	setACLsWithPaths(ctx, badDirACL, volDirPath3)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertBadDirACL(children, dirPath3Key)
	assertNewDirACL(children, dirPath4Key)
	assertNewFileACL(children, filePath3Key)
	assertNoError(t, client.ForceDeleteVolume(ctx, volumeName))

	// assert a single, root dir with good perms that has a single sub-dir
	// with bad perms, and the sub-dir contains a file with good perms prevents
	// a delete
	createObjs(ctx, 5)
	setACLsWithPaths(ctx, badDirACL, volDirPath4)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertNewDirACL(children, dirPath3Key)
	assertBadDirACL(children, dirPath4Key)
	assertNewFileACL(children, filePath3Key)
	assertError(t, client.DeleteVolume(ctx, volumeName))
	setACLsWithPaths(ctx, newDirACL, volDirPath4)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert the previous scenario will be handled by a force delete
	createObjs(ctx, 5)
	setACLsWithPaths(ctx, badDirACL, volDirPath4)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertNewDirACL(children, dirPath3Key)
	assertBadDirACL(children, dirPath4Key)
	assertNewFileACL(children, filePath3Key)
	assertNoError(t, client.ForceDeleteVolume(ctx, volumeName))

	// test /dA/dAA/dAAA where dA has bad perms; the volume delete
	// should fail
	createObjs(ctx, 6)
	setACLsWithPaths(ctx, badDirACL, volDirPath0a)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertNewDirACL(children, dirPath0Key)
	assertBadDirACL(children, dirPath0aKey)
	assertNewDirACL(children, dirPath1Key)
	assertError(t, client.DeleteVolume(ctx, volumeName))
	setACLsWithPaths(ctx, newDirACL, volDirPath0a)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertNoError(t, client.DeleteVolume(ctx, volumeName))

	// assert the previous scenario will be handled by a force delete
	createObjs(ctx, 6)
	setACLsWithPaths(ctx, badDirACL, volDirPath0a)
	children, err = client.QueryVolumeChildren(ctx, volumeName)
	assertNoError(t, err)
	assertLen(t, children, 3)
	assertNewDirACL(children, dirPath0Key)
	assertBadDirACL(children, dirPath0aKey)
	assertNewDirACL(children, dirPath1Key)
	assertNoError(t, client.ForceDeleteVolume(ctx, volumeName))
}

type bufReadCloser struct {
	b *bytes.Buffer
}

func (b *bufReadCloser) Read(p []byte) (n int, err error) {
	return b.b.Read(p)
}

func (b *bufReadCloser) Close() error {
	return nil
}
