package csi_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	modcsi "github.com/AVENTER-UG/rexray/agent/csi"
)

const (
	sockFile          = ".gocsi.sock"
	readBufLen  int64 = 1024
	writeBufLen int64 = 1024
)

var (
	benchmarkCount       = 100
	concurrentGets       = 100
	responseDataSz int64 = 1024 * 1024 // 1MiB
	rg                   = rand.New(rand.NewSource(time.Now().UnixNano()))
	writeBuf             = make([]byte, writeBufLen)
)

func init() {
	if v := os.Getenv("CSI_BENCHMARK_COUNT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			benchmarkCount = i
		}
	}
	if v := os.Getenv("CSI_BENCHMARK_COCUR"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			concurrentGets = i
		}
	}
	if v := os.Getenv("CSI_BENCHMARK_RESSZ"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			responseDataSz = i
		}
	}
}

func BenchmarkPipe(b *testing.B) {

	// create a new pipe connection. this acts as both
	// the listener used by the http server to accept
	// new connections and also provides the http client's
	// dialer
	pipcon := modcsi.NewPipeConn("BenchmarkPipeConn")

	// create a client using the pipe connection as the
	// client's dialer. any dial attempt will cause the
	// server to accept a new connection -- the other
	// side of the pipe
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: pipcon.DialHTTP17,
		},
	}

	benchmarkNetConnParallel(b, pipcon, client)
}

func BenchmarkTCP(b *testing.B) {
	benchmarkTCPOrUnix(b, "tcp", "127.0.0.1:9090")
}

func BenchmarkUnix(b *testing.B) {
	defer os.RemoveAll(sockFile)
	benchmarkTCPOrUnix(b, "unix", sockFile)
}

func benchmarkTCPOrUnix(
	b *testing.B, network, addr string) {

	listen, err := net.Listen(network, addr)
	if err != nil {
		b.Fatalf("error: listen %s://%s failed: %v", network, addr, err)
	}
	defer os.RemoveAll(sockFile)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(
				context.Context, string, string) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	benchmarkNetConnParallel(b, listen, client)
}

func benchmarkNetConnParallel(
	b *testing.B,
	listen net.Listener,
	client *http.Client) {

	var (
		ctx    = context.Background()
		closed = make(chan int)
		server = newHTTPServer(b)
	)

	// start the http serve
	go func() {
		server.Serve(listen)
		closed <- 1
	}()

	// make sure the server is shutdown
	cctx, cancel := context.WithTimeout(
		ctx, time.Duration(1)*time.Second)
	defer server.Shutdown(cctx)
	defer cancel()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest(http.MethodGet, "http://host/", nil)
			if err != nil {
				b.Fatalf("error: new http request failed: %v", err)
			}

			// set the header "r" to a random number that will be
			// sent back as a response header. this enables the
			// ability to assert the piped connections are tracking
			// the client/server pairings correctly, even under high,
			// concurrent load
			szrand := fmt.Sprintf("%d", rg.Int())
			req.Header.Set("r", szrand)

			res, err := client.Do(req)
			if err != nil {
				b.Fatalf("error: http request failed: %v", err)
			}

			if res.StatusCode != 200 {
				b.Fatalf("error: http status not okay: %d", res.StatusCode)
			}

			if v := res.Header.Get("r"); v != szrand {
				b.Fatalf(
					"error: rand roundtrip failed: exp: %s, act: %s",
					szrand, v)
			}

			if res.Body == nil {
				b.Fatalf("error: http response body nil")
			}

			n, err := readData(res.Body)
			if err != nil {
				b.Fatalf(
					"error: failed to read http response body: %v",
					err)
			}

			if n64 := int64(n); n64 != responseDataSz {
				b.Fatalf(
					"error: http response body len invalid: exp: %d, act: %d",
					responseDataSz, n64)
			}
		}
	})
}

func newHTTPServer(b *testing.B) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("r", r.Header.Get("r"))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", responseDataSz))
		n, err := writeData(w, responseDataSz)
		if err != nil {
			b.Fatalf("error: write http response failed: %v", err)
		}
		if n != responseDataSz {
			b.Fatalf("error: wrote %d bytes, expected %d", n, responseDataSz)
		}
	})
	return &http.Server{Handler: mux}
}

func readData(r io.Reader) (int, error) {

	var (
		read int = 0
		rbuf     = make([]byte, readBufLen)
	)

	// loop until all the data is read
	for {
		n, err := r.Read(rbuf)
		read += n
		if err != nil {
			if err == io.EOF {
				return read, nil
			}
			return read, err
		}
		if n == 0 {
			return read, nil
		}
	}
}

func writeData(w io.Writer, sizeOfData int64) (int64, error) {
	var written int64 = 0

	// loop until all the data is written
	for written < sizeOfData {

		var (
			// an io.Reader used to copy the data buffer to w
			r = bytes.NewReader(writeBuf)

			// the number of bytes that need to be written
			towrite int64 = sizeOfData - written
		)

		// loop until towrite bytes from r are copied to w
		for {
			n, err := io.CopyN(w, r, towrite)
			written += n
			towrite -= n
			if err != nil {
				if err == io.EOF {
					break
				}
				return written, err
			}
			if n == 0 {
				break
			}
		}
	}

	return written, nil
}
