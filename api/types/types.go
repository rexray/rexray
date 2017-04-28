package types

import (
	"io"
	"os"
	"strconv"
)

var (
	// Debug is a flag that indicates whether or not the environment variable
	// `LIBSTORAGE_DEBUG` is set to a boolean true value.
	Debug, _ = strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG"))

	// Stdout is the writer used when a component in libStorage wishes to
	// write to the standard output stream.
	Stdout io.Writer = os.Stdout

	// Stderr is the writer used when a component in libStorage wishes to
	// write to the standard error stream.
	Stderr io.Writer = os.Stderr
)
