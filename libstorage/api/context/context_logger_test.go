package context

import (
	"fmt"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func BenchmarkStandardLog1(b *testing.B) {
	benchmarkStandardLog(b, 1)
}

func BenchmarkStandardLog10(b *testing.B) {
	benchmarkStandardLog(b, 10)
}

func BenchmarkStandardLog100(b *testing.B) {
	benchmarkStandardLog(b, 100)
}

func BenchmarkStandardLog1000(b *testing.B) {
	benchmarkStandardLog(b, 1000)
}

func benchmarkStandardLog(b *testing.B, f int) {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stderr)
	fields := log.Fields{
		"time": time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}
	for i := 0; i < f; i++ {
		fields[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("val-%d", i)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		log.WithFields(fields).Debug("hi")
	}
}

func BenchmarkContextLogDebug1(b *testing.B) {
	benchmarkContextLog(b, 1, log.DebugLevel)
}

func BenchmarkContextLogDebug10(b *testing.B) {
	benchmarkContextLog(b, 10, log.DebugLevel)
}

func BenchmarkContextLogDebug100(b *testing.B) {
	benchmarkContextLog(b, 100, log.DebugLevel)
}

func BenchmarkContextLogDebug1000(b *testing.B) {
	benchmarkContextLog(b, 1000, log.DebugLevel)
}

func BenchmarkContextLogError1(b *testing.B) {
	benchmarkContextLog(b, 1, log.ErrorLevel)
}

func BenchmarkContextLogError10(b *testing.B) {
	benchmarkContextLog(b, 10, log.ErrorLevel)
}

func BenchmarkContextLogError100(b *testing.B) {
	benchmarkContextLog(b, 100, log.ErrorLevel)
}

func BenchmarkContextLogError1000(b *testing.B) {
	benchmarkContextLog(b, 1000, log.ErrorLevel)
}

func benchmarkContextLog(b *testing.B, f int, lvl log.Level) {
	log.SetLevel(lvl)
	log.SetOutput(os.Stderr)
	keys := make([]string, f)
	vals := make([]string, f)
	for i := 0; i < f; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		vals[i] = fmt.Sprintf("val-%d", i)
		RegisterCustomKey(keys[i], CustomLoggerKey)
	}
	ctx := Background()
	for i := 0; i < f; i++ {
		ctx = ctx.WithValue(keys[i], vals[i])
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ctx.WithField("customField", "customValue").Debug("hi")
	}
}
