package main

import "os"
import "github.com/emccode/rexray/rexray/commands"

func main() {
	defer func() {
		r := recover()
		if r == nil {
			os.Exit(0)
		}
		switch r := r.(type) {
		case int:
			os.Exit(r)
		case error:
			panic(r)
		default:
			os.Exit(1)
		}
	}()

	commands.Exec()
}
