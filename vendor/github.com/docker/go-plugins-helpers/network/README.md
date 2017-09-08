# Docker network extension API

Go handler to create external network extensions for Docker.

## Usage

This library is designed to be integrated in your program.

1. Implement the `network.Driver` interface.
2. Initialize a `network.Handler` with your implementation.
3. Call either `ServeTCP`, `ServeUnix` or `ServeWindows` from the `network.Handler`.
4. On Windows, docker daemon data dir must be provided for ServeTCP and ServeWindows functions.
On Unix, this parameter is ignored.

### Example using TCP sockets:

```go
  import "github.com/docker/go-plugins-helpers/network"

  d := MyNetworkDriver{}
  h := network.NewHandler(d)
  h.ServeTCP("test_network", ":8080", "")
  // on windows:
  h.ServeTCP("test_network", ":8080", WindowsDefaultDaemonRootDir())
```

### Example using Unix sockets:

```go
  import "github.com/docker/go-plugins-helpers/network"

  d := MyNetworkDriver{}
  h := network.NewHandler(d)
  h.ServeUnix("test_network", 0)
```

### Example using Windows named pipes:

```go
import "github.com/docker/go-plugins-helpers/network"
import "github.com/docker/go-plugins-helpers/sdk"

d := MyNetworkDriver{}
h := network.NewHandler(d)

config := sdk.WindowsPipeConfig{
  // open, read, write permissions for everyone 
  // (uses Windows Security Descriptor Definition Language)
  SecurityDescriptor: AllowServiceSystemAdmin,
  InBufferSize:       4096,
  OutBufferSize:      4096,
}

h.ServeWindows("//./pipe/testpipe", "test_network", WindowsDefaultDaemonRootDir(), &config)
```

## Full example plugins

- [docker-ovs-plugin](https://github.com/gopher-net/docker-ovs-plugin) - An Open vSwitch Networking Plugin
