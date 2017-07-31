package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/goof"
)

func getDividend() int {
	return 2
}

func getDivisor() int {
	return 0
}

func divide() (int, error) {
	x := getDividend()
	y := getDivisor()
	if y == 0 {
		return -1, goof.WithFields(goof.Fields{
			"dividend": x,
			"divisor":  y,
		}, "divide by zero")
	}

	return x / y, nil
}

func calculate(op string) error {
	switch op {
	case "divide":
		z, err := divide()
		if err != nil {
			return err
		}
		fmt.Printf("division = %d\n", z)
		return nil
	}
	return nil
}

func main() {
	if err := calculate("divide"); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
