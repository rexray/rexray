package main

import (
	"log"
	"os"

	client "github.com/appropriate/go-virtualboxclient/virtualboxclient"
)

func main() {
	url := "http://127.0.0.1:18083"
	if len(os.Args) >= 2 {
		url = os.Args[1]
	}

	client := client.New("", "", url)
	if err := client.Logon(); err != nil {
		log.Fatalf("Unable to log on to vboxwebsrv: %v\n", err)
	}
}
