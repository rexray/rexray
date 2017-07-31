// validate-json is a utility for validation JSON values with JSON Schema.
//
// Usage example:
//
//   validate-json --schema path/to/schema.json --input path/to/data.json
//
// If everything is fine it will exit with status 0 and without any output. If
// there was any errors, exit code will be non-zero and errors will be printed
// to stderr.
//
// Additional flags:
//
//   --extra "schema1.json schema2.json ..."
// Space-separated list of additional schema files to load so they can be
// referenced from the primary schema. Each of the schemas in these files needs
// to have "id" set.
//
//   -n
// If present, referenced schemas will be fetched from the remote hosts.
//
//   -nodraft04schema
// If present, copy of http://json-schema.org/draft-04/schema embedded in the
// binary will not be pre-loaded.
package main

// go get github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -nocompress -ignore=.*\.go$ -prefix=schema/ schema/

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	json "github.com/cesanta/ucl"
	"github.com/cesanta/validate-json/schema"
)

var (
	schemaFile        = flag.String("schema", "", "Path to schema to use.")
	inputFile         = flag.String("input", "", "Path to the JSON data to validate.")
	network           = flag.Bool("n", false, "If true, fetching of referred schemas from remote hosts will be enabled.")
	extra             = flag.String("extra", "", "Space-separated list of schema files to pre-load for the purpose of remote references. Each schema needs to have 'id' property.")
	skipDefaultSchema = flag.Bool("nodraft04schema", false, "If set to true, http://json-schema.org/draft-04/schema will not be pre-loaded.")
)

func main() {
	flag.Parse()

	if *schemaFile == "" || *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Need --schema and --input\n")
		os.Exit(1)
	}

	f, err := os.Open(*schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %q: %s\n", *schemaFile, err)
		os.Exit(1)
	}

	s, err := json.Parse(f)
	f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read schema: %s\n", err)
		os.Exit(1)
	}

	loader := schema.NewLoader()
	loader.EnableNetworkAccess(*network)
	if *extra != "" {
		for _, file := range strings.Split(*extra, " ") {
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open %q: %s\n", file, err)
				os.Exit(1)
			}
			s, err := json.Parse(f)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse %q: %s\n", file, err)
				os.Exit(1)
			}
			loader.Add(s)
		}
	}
	if !*skipDefaultSchema {
		ds, err := json.Parse(bytes.NewBuffer(MustAsset("draft04schema.json")))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse embedded draft 04 schema: %s\n", err)
			os.Exit(1)
		}
		// Just to be sure, schema.ParseDraft04Schema exercises different code path.
		v, err := schema.NewValidator(ds, loader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create validator for draft04 schema, please file a bug: %s\n", err)
		} else {
			if err := v.Validate(s); err != nil {
				fmt.Fprintln(os.Stderr, "If you see this message, please file a bug and attach the schema you're using.")
				fmt.Fprintf(os.Stderr, "Warning: failed to validate %q with draft 04 schema: %s\n", *schemaFile, err)
			}
		}
	}

	validator, err := schema.NewValidator(s, loader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create validator: %s\n", err)
		os.Exit(1)
	}

	f, err = os.Open(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open input file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()
	data, err := json.Parse(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse input file: %s\n", err)
		os.Exit(1)
	}
	if err := validator.Validate(data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
