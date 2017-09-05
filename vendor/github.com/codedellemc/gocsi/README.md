# GoCSI
The Container Storage Interface
([CSI](https://github.com/container-storage-interface/spec))
is an industry standard specification for creating storage plug-ins
for container orchestrators. GoCSI aids in the development and testing
of CSI plug-ins and provides the following:

| Component | Description |
|-----------|-------------|
| [GoCSI](#library) | CSI Go library |
| [csc](#client) | CSI command line interface (CLI) client |
| [mock](#mock-plug-in) | CSI mock plug-in |

## Library
The root of the GoCSI project is a general purpose library for CSI
that contains package-level functions for invoking the CSI Controller,
Identity, and Node RPCs in addition to the following gRPC client and
server interceptors:

| Type | Name | Description |
|------|------|-------------|
| Unary, client interceptor | `ClientCheckResponseError` | Parses CSI errors into Go errors
|                           | `ClientResponseValidator` | Validates responses |
| Unary, server interceptor | `RequestIDInjector` | Injects a unique ID into a gRPC request |
|                           | `ServerRequestLogger` | Logs requests |
|                           | `ServerResponseLogger` | Logs responses |
|                           | `ServerRequestVersionValidator` | Validates request versions |
|                           | `ServerRequestValidator` | Validates requests |

Examples illustrating the above interceptors and invoking the CSI RPCs
may be found in the GoCSI test suite, the CSI client, and the CSI mock
plug-in.

## Client
The CSI client `csc` is useful when developing and testing CSI plug-ins
because `csc` can easily and directly invoke any of CSI's RPCs directly
from the command line.

Please see the `csc` package for
[additional documentation](./csc/README.md).

## Mock Plug-in
The mock plug-in is a stand-alone binary that implements the CSI
Controller, Identity, and Node RPCs in addition to the specification's
requirements regarding idempotency.

Please see the `mock` package for
[additional documentation](./mock/README.md).

## CSI Specification Version
GoCSI references the
[CSI spec](https://github.com/container-storage-interface/spec)
project in order to obtain the CSI specification.

Please see the `csi` package for
[additional documentation](./csi/README.md).

## Build Reference
GoCSI is _go gettable_ - that means it is possible to build GoCSI with
the following command:

```bash
$ go get github.com/codedellemc/gocsi
```

If GoCSI has already been cloned locally via Git or the aforementioned
`go get` command then there are two ways to build GoCSI:

1. [Build with Go](#building-with-go)
2. [Build with Make](#building-with-make)

### Building with Go
The following commands will build GoCSI's client and mock plug-in:

```bash
$ go build -o csc/csc ./csc && go build -o mock/mock ./mock
```

The above command produces CSI client and mock plug-in binaries at
`csc/csc` and `mock/mock`.

### Building with Make
Make can be used to build all of GoCSI's components with a single command:

```bash
$ make
```

The above command will verify that the generated protobuf and Go
language bindings are up-to-date and that the GoCSI library, client,
and mock plug-in all build successfully.

## Testing
The GoCSI test suite makes use of the Mock plug-in in order to provide
a CSI endpoint that hosts the Controller, Identity, and Node services.
The following command can be used to execute the test suite:

```bash
$ go test
```

The above command is the simplest way to run the GoCSI tests. However,
there are more advanced test scenarios:

| Test Scenario | Description |
|---------------|-------------|
| [Ginkgo](#ginkgo) | Using the Ginkgo test runner |
| [`CSI_ENDPOINT`](#csi_endpoint) | Using an external server endpoint |
| [`GOCSI_MOCK`](#gocsi_mock) | Specifying the path to the server binary |
| [`GOCSI_TEST_DEBUG`](#gocsi_test_debug) | Showing the server process output |

### Ginkgo
The GoCSI tests are written using the
[Ginkgo](http://onsi.github.io/ginkgo/) natural language testing
domain-specific-language (DSL). To take full advantage of GoCSI's test
capabilities the Ginkgo and Gomega dependencies are required and are
included in the GoCSI's `vendor` directory.

Either of the following two commands may be used to build the Ginkgo
test runner:

1. `$ go build ./vendor/github.com/onsi/ginkgo/ginkgo`
2. `$ make ginkgo`

Both of the above commands will place the `ginkgo` binary in the GoCSI
project's root directory.

To execute the GoCSI test suite with Ginkgo please use one of the two
commands below:

1. `$ ./ginkgo`
2. `$ make test`

The first command produces output nearly identical to the output of
`go test`. Use `./ginkgo -?` to print a list of all the flags and
options available to the Ginkgo test runner.

The second command simply ensures that GoCSI's generated sources and
client and mock binaries are up-to-date before executing `./ginkgo -v`.
The `-v` flag executes the tests in verbose mode, printing the names
of each of the test cases as they're executed.

### `CSI_ENDPOINT`
When the GoCSI test suite is executed the first thing that occurs is
checking for the value of the environment variable `CSI_ENDPOINT`.
If `CSI_ENDPOINT` is set then the test suite will **not** create a new
server for every test case using the mock plug-in. Instead all test
cases are executed against the server specified by `CSI_ENDPOINT`.

Please note that using an external CSI server with the test suite will
likely result in failure as the test cases expect a new Mock server
instance at the start of each test case. Still, this is a helpful
feature when wanting to run the Mock plug-in separately in order to
watch its output when executing a specific test case against it.

### `GOCSI_MOCK`
If the `CSI_ENDPOINT` environment variable is not set the environment
variable `GOCSI_MOCK` is checked. This variable's value should be the
fully-qualified to a binary that starts a CSI plug-in. The binary is
started with a clean environment except for a single environment
variable, `CSI_ENDPOINT`, which points to a temporary file that the
server should use as the UNIX socket for serving the CSI gRPC services.

If the `GOCSI_MOCK` environment variable is **not** set then the GoCSI
test suite will automatically build the GoCSI Mock plug-in binary
before executing any test cases:

```bash
$ go build -o mock github.com/codedellemc/gocsi/mock
```

The above command builds the Mock plug-in binary in the working,
temporary directory of the test process. The GoCSI test suite then uses
this binary to launch a new CSI server for each test case. This binary
is removed when the test execution completes and the temporary test
directory and its contents are removed.

Because the Mock plug-in binary is built automatically if `GOCSI_MOCK`
is not set, it means the following two commands are in fact identical:

```bash
# Build the Mock plug-in binary and then launch the test suite while
# specifying the path to the binary with GOCSI_MOCK
$ go build -o mock/mock ./mock && GOCSI_MOCK=$(pwd)/mock/mock go test

# Launch the GoCSI test suite without specifying GOCSI_MOCK. This causes
# the test process to automatically build the Mock plug-in binary and
# use it for the test run
$ go test
```

### `GOCSI_TEST_DEBUG`
Setting the environment variable `GOCSI_TEST_DEBUG` to `true` causes
the GoCSI test suite to read the `STDOUT` and `STDERR` pipes of the
server process launched with each test case and copy the contents of
the streams to the test process's own `STDOUT` and `STDERR` pipes.

This will clutter the test output with the output of the CSI server
processes, but it is useful when debugging.

Still, a cleaner solution for viewing the server output might be to
launch the Mock plug-in in a stand-alone process and then run the test
suite while setting [`CSI_ENDPOINT`](#csi_endpoint) to the same UNIX
socket used by the stand-alone Mock plug-in instance.

Please note that `GOCSI_TEST_DEBUG` is not supported when used in
conjunction with `CSI_ENDPOINT` as the test suite cannot read the
`STDOUT` and `STDERR` pipes of an existing process.
