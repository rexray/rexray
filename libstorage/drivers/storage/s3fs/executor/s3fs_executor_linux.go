// +build linux

package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

func getMountedBuckets(
	ctx types.Context,
	s3fsBinName string) (map[string]string, error) {

	s3fsBinRX := regexp.MustCompile(fmt.Sprintf(`^.*%s$`, s3fsBinName))

	infc, errc, err := walkProc(ctx)
	if err != nil {
		return nil, err
	}

	var (
		wg            sync.WaitGroup
		argc          = make(chan []string)
		numInspectors = runtime.NumCPU() + 1
	)
	wg.Add(numInspectors)
	for i := 0; i < numInspectors; i++ {
		go func() {
			inspect(ctx, infc, argc)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(argc)
	}()

	m := map[string]string{}
	for args := range argc {
		if len(args) < 3 {
			continue
		}
		if !s3fsBinRX.MatchString(args[0]) {
			continue
		}
		m[args[1]] = args[2]
	}
	if err := <-errc; err != nil {
		return nil, err
	}
	return m, nil
}

func walkProc(ctx types.Context) (<-chan os.FileInfo, <-chan error, error) {

	proc, err := os.Open("/proc")
	if err != nil {
		return nil, nil, err
	}
	defer proc.Close()
	infos, err := proc.Readdir(-1)
	if err != nil {
		return nil, nil, err
	}
	var (
		infc = make(chan os.FileInfo)
		errc = make(chan error, 1)
	)
	go func() {
		defer close(infc)
		errc <- func() error {
			for _, i := range infos {
				select {
				case infc <- i:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}()
	}()
	return infc, errc, nil
}

var pidDirRX = regexp.MustCompile(`^\d+$`)

func inspect(
	ctx types.Context,
	infc <-chan os.FileInfo,
	c chan<- []string) {

	for i := range infc {
		if !i.IsDir() {
			continue
		}
		if !pidDirRX.MatchString(i.Name()) {
			continue
		}
		func() {
			cmdLineFile, err := os.Open(path.Join("/proc", i.Name(), "cmdline"))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			defer cmdLineFile.Close()
			buf, err := ioutil.ReadAll(cmdLineFile)
			if err != nil {
				return
			}
			if len(buf) == 0 {
				return
			}
			select {
			case c <- strings.Split(string(buf), "\x00"):
			case <-ctx.Done():
				return
			}
		}()
	}
}
