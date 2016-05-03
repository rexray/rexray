package executors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/utils"
	apiconfig "github.com/emccode/libstorage/api/utils/config"

	_ "github.com/emccode/libstorage/imports/config"
	_ "github.com/emccode/libstorage/imports/executors"
)

const (
	// InstanceID is the command to execute to get the instance ID.
	InstanceID = "instanceID"

	// LocalDevices is the command to execute to get the local devices map.
	LocalDevices = "localDevices"

	// NextDevice is the command to execute to get the next device.
	NextDevice = "nextDevice"

	// WaitForDevice is the command to execute to wait until a device,
	// identified by volume ID, is presented to the system.
	WaitForDevice = "wait"
)

var (
	cmdRx = regexp.MustCompile(
		`(?i)^instanceid|nextdevice|localdevices|wait$`)
)

// Run runs the executor CLI.
func Run() {

	args := os.Args
	if !(len(args) == 3 || len(args) == 5) {
		printUsageAndExit()
	}

	d, err := registry.NewStorageExecutor(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	config, err := apiconfig.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	apiconfig.UpdateLogLevel(config)
	ctx := context.Background()

	if err := d.Init(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cmd := cmdRx.FindString(args[2])
	if cmd == "" {
		printUsageAndExit()
	}
	cmd = strings.ToLower(cmd)

	store := utils.NewStore()

	var (
		result   interface{}
		op       string
		exitCode int
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
	} else if cmd == "wait" {
		if len(args) != 5 {
			printUsageAndExit()
			os.Exit(1)
		}
		op = "wait"
		volID := args[3]
		timeout, parseErr := time.ParseDuration(args[4])
		if parseErr != nil {
			err = parseErr
		} else {

			ldl := func() (bool, map[string]string, error) {
				ldm, err := d.LocalDevices(ctx, store)
				if err != nil {
					return false, nil, err
				}
				for k, _ := range ldm {
					if strings.ToLower(k) == strings.ToLower(volID) {
						return true, ldm, nil
					}
				}
				return false, ldm, nil
			}

			var (
				found    bool
				opErr    error
				opResult map[string]string
				timeoutC = time.After(timeout)
				tick     = time.Tick(500 * time.Millisecond)
			)

		TimeoutLoop:

			for {
				select {
				case <-timeoutC:
					exitCode = 255
					break TimeoutLoop
				case <-tick:
					if found, opResult, opErr = ldl(); found || opErr != nil {
						break TimeoutLoop
					}
				}
			}

			if opErr != nil {
				err = opErr
			} else {
				result = opResult
			}
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

	os.Exit(exitCode)
}

func printUsage() {
	buf := &bytes.Buffer{}
	waitCmd := "wait {volumeID} {timeout}"
	fmt.Fprintf(buf, "usage: %s {executor} ", os.Args[0])
	lpad := fmt.Sprintf("%%%ds\n", buf.Len()+len(waitCmd))
	fmt.Fprintf(buf, "instanceID|nextDevice|localDevices\n")
	fmt.Fprintf(buf, lpad, waitCmd)
	fmt.Fprint(os.Stderr, buf.String())
}

func printUsageAndExit() {
	printUsage()
	os.Exit(1)
}
