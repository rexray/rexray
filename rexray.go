package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emccode/rexray/storagedriver"
	_ "github.com/emccode/rexray/storagedriver/ec2"
	_ "github.com/emccode/rexray/storagedriver/rackspace"
	"gopkg.in/yaml.v2"
)

var debug string
var storageDrivers string

func init() {
	debug = strings.ToUpper(os.Getenv("REXRAY_DEBUG"))
	storageDrivers = strings.ToLower(os.Getenv("REXRAY_STORAGEDRIVERS"))
}

func main() {
	drivers, err := storagedriver.GetDrivers(storageDrivers)
	if err != nil && debug == "TRUE" {
		fmt.Println(err)
	}

	var allBlockDevices []*storagedriver.BlockDevice
	for _, driver := range drivers {
		blockDevices, err := driver.GetBlockDeviceMapping()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		if len(blockDevices.([]*storagedriver.BlockDevice)) > 0 {
			for _, blockDevice := range blockDevices.([]*storagedriver.BlockDevice) {
				allBlockDevices = append(allBlockDevices, blockDevice)
			}
		}
	}

	if len(allBlockDevices) > 0 {
		yamlOutput, err := yaml.Marshal(&allBlockDevices)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf(string(yamlOutput))
	}

}
