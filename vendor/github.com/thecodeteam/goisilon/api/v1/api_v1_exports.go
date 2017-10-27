package v1

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/thecodeteam/goisilon/api"
)

// Export enables an NFS export on the cluster to access the volumes.  Return the path to the export
// so other processes can mount the volume directory
func Export(
	ctx context.Context,
	client api.Client,
	path string) (err error) {
	// PAPI call: POST https://1.2.3.4:8080/platform/1/protocols/nfs/exports/
	//            Content-Type: application/json
	//            {paths: ["/path/to/volume"]}

	if path == "" {
		return errors.New("no path set")
	}

	var data = &ExportPathList{Paths: []string{path}}
	data.MapAll.User = client.User()
	if group := client.Group(); group != "" {
		data.MapAll.Groups = append(data.MapAll.Groups, group)
	}
	var resp *postIsiExportResp

	err = client.Post(ctx, exportsPath, "", nil, nil, data, &resp)

	if err != nil {
		return err
	}

	return nil
}

// SetExportClients limits access to an NFS export on the cluster to a specific client address.
func SetExportClients(
	ctx context.Context,
	client api.Client,
	Id int, clients []string) (err error) {
	// PAPI call: PUT https://1.2.3.4:8080/platform/1/protocols/nfs/exports/Id
	//            Content-Type: application/json
	//            {clients: ["client_ip_address"]}

	var data = &ExportClientList{Clients: clients}
	var resp *postIsiExportResp

	err = client.Put(ctx, exportsPath, strconv.Itoa(Id), nil, nil, data, &resp)

	return err
}

// Unexport disables the NFS export on the cluster that points to the volumes directory.
func Unexport(
	ctx context.Context,
	client api.Client,
	Id int) (err error) {
	// PAPI call: DELETE https://1.2.3.4:8080/platform/1/protocols/nfs/exports/23

	if Id == 0 {
		return errors.New("no path Id set")
	}

	exportPath := fmt.Sprintf("%s/%d", exportsPath, Id)

	var resp postIsiExportResp
	err = client.Delete(ctx, exportPath, "", nil, nil, &resp)

	return err
}

// GetIsiExports queries a list of all exports on the cluster
func GetIsiExports(
	ctx context.Context,
	client api.Client) (resp *getIsiExportsResp, err error) {

	// PAPI call: GET https://1.2.3.4:8080/platform/1/protocols/nfs/exports
	err = client.Get(ctx, exportsPath, "", nil, nil, &resp)

	return resp, err
}
