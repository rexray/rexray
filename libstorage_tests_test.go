package libstorage

import (
	"testing"

	"github.com/emccode/libstorage/api/client"
)

func TestRoot(t *testing.T) {
	skipIfTestRun(t)

	tf := func(client client.Client, t *testing.T) {
		reply, err := client.Root()
		if err != nil {
			t.Fatal(err)
		}
		testLogAsJSON(reply, t)
	}

	testTCP(tf, t)
	testTCPTLS(tf, t)
	testSock(tf, t)
	testSockTLS(tf, t)
}

func TestVolumes(t *testing.T) {
	skipIfTestRun(t)

	tf := func(client client.Client, t *testing.T) {
		reply, err := client.Volumes()
		if err != nil {
			t.Fatal(err)
		}
		testLogAsJSON(reply, t)
	}

	testTCP(tf, t)
	testTCPTLS(tf, t)
	testSock(tf, t)
	testSockTLS(tf, t)
}
