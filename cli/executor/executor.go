package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types/context"
	"github.com/emccode/libstorage/api/utils"
)

var (
	cmdRx = regexp.MustCompile("(?i)instanceid|nextdevice|localdevices")
)

// Run runs the executor CLI.
func Run() {

	args := os.Args
	if len(args) != 2 {
		printUsageAndExit()
	}

	cmd := cmdRx.FindString(args[1])
	if cmd == "" {
		printUsageAndExit()
	}
	cmd = strings.ToLower(cmd)

	d := <-registry.StorageExecutors()
	if d == nil {
		fmt.Fprintln(os.Stderr, "error: no storage driver available")
		os.Exit(1)
	}

	ctx := context.Background()
	store := utils.NewStore()

	var (
		err    error
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
		"usage: %s instanceID|nextDevice|localDevices\n",
		path.Base(os.Args[0]))
}

func printUsageAndExit() {
	printUsage()
	os.Exit(1)
}
