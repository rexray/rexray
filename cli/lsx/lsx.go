package lsx

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/akutz/goof"

	"github.com/codedellemc/libstorage/api/context"
	"github.com/codedellemc/libstorage/api/registry"
	apitypes "github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
	apiconfig "github.com/codedellemc/libstorage/api/utils/config"

	// load these packages
	_ "github.com/codedellemc/libstorage/imports/config"
	_ "github.com/codedellemc/libstorage/imports/executors"
)

var cmdRx = regexp.MustCompile(
	`(?i)^((?:un?)?mounts?|supported|instanceid|nextdevice|localdevices|wait)$`)

// Run runs the executor CLI.
func Run() {

	ctx := context.Background()
	ctx = ctx.WithValue(context.PathConfigKey, utils.NewPathConfig(ctx, "", ""))
	registry.ProcessRegisteredConfigs(ctx)

	args := os.Args
	if len(args) < 3 {
		printUsageAndExit()
	}

	d, err := registry.NewStorageExecutor(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	driverName := strings.ToLower(d.Name())

	config, err := apiconfig.NewConfig(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	apiconfig.UpdateLogLevel(config)

	if err := d.Init(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cmd := cmdRx.FindString(args[2])
	if cmd == "" {
		printUsageAndExit()
	}
	store := utils.NewStore()

	var (
		result   interface{}
		op       string
		exitCode int
	)

	if strings.EqualFold(cmd, apitypes.LSXCmdSupported) {
		op = apitypes.LSXCmdSupported
		if dws, ok := d.(apitypes.StorageExecutorWithSupported); ok {
			opResult, opErr := dws.Supported(ctx, store)
			if opErr != nil {
				err = opErr
			} else if opResult {
				rflags := apitypes.LSXOpAllNoMount
				if _, ok := dws.(apitypes.StorageExecutorWithMount); ok {
					rflags = rflags | apitypes.LSXSOpMount
				}
				if _, ok := dws.(apitypes.StorageExecutorWithUnmount); ok {
					rflags = rflags | apitypes.LSXSOpUmount
				}
				if _, ok := dws.(apitypes.StorageExecutorWithMounts); ok {
					rflags = rflags | apitypes.LSXSOpMounts
				}
				result = rflags
			} else {
				result = apitypes.LSXSOpNone
			}
		} else {
			err = apitypes.ErrNotImplemented
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdMounts) {
		op = apitypes.LSXCmdMounts
		dd, ok := d.(apitypes.StorageExecutorWithMounts)
		if !ok {
			err = apitypes.ErrNotImplemented
		} else {
			mounts, opErr := dd.Mounts(ctx, store)
			if opErr != nil {
				err = opErr
			} else if mounts == nil {
				result = []*apitypes.MountInfo{}
			} else {
				result = mounts
			}
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdMount) {
		op = apitypes.LSXCmdMount
		dd, ok := d.(apitypes.StorageExecutorWithMount)
		if !ok {
			err = apitypes.ErrNotImplemented
		} else {
			var (
				deviceName string
				mountPath  string
				mountOpts  = &apitypes.DeviceMountOpts{Opts: store}
			)
			mountArgs := args[3:]
			if len(mountArgs) == 0 {
				printUsageAndExit()
			}

			remArgs := []string{}
			for x := 0; x < len(mountArgs); {
				a := mountArgs[x]
				if x < len(mountArgs)-1 {
					switch a {
					case "-l":
						mountOpts.MountLabel = mountArgs[x+1]
						x = x + 2
						continue
					case "-o":
						mountOpts.MountOptions = mountArgs[x+1]
						x = x + 2
						continue
					}
				}
				remArgs = append(remArgs, a)
				x++
			}

			if len(remArgs) != 2 {
				printUsageAndExit()
			}

			deviceName = remArgs[0]
			mountPath = remArgs[1]

			opErr := dd.Mount(ctx, deviceName, mountPath, mountOpts)
			if opErr != nil {
				err = opErr
			} else {
				result = mountPath
			}
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdUmount) ||
		strings.EqualFold(cmd, "unmount") {
		op = apitypes.LSXCmdUmount
		dd, ok := d.(apitypes.StorageExecutorWithUnmount)
		if !ok {
			err = apitypes.ErrNotImplemented
		} else {
			if len(args) < 4 {
				printUsageAndExit()
			}
			mountPath := args[3]
			opErr := dd.Unmount(ctx, mountPath, store)
			if opErr != nil {
				err = opErr
			} else {
				result = mountPath
			}
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdInstanceID) {
		op = apitypes.LSXCmdInstanceID
		opResult, opErr := d.InstanceID(ctx, store)
		if opErr != nil {
			err = opErr
		} else {
			opResult.Driver = driverName
			result = opResult
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdNextDevice) {
		op = apitypes.LSXCmdNextDevice
		opResult, opErr := d.NextDevice(ctx, store)
		if opErr != nil && opErr != apitypes.ErrNotImplemented {
			err = opErr
		} else {
			result = opResult
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdLocalDevices) {
		if len(args) < 4 {
			printUsageAndExit()
		}
		op = apitypes.LSXCmdLocalDevices
		opResult, opErr := d.LocalDevices(ctx, &apitypes.LocalDevicesOpts{
			ScanType: apitypes.ParseDeviceScanType(args[3]),
			Opts:     store,
		})
		if opErr != nil {
			err = opErr
		} else {
			opResult.Driver = driverName
			result = opResult
		}
	} else if strings.EqualFold(cmd, apitypes.LSXCmdWaitForDevice) {
		if len(args) < 6 {
			printUsageAndExit()
		}
		op = apitypes.LSXCmdWaitForDevice
		opts := &apitypes.WaitForDeviceOpts{
			LocalDevicesOpts: apitypes.LocalDevicesOpts{
				ScanType: apitypes.ParseDeviceScanType(args[3]),
				Opts:     store,
			},
			Token:   strings.ToLower(args[4]),
			Timeout: utils.DeviceAttachTimeout(args[5]),
		}

		ldl := func() (bool, *apitypes.LocalDevices, error) {
			ldm, err := d.LocalDevices(ctx, &opts.LocalDevicesOpts)
			if err != nil {
				return false, nil, err
			}
			for k := range ldm.DeviceMap {
				if strings.ToLower(k) == opts.Token {
					return true, ldm, nil
				}
			}
			return false, ldm, nil
		}

		var (
			found    bool
			opErr    error
			opResult *apitypes.LocalDevices
			timeoutC = time.After(opts.Timeout)
			tick     = time.Tick(500 * time.Millisecond)
		)

	TimeoutLoop:

		for {
			select {
			case <-timeoutC:
				exitCode = apitypes.LSXExitCodeTimedOut
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
			opResult.Driver = driverName
			result = opResult
		}
	}

	if err != nil {
		// if the function is not implemented then exit with
		// apitypes.LSXExitCodeNotImplemented to let callers
		// know that the function is unsupported on this system
		exitCode = 1
		if strings.EqualFold(err.Error(), apitypes.ErrNotImplemented.Error()) {
			exitCode = apitypes.LSXExitCodeNotImplemented
		}
		var errStr string
		switch e := err.(type) {
		case goof.Goof:
			e.IncludeFieldsInError(true)
			errStr = e.Error()
		default:
			errStr = e.Error()
		}
		fmt.Fprintf(os.Stderr,
			"error: error getting %s: %v\n", op, errStr)
		os.Exit(exitCode)
	}

	switch tr := result.(type) {
	case bool:
		fmt.Fprintf(os.Stdout, "%v", result)
	case string:
		fmt.Fprintln(os.Stdout, result)
	case encoding.TextMarshaler:
		buf, err := tr.MarshalText()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: error encoding %s: %v\n", op, err)
			os.Exit(1)
		}
		os.Stdout.Write(buf)
	default:
		buf, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: error encoding %s: %v\n", op, err)
			os.Exit(1)
		}
		if isNullBuf(buf) {
			os.Stdout.Write(emptyJSONBuff)
		} else {
			os.Stdout.Write(buf)
		}
	}

	os.Exit(exitCode)
}

const (
	newline = 10
)

var (
	nullBuff      = []byte{110, 117, 108, 108}
	emptyJSONBuff = []byte{123, 125}
)

func isNullBuf(buf []byte) bool {
	return len(buf) == len(nullBuff) &&
		buf[0] == nullBuff[0] && buf[1] == nullBuff[1] &&
		buf[2] == nullBuff[2] && buf[3] == nullBuff[3]
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
	fmt.Fprintf(w, "supported\n")
	printUsageLeftPadded(w, lpad2, "instanceID\n")
	printUsageLeftPadded(w, lpad2, "nextDevice\n")
	printUsageLeftPadded(w, lpad2, "localDevices <scanType>\n")
	printUsageLeftPadded(w, lpad2, "wait <scanType> <attachToken> <timeout>\n")
	printUsageLeftPadded(w, lpad2, "mounts\n")
	printUsageLeftPadded(w, lpad2, "mount [-l label] [-o options] device path\n")
	printUsageLeftPadded(w, lpad2, "umount path\n")
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
