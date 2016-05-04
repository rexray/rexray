package tests

import (
	"io/ioutil"
	"testing"

	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api/types"
)

// TestExecutors tests the GET /executors route.
var TestExecutors = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().Executors(nil)
	if err != nil {
		t.Fatal(err)
	}
	//assertLSXWindows(t, reply["lsx-windows.exe"])
	assertLSXLinux(t, reply["lsx-linux"])
	assertLSXDarwin(t, reply["lsx-darwin"])
}

// TestHeadExecutorWindows tests the HEAD /executors/lsx-windows.exe route.
/*var TestHeadExecutorWindows = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().ExecutorHead(nil, "lsx-windows.exe")
	if err != nil {
		t.Fatal(err)
	}
	assertLSXWindows(t, reply)
}*/

// TestHeadExecutorLinux tests the HEAD /executors/lsx-linux route.
var TestHeadExecutorLinux = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().ExecutorHead(nil, "lsx-linux")
	if err != nil {
		t.Fatal(err)
	}
	assertLSXLinux(t, reply)
}

// TestHeadExecutorDarwin tests the HEAD /executors/lsx-darwin route.
var TestHeadExecutorDarwin = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().ExecutorHead(nil, "lsx-darwin")
	if err != nil {
		t.Fatal(err)
	}
	assertLSXDarwin(t, reply)
}

// TestGetExecutorWindows tests the GET /executors/lsx-windows.exe route.
/*var TestGetExecutorWindows = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().ExecutorGet(nil, "lsx-windows.exe")
	if err != nil {
		t.Fatal(err)
	}
	defer reply.Close()
	buf, err := ioutil.ReadAll(reply)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, lsxWindowsInfo.Size, len(buf))
}*/

// TestGetExecutorLinux tests the GET /executors/lsx-linux route.
var TestGetExecutorLinux = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().ExecutorGet(nil, "lsx-linux")
	if err != nil {
		t.Fatal(err)
	}
	defer reply.Close()
	buf, err := ioutil.ReadAll(reply)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, lsxLinuxInfo.Size, len(buf))
}

// TestGetExecutorDarwin tests the GET /executors/lsx-darwin route.
var TestGetExecutorDarwin = func(
	config gofig.Config,
	client types.Client, t *testing.T) {

	reply, err := client.API().ExecutorGet(nil, "lsx-darwin")
	if err != nil {
		t.Fatal(err)
	}
	defer reply.Close()
	buf, err := ioutil.ReadAll(reply)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, lsxDarwinInfo.Size, len(buf))
}

/*func assertLSXWindows(t *testing.T, i *types.ExecutorInfo) {
	assert.Equal(t, lsxWindowsInfo.Name, i.Name)
	assert.EqualValues(t, lsxWindowsInfo.Size, i.Size)
	assert.Equal(t, lsxWindowsInfo.MD5Checksum, i.MD5Checksum)
}*/

func assertLSXLinux(t *testing.T, i *types.ExecutorInfo) {
	assert.Equal(t, lsxLinuxInfo.Name, i.Name)
	assert.EqualValues(t, lsxLinuxInfo.Size, i.Size)
	assert.Equal(t, lsxLinuxInfo.MD5Checksum, i.MD5Checksum)
}

func assertLSXDarwin(t *testing.T, i *types.ExecutorInfo) {
	assert.Equal(t, lsxDarwinInfo.Name, i.Name)
	assert.EqualValues(t, lsxDarwinInfo.Size, i.Size)
	assert.Equal(t, lsxDarwinInfo.MD5Checksum, i.MD5Checksum)
}
