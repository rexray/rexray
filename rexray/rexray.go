package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/emccode/rexray/rexray/cli"

	// load REX-Ray
	_ "github.com/emccode/rexray"
)

func init() {
	if p := os.Getenv("REXRAY_TRACE_PROFILE"); p != "" {
		f, err := os.Create(p)
		if err != nil {
			panic(err)
		}
		if err := trace.Start(f); err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	if p := os.Getenv("REXRAY_CPU_PROFILE"); p != "" {
		f, err := os.Create(p)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if p := os.Getenv("REXRAY_PROFILE_ADDR"); p != "" {
		go http.ListenAndServe(p, http.DefaultServeMux)
	}
}

func main() {
	cli.Execute()
}
