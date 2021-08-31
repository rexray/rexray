package util_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/util"
)

var r10 string
var tmpPrefixDirs []string

func TestMain(m *testing.M) {
	r10 = gotil.RandomString(10)

	exitCode := m.Run()
	for _, d := range tmpPrefixDirs {
		os.RemoveAll(d)
	}
	os.Exit(exitCode)
}

func newTestContext(testName string, t *testing.T) types.Context {
	tmpDir, err := ioutil.TempDir(
		"", fmt.Sprintf("rexray-util_test-%s", testName))
	if err != nil {
		t.Fatal(err)
	}

	pathConfig := utils.NewPathConfig(tmpDir)
	tmpPrefixDirs = append(tmpPrefixDirs, tmpDir)
	return context.WithValue(
		context.Background(),
		context.PathConfigKey, pathConfig)
}

func TestStdOutAndLogFile(t *testing.T) {
	ctx := newTestContext("TestStdOutAndLogFile", t)

	if _, err := util.StdOutAndLogFile(ctx, "BadFile/ (*$"); err == nil {
		t.Fatal("error expected in created BadFile")
	}

	out, err := util.StdOutAndLogFile(ctx, "TestStdOutAndLogFile")

	if err != nil {
		t.Fatal(err)
	}

	if out == nil {
		t.Fatal("out == nil")
	}
}

func TestWriteReadCurrentPidFile(t *testing.T) {
	ctx := newTestContext("TestWriteReadPidFile", t)

	var err error
	var pidRead int

	pid := os.Getpid()

	if err = util.WritePidFile(ctx, -1); err != nil {
		t.Fatalf("error writing pidfile=%s", util.PidFilePath(ctx))
	}

	if pidRead, err = util.ReadPidFile(ctx); err != nil {
		t.Fatalf("error reading pidfile=%s", util.PidFilePath(ctx))
	}

	if pidRead != pid {
		t.Fatalf("pidRead=%d != pid=%d", pidRead, pid)
	}
}

func TestWriteReadCustomPidFile(t *testing.T) {
	ctx := newTestContext("TestWriteReadCustomPidFile", t)

	var err error
	if _, err = util.ReadPidFile(ctx); err == nil {
		t.Fatal("error expected in reading pid file")
	}

	pidWritten := int(time.Now().Unix())
	if err = util.WritePidFile(ctx, pidWritten); err != nil {
		t.Fatalf("error writing pidfile=%s", util.PidFilePath(ctx))
	}

	var pidRead int
	if pidRead, err = util.ReadPidFile(ctx); err != nil {
		t.Fatalf("error reading pidfile=%s", util.PidFilePath(ctx))
	}

	if pidRead != pidWritten {
		t.Fatalf("pidRead=%d != pidWritten=%d", pidRead, pidWritten)
	}
}

func TestReadPidFileWithErrors(t *testing.T) {
	ctx := newTestContext("TestReadPidFileWithErrors", t)

	var err error
	if _, err = util.ReadPidFile(ctx); err == nil {
		t.Fatal("error expected in reading pid file")
	}

	gotil.WriteStringToFile("hello", util.PidFilePath(ctx))

	if _, err = util.ReadPidFile(ctx); err == nil {
		t.Fatal("error expected in reading pid file")
	}
}

func TestInstall(t *testing.T) {
	util.Install()
}

func TestInstallChownRoot(t *testing.T) {
	util.InstallChownRoot()
}

func TestInstallDirChownRoot(t *testing.T) {
	util.InstallDirChownRoot("--help")
}

func TestFindFlagArgs(t *testing.T) {
	{
		val, indices := util.FindFlagVal(
			"-l", "-f", "-l", "debug", "--service=vfs")
		if val != "debug" {
			t.Fatalf("val != debug: %s", val)
		}
		if len(indices) != 2 && indices[0] != 1 && indices[1] != 2 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}

	{
		val, indices := util.FindFlagVal(
			"-l", "-f", "--service=vfs", "-l", "debug")
		if val != "debug" {
			t.Fatalf("val != debug: %s", val)
		}
		if len(indices) != 2 && indices[0] != 2 && indices[1] != 3 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}

	{
		val, indices := util.FindFlagVal(
			"--logLevel", "-f", "--service=vfs", "--loglevel", "debug")
		if val != "debug" {
			t.Fatalf("val != debug: %s", val)
		}
		if len(indices) != 2 && indices[0] != 2 && indices[1] != 3 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}

	{
		val, indices := util.FindFlagVal(
			"--logLevel", "-f", "--service=vfs", "--loglevel=debug")
		if val != "debug" {
			t.Fatalf("val != debug: %s", val)
		}
		if len(indices) != 1 && indices[0] != 2 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}

	{
		val, indices := util.FindFlagVal(
			"-l", "rexray", "-l")
		if val != "" {
			t.Fatalf("val != '': %s", val)
		}
		if len(indices) != 1 && indices[0] != 1 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}

	{
		val, indices := util.FindFlagVal(
			"--logLevel", "rexray", "--logLevel")
		if val != "" {
			t.Fatalf("val != '': %s", val)
		}
		if len(indices) != 1 && indices[0] != 1 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}

	{
		val, indices := util.FindFlagVal(
			"--logLevel", "rexray", "--logLevel=")
		if val != "" {
			t.Fatalf("val != '': %s", val)
		}
		if len(indices) != 1 && indices[0] != 1 {
			t.Fatalf("invalid indices: %v", indices)
		}
	}
}
