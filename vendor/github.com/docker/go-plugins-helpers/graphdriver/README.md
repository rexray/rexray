# Docker volume extension api.

Go handler to create external graphdriver extensions for Docker.

## Usage

This library is designed to be integrated in your program.

1. Implement the `graphdriver.Driver` interface.
2. Initialize a `graphdriver.Handler` with your implementation.
3. Call either `ServeTCP` or `ServeUnix` from the `graphdriver.Handler`.

### Example using TCP sockets:

```go
  d := MyGraphDriver{}
  h := graphdriver.NewHandler(d)
  h.ServeTCP("test_graph", ":8080")
```

### Example using Unix sockets:

```go
  d := MyGraphDriver{}
  h := graphdriver.NewHandler(d)
  h.ServeUnix("root", "test_graph")
```
