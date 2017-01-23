// +build !libstorage_storage_executor libstorage_storage_executor_rbd

package executor

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
	"strings"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/registry"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/drivers/storage/rbd"
	"github.com/codedellemc/libstorage/drivers/storage/rbd/utils"
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(rbd.Name, newdriver)
}

func newdriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return rbd.Name
}

func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	if !gotil.FileExistsInPath("ceph") {
		return false, nil
	}

	if !gotil.FileExistsInPath("rbd") {
		return false, nil
	}

	if !gotil.FileExistsInPath("ip") {
		return false, nil
	}

	if err := exec.Command("modprobe", "rbd").Run(); err != nil {
		return false, nil
	}

	return true, nil
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	return "", types.ErrNotImplemented
}

// LocalDevices returns a map of the system's local devices.
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	devMap, err := utils.GetMappedRBDs()
	if err != nil {
		return nil, err
	}

	ld := &types.LocalDevices{Driver: d.Name()}
	if len(devMap) > 0 {
		ld.DeviceMap = devMap
	}

	return ld, nil
}

// InstanceID returns the local system's InstanceID.
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	return GetInstanceID(nil, nil)
}

// GetInstanceID returns the instance ID object
func GetInstanceID(
	monIPs []net.IP,
	localIntfs []net.Addr) (*types.InstanceID, error) {
	/* Ceph doesn't have only one unique identifier per client, it can have
	   several. With the way the RBD driver is used, we will see multiple
	   identifiers used, and therefore returning any of those identifiers
	   is actually confusing rather than helpful. Instead, we use the client
	   IP address that is on the interface that can reach the monitors.

	   We loop through all the monitor IPs, looking for a local interface
	   that is on the same L2 segment. If these all fail, We are on an L3
	   segment so we grab the IP from the default route.
	*/

	var err error
	if nil == monIPs {
		monIPs, err = getCephMonIPs()
		if err != nil {
			return nil, err
		}
	}
	if len(monIPs) == 0 {
		return nil, goof.New("No Ceph Monitors found")
	}

	if nil == localIntfs {
		localIntfs, err = net.InterfaceAddrs()
		if err != nil {
			return nil, err
		}
	}

	iid := &types.InstanceID{Driver: rbd.Name}
	for _, intf := range localIntfs {
		localIP, localNet, _ := net.ParseCIDR(intf.String())
		for _, monIP := range monIPs {
			if localNet.Contains(monIP) {
				// Monitor reachable over L2
				iid.ID = localIP.String()
				return iid, nil
			}
		}
	}

	// No luck finding L2 match, check for default/static route to monitor
	localIP, err := getSrcIP(monIPs[0].String())
	if err != nil {
		return nil, err
	}
	iid.ID = localIP

	return iid, nil
}

func getCephMonIPs() ([]net.IP, error) {
	out, err := exec.Command("ceph-conf", "--lookup", "mon_host").Output()
	if err != nil {
		return nil, goof.WithError("Unable to get Ceph monitors", err)
	}

	monStrings := strings.Split(strings.TrimSpace(string(out)), ",")

	monIps := make([]net.IP, 0, 4)

	for _, mon := range monStrings {
		ip := net.ParseIP(mon)
		if ip != nil {
			monIps = append(monIps, ip)
		} else {
			ipSlice, err := net.LookupIP(mon)
			if err == nil {
				monIps = append(monIps, ipSlice...)
			}
		}
	}

	return monIps, nil
}

func getSrcIP(destIP string) (string, error) {
	out, err := exec.Command(
		"ip", "-oneline", "route", "get", destIP).Output()
	if err != nil {
		return "", goof.WithError("Unable get IP routes", err)
	}

	byteReader := bytes.NewReader(out)
	scanner := bufio.NewScanner(byteReader)
	scanner.Split(bufio.ScanWords)
	found := false
	for scanner.Scan() {
		if !found {
			if scanner.Text() == "src" {
				found = true
				continue
			}
		}
		return scanner.Text(), nil
	}
	return "", goof.New("Unable to parse ip output")
}
