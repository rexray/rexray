package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
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
		log.Errorf("error dividing by 0 with %d / %d", x, y)
		return -1, fmt.Errorf("error dividing by 0 with %d / %d", x, y)
	}

	return x / y, nil
}

func calculate(op string) error {
	switch op {
	case "divide":
		z, err := divide()
		if err != nil {
			log.Errorf("division error %v", err)
			return fmt.Errorf("division error %v", err)
		}
		fmt.Printf("division = %d\n", z)
		return nil
	}
	return nil
}

func main() {
	if err := calculate("divide"); err != nil {
		log.Errorf("calculation error %v", err)
		os.Exit(1)
	}
}
