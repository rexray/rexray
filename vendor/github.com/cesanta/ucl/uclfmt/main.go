// uclfmt is a utility for formatting UCL files.
// It looks for .ucl-format file in any of the parent directories of the input file
// (or current working dir, if reading from stdin), parses it as JSON into ucl.FormatConfig
// structure and uses that to re-format the input file.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cesanta/ucl"
	"github.com/golang/glog"
)

var (
	input      = flag.String("f", "", "Input file name. If empty, will read from stdin.")
	configPath = flag.String("c", "", "Formatting config file. If empty, will look for '.ucl-format' parent directories of the input file (or current working directory, if reading from stdin).")
	overwrite  = flag.Bool("w", false, "If true, re-format file in place.")
)

func parseConfig(filename string) (*ucl.FormatConfig, error) {
	glog.Infof("Parsing config file %q", filename)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := &ucl.FormatConfig{}
	if err := json.NewDecoder(f).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func getConfig(start string) (*ucl.FormatConfig, error) {
	if *configPath != "" {
		return parseConfig(*configPath)
	}
	for {
		f := filepath.Join(start, ".ucl-format")
		glog.Infof("Trying %q...", f)
		fi, err := os.Stat(f)
		if err == nil && !fi.IsDir() {
			return parseConfig(f)
		}

		if len(start) == 1 {
			break
		}
		start = filepath.Dir(start)
	}
	return nil, nil // means empty config
}

func exit(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
	os.Exit(1)
}

func main() {
	flag.Parse()

	go func() { time.Sleep(5 * time.Second); panic("wat") }()

	if *input == "" && *overwrite {
		exit("Cannot overwrite stdin.")
	}

	var err error
	in := os.Stdin
	if *input != "" {
		in, err = os.Open(*input)
		if err != nil {
			exit("Failed to open input file: %s", err)
		}
		defer in.Close()
		glog.Infof("Will read from %q", *input)
	} else {
		glog.Infof("Will read from stdin.")
	}

	glog.Infof("Parsing...")
	v, err := ucl.Parse(in)
	if err != nil {
		exit("Failed to parse input: %s", err)
	}
	glog.Infof("Done.")

	out := os.Stdout
	if *overwrite {
		out, err = ioutil.TempFile(os.TempDir(), "ucl-format")
		if err != nil {
			exit("Cannot create temporary file: %s", err)
		}
		defer os.Remove(out.Name())
		glog.Infof("Will write to %q", out.Name())
	} else {
		glog.Infof("Will write to stdout.")
	}

	var config *ucl.FormatConfig
	if *input != "" {
		f, err := filepath.Abs(*input)
		if err != nil {
			exit("Failed to get absolute path to input file: %s", err)
		}
		glog.Infof("Absolute path to the input file: %s", f)
		config, err = getConfig(filepath.Dir(f))
		if err != nil {
			exit("Failed to read config: %s", err)
		}
	} else {
		p, err := os.Getwd()
		if err != nil {
			exit("Failed to get current working directory: %s", err)
		}
		config, err = getConfig(p)
		if err != nil {
			exit("Failed to read config: %s", err)
		}
	}
	glog.Infof("Formatting the data...")
	err = ucl.Format(v, config, out)
	if err != nil {
		exit("Failed to format input: %s", err)
	}
	fmt.Fprintf(out, "\n")
	glog.Infof("Done.")

	if *overwrite {
		f := out.Name()
		out.Close()
		glog.Infof("Moving temporary file in place...")
		if err := os.Rename(f, *input); err != nil {
			exit("Failed to rename %q to %q: %s", f, *input, err)
		}
		glog.Infof("Done.")
	}
}
