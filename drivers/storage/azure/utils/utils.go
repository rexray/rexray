// +build !libstorage_storage_driver libstorage_storage_driver_azure

package utils

import (
	"bufio"
	"bytes"
	"os"

	"github.com/akutz/goof"

	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/azure"
)

func checkAzureMarkInFile(ctx types.Context) bool {
	file := "/var/lib/dhcp/dhclient.eth0.leases"
	pattern := []byte("unknown-245")

	f, err := os.Open(file)
	if err != nil {
		ctx.Debug("Specific file (" + file + ") could not be opened:")
		ctx.Debug(err)
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if bytes.Contains(scanner.Bytes(), pattern) {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		ctx.Debugf("Specific file %s could not be read: %s", file, err)
	}
	return false
}

// IsAzureInstance returns a flag indicating whether the executing host
// is an Azure instance .
func IsAzureInstance(ctx types.Context) (bool, error) {
	// http://blog.mszcool.com/index.php/2015/04/
	// detecting-if-a-virtual-machine-runs-in-microsoft-azure-linux-
	// windows-to-protect-your-software-when-distributed-via-the-
	// azure-marketplace/
	if id := os.Getenv("AZURE_INSTANCE_ID"); id != "" {
		return true, nil
	}
	result := checkAzureMarkInFile(ctx)
	return result, nil
}

// InstanceID returns the instance ID for the local host.
func InstanceID(ctx types.Context) (*types.InstanceID, error) {
	hostname := os.Getenv("AZURE_INSTANCE_ID")
	if hostname == "" {
		isAzure, err := IsAzureInstance(ctx)
		if err != nil {
			return nil, err
		}
		if !isAzure {
			return nil, goof.New("Executing outside of Instance.")
		}

		// UUID can be obtained as descried in
		// https://azure.microsoft.com/en-us/blog/accessing-and-using-
		// azure-vm-unique-id/
		// but this code will use hostname as ID

		hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	} else {
		ctx.Info("Use InstanceID from env " + hostname)
	}
	return &types.InstanceID{
		ID:     hostname,
		Driver: azure.Name,
	}, nil
}
