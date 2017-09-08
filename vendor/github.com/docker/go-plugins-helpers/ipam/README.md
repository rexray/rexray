# Docker IPAM extension API

Go handler to create external IPAM extensions for Docker.

## Usage

This library is designed to be integrated in your program.

1. Implement the `ipam.Driver` interface.
2. Initialize a `ipam.Handler` with your implementation.
3. Call either `ServeTCP` or `ServeUnix` from the `ipam.Handler`.

### Example using TCP sockets:

```go
  import "github.com/docker/go-plugins-helpers/ipam"

  d := MyIPAMDriver{}
  h := ipam.NewHandler(d)
  h.ServeTCP("test_ipam", ":8080")
```

### Example using Unix sockets:

```go
  import "github.com/docker/go-plugins-helpers/ipam"

  d := MyIPAMDriver{}
  h := ipam.NewHandler(d)
  h.ServeUnix("root", "test_ipam")
```
