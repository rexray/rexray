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

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/rbd"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/rbd/utils"
)

var (
	// ctxConfigKey is an interface-wrapped key used to access a possible
	// config object in the context
	ctxConfigKey = interface{}("rbd.config")
)

type driver struct {
	config     gofig.Config
	doModprobe bool
}

func init() {
	registry.RegisterStorageExecutor(rbd.Name, newdriver)
}

func newdriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(context types.Context, config gofig.Config) error {
	d.config = config
	d.doModprobe = config.GetBool(rbd.ConfigTestModule)
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

	if d.doModprobe {
		cmd := exec.Command("modprobe", "rbd")
		if _, _, err := utils.RunCommand(ctx, cmd); err != nil {
			return false, nil
		}
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

	devMap, err := utils.GetMappedRBDs(ctx)
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

	// Inject the config into the context, so that the utils package can use
	// to pick up any extra Ceph args
	ctx = ctx.WithValue(ctxConfigKey, d.config)

	return GetInstanceID(ctx, nil, nil)
}

// GetInstanceID returns the instance ID object
func GetInstanceID(
	ctx types.Context,
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

	if ctx == nil {
		ctx = context.Background()
	}

	var err error
	if nil == monIPs {
		monIPs, err = getCephMonIPs(ctx)
		if err != nil {
			return nil, err
		}
	}
	if len(monIPs) == 0 {
		return nil, goof.New("no ceph monitors found")
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
	localIP, err := getSrcIP(ctx, monIPs[0].String())
	if err != nil {
		return nil, err
	}
	iid.ID = localIP

	return iid, nil
}

func getCephMonIPs(ctx types.Context) ([]net.IP, error) {

	cmd := exec.Command("ceph-conf", "--lookup", "mon_host")
	out, _, err := utils.RunCommand(ctx, cmd)
	if err != nil {
		return nil, goof.WithError("unable to get ceph monitors", err)
	}

	monStrings := strings.Split(strings.TrimSpace(string(out)), ",")

	monIps, err := utils.ParseMonitorAddresses(monStrings)
	if err != nil {
		return nil, err
	}

	return monIps, nil
}

func getSrcIP(
	ctx types.Context,
	destIP string) (string, error) {

	cmd := exec.Command("ip", "-oneline", "route", "get", destIP)
	out, _, err := utils.RunCommand(ctx, cmd)
	if err != nil {
		return "", goof.WithError("unable get ip routes", err)
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
	return "", goof.New("unable to parse ip output")
}
