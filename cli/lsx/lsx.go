package lsx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/emccode/libstorage/api/context"
	"github.com/emccode/libstorage/api/registry"
	apitypes "github.com/emccode/libstorage/api/types"
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
	if len(args) < 3 {
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
		if len(args) < 4 {
			printUsageAndExit()
		}
		op = "local devices"
		opResult, opErr := d.LocalDevices(ctx, &apitypes.LocalDevicesOpts{
			ScanType: apitypes.ParseDeviceScanType(args[3]),
			Opts:     store,
		})
		if opErr != nil {
			err = opErr
		} else {
			result = opResult
		}
	} else if cmd == "wait" {
		if len(args) < 5 {
			printUsageAndExit()
		}
		op = "wait"
		opts := &apitypes.WaitForDeviceOpts{
			LocalDevicesOpts: apitypes.LocalDevicesOpts{
				ScanType: apitypes.ParseDeviceScanType(args[3]),
				Opts:     store,
			},
			Token:   strings.ToLower(args[4]),
			Timeout: utils.DeviceAttachTimeout(args[5]),
		}

		ldl := func() (bool, map[string]string, error) {
			ldm, err := d.LocalDevices(ctx, &opts.LocalDevicesOpts)
			if err != nil {
				return false, nil, err
			}
			for k := range ldm {
				if strings.ToLower(k) == opts.Token {
					return true, ldm, nil
				}
			}
			return false, ldm, nil
		}

		var (
			found    bool
			opErr    error
			opResult map[string]string
			timeoutC = time.After(opts.Timeout)
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

func executorNames() <-chan string {
	c := make(chan string)
	go func() {
		for se := range registry.StorageExecutors() {
			c <- strings.ToLower(se.Name())
		}
		close(c)
	}()
	return c
}

func printUsage() {
	buf := &bytes.Buffer{}
	w := io.MultiWriter(buf, os.Stderr)

	fmt.Fprintf(w, "usage: ")
	lpad1 := buf.Len()
	fmt.Fprintf(w, "%s <executor> ", os.Args[0])
	lpad2 := buf.Len()
	fmt.Fprintf(w, "instanceID\n")
	printUsageLeftPadded(w, lpad2, "nextDevice\n")

	printUsageLeftPadded(w, lpad2, "localDevices <scanType>\n")
	printUsageLeftPadded(w, lpad2, "wait <scanType> <attachToken> <timeout>\n")
	fmt.Fprintln(w)
	executorVar := "executor:    "
	printUsageLeftPadded(w, lpad1, executorVar)
	lpad3 := lpad1 + len(executorVar)

	execNames := []string{}
	for en := range executorNames() {
		execNames = append(execNames, en)
	}

	if len(execNames) > 0 {
		execNames = utils.SortByString(execNames)
		fmt.Fprintf(w, "%s\n", execNames[0])
		if len(execNames) > 1 {
			for x, en := range execNames {
				if x == 0 {
					continue
				}
				printUsageLeftPadded(w, lpad3, "%s\n", en)
			}
		}
		fmt.Fprintln(w)
	}

	printUsageLeftPadded(w, lpad1, "scanType:    0,quick | 1,deep\n\n")
	printUsageLeftPadded(w, lpad1, "attachToken: <token>\n\n")
	printUsageLeftPadded(w, lpad1, "timeout:     30s | 1h | 5m\n\n")
}

func printUsageLeftPadded(
	w io.Writer, lpadLen int, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	lpadFmt := fmt.Sprintf("%%%ds", lpadLen+len(text))
	fmt.Fprintf(w, lpadFmt, text)
}

func printUsageAndExit() {
	printUsage()
	os.Exit(1)
}
