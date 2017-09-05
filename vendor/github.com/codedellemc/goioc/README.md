# GoIoC [![Build Status](http://travis-ci.org/codedellemc/goioc.svg?branch=master)](https://travis-ci.org/codedellemc/goioc)
GoIoC introduces inversion of control (IoC) to Go by providing a basic
type registry.

## Registering a New Type
The first step to using GoIoC is registering a new type:

```go
type reader struct {
}

func (r *reader) Read(data []byte) (int, error) {
	return 0, nil
}

func main() {
	// Register a new type "reader" and a func() interface{} that
	// returns a new instance of the reader struct.
	goioc.Register("reader", func() interface{} { return &reader{} })
}
```

The GoIoC type registry now has a type named "reader" that can be
recalled at anytime. Please note, however, that subsequent calls to
`Register` will override existing types with the same name.

## Constructing a Registered Type
The `New` function is used to construct a new instance of a registered
type:

```go
func main() {
	// Construct a new instance of the registered type "reader",
	// assert that it conforms to the io.Reader interface, and
	// then read some of its data into the data buffer.
	r := goioc.New("reader").(io.Reader)
	data := make([]byte, 1024)
	r.Read(data)
}
```

## Request Types that Implement an Interface
GoIoC even makes it simple to request new instances of all types that
implement a specific interface:

```go
func main() {
	// Print the type name of all the registered types that implement
	// io.Reader. The code below will print *main.reader.
	for o := range goioc.Implements((*io.Reader)(nil)) {
		fmt.Printf("%T\n", o)
	}
}
