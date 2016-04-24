package executors

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/api/utils/config"

	// load the executors
	//_ "github.com/emccode/libstorage/drivers/storage/ec2/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/gce/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/isilon/executor"
	_ "github.com/emccode/libstorage/drivers/storage/mock/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/openstack/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/rackspace/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/scaleio/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/vbox/executor"
	_ "github.com/emccode/libstorage/drivers/storage/vfs/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/vmax/executor"
	//_ "github.com/emccode/libstorage/drivers/storage/xtremio/executor"
)

const (
	// InstanceID is the command to execute to get the instance ID.
	InstanceID = "instanceID"

	// LocalDevices is the command to execute to get the local devices map.
	LocalDevices = "localDevices"

	// NextDevice is the command to execute to get the next device.
	NextDevice = "nextDevice"
)

var (
	// LSX is the default  name of the libStorage executor for the current OS.
	LSX string

	cmdRx = regexp.MustCompile("(?i)instanceid|nextdevice|localdevices")
)

func init() {
	if runtime.GOOS == "windows" {
		LSX = "lsx-windows.exe"
	} else {
		LSX = fmt.Sprintf("lsx-%s", runtime.GOOS)
	}
}

// Run runs the executor CLI.
func Run() {

	args := os.Args
	if len(args) != 3 {
		printUsageAndExit()
	}

	d, err := registry.NewStorageExecutor(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	config, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := d.Init(config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cmd := cmdRx.FindString(args[2])
	if cmd == "" {
		printUsageAndExit()
	}
	cmd = strings.ToLower(cmd)

	ctx := context.Background()
	store := utils.NewStore()

	var (
		result interface{}
		op     string
	)

	if cmd == "instanceid" {
		op = "instance ID"
		opResult, opErr := d.InstanceID(ctx, store)
		if opErr != nil {
			err = opErr
		} else {
			result = opResult
		}
	} else if cmd == "nextdevice" {
		op = "next device"
		opResult, opErr := d.NextDevice(ctx, store)
		if opErr != nil {
			err = opErr
		} else {
			result = opResult
		}
	} else if cmd == "localdevices" {
		op = "local devices"
		opResult, opErr := d.LocalDevices(ctx, store)
		if opErr != nil {
			err = opErr
		} else {
			result = opResult
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr,
			"error: error getting %s: %v\n", op, err)
		os.Exit(1)
	}

	switch result.(type) {
	case string:
		fmt.Fprintln(os.Stdout, result)
	default:
		buf, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: error encoding %s: %v\n", op, err)
			os.Exit(1)
		}
		str := string(buf)
		if str == "null" {
			fmt.Fprintf(os.Stdout, "{}\n")
		} else {
			fmt.Fprintf(os.Stdout, "%s\n", str)
		}
	}
}

func printUsage() {
	fmt.Printf(
		"usage: %s executor instanceID|nextDevice|localDevices\n",
		path.Base(os.Args[0]))
}

func printUsageAndExit() {
	printUsage()
	os.Exit(1)
}
