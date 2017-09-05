package goioc_test

import (
	"io"
	"testing"

	"github.com/codedellemc/goioc"
)

func TestNew(t *testing.T) {
	r := goioc.New("reader")
	if r != nil {
		t.Fatal("r should be nil")
	}

	goioc.Register(
		"reader",
		func() interface{} { return &testReader{} })

	r = goioc.New("reader")
	if r == nil {
		t.Fatal("r should not be nil")
	}
	if _, ok := r.(io.Reader); !ok {
		t.Fatal("r should implement io.Reader")
	}

	w := goioc.New("writer")
	if w != nil {
		t.Fatal("w should be nil")
	}

	goioc.Register(
		"writer",
		func() interface{} { return &testWriter{} })

	w = goioc.New("writer")
	if w == nil {
		t.Fatal("w should not be nil")
	}
	if _, ok := w.(io.Writer); !ok {
		t.Fatal("w should implement io.Writer")
	}

	// overwrite the previous "writer"
	goioc.Register(
		"writer",
		func() interface{} { return &testWriteCloser{testWriter{}} })

	w = goioc.New("writer")
	if w == nil {
		t.Fatal("w should not be nil")
	}
	if _, ok := w.(io.WriteCloser); !ok {
		t.Fatal("w should implement io.WriteCloser")
	}
}

func TestImplements(t *testing.T) {
	goioc.Register(
		"reader",
		func() interface{} { return &testReader{} })
	goioc.Register(
		"writer",
		func() interface{} { return &testWriter{} })
	goioc.Register(
		"writeCloser",
		func() interface{} { return &testWriteCloser{testWriter{}} })

	var (
		hasReader      = false
		hasWriter      = false
		hasWriteCloser = false
	)

	for o := range goioc.Implements((*io.Reader)(nil)) {
		if _, ok := o.(io.Reader); ok {
			hasReader = true
		}
	}
	for o := range goioc.Implements((*io.Writer)(nil)) {
		if _, ok := o.(io.Writer); ok {
			hasWriter = true
		}
	}
	for o := range goioc.Implements((*io.WriteCloser)(nil)) {
		if _, ok := o.(io.WriteCloser); ok {
			hasWriteCloser = true
		}
	}

	if !hasReader {
		t.Error("no reader")
		t.Fail()
	}
	if !hasWriter {
		t.Error("no writer")
		t.Fail()
	}
	if !hasWriteCloser {
		t.Error("no write closer")
		t.Fail()
	}
}

type testReader struct {
}

func (tr *testReader) Read(data []byte) (int, error) {
	return 0, nil
}

type testWriter struct {
}

func (tw *testWriter) Write(data []byte) (int, error) {
	return 0, nil
}

type testWriteCloser struct {
	testWriter
}

func (twc *testWriteCloser) Close() error {
	return nil
}
