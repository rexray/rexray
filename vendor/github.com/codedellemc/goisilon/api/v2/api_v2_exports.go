package v2

import (
	"errors"
	"strconv"

	"golang.org/x/net/context"

	"github.com/codedellemc/goisilon/api"
	"github.com/codedellemc/goisilon/api/json"
)

// Export is an Isilon Export.
type Export struct {
	ID          int          `json:"id,omitmarshal"`
	Paths       *[]string    `json:"paths,omitempty"`
	Clients     *[]string    `json:"clients,omitempty"`
	RootClients *[]string    `json:"root_clients,omitempty"`
	MapAll      *UserMapping `json:"map_all,omitempty"`
	MapRoot     *UserMapping `json:"map_root,omitempty"`
	MapNonRoot  *UserMapping `json:"map_non_root,omitempty"`
	MapFailure  *UserMapping `json:"map_failure,omitempty"`
}

// ExportList is a list of Isilon Exports.
type ExportList []*Export

// MarshalJSON marshals an ExportList to JSON.
func (l ExportList) MarshalJSON() ([]byte, error) {
	exports := struct {
		Exports []*Export `json:"exports,omitempty"`
	}{l}
	return json.Marshal(exports)
}

// UnmarshalJSON unmarshals an ExportList from JSON.
func (l *ExportList) UnmarshalJSON(text []byte) error {
	exports := struct {
		Exports []*Export `json:"exports,omitempty"`
	}{}
	if err := json.Unmarshal(text, &exports); err != nil {
		return err
	}
	*l = exports.Exports
	return nil
}

// ExportList GETs all exports.
func ExportsList(
	ctx context.Context,
	client api.Client) ([]*Export, error) {

	var resp ExportList

	if err := client.Get(
		ctx,
		exportsPath,
		"",
		nil,
		nil,
		&resp); err != nil {

		return nil, err
	}

	return resp, nil
}

// ExportInspect GETs an export.
func ExportInspect(
	ctx context.Context,
	client api.Client,
	id int) (*Export, error) {

	var resp ExportList

	if err := client.Get(
		ctx,
		exportsPath,
		strconv.Itoa(id),
		nil,
		nil,
		&resp); err != nil {

		return nil, err
	}

	if len(resp) == 0 {
		return nil, nil
	}

	return resp[0], nil
}

// ExportCreate POSTs an Export object to the Isilon server.
func ExportCreate(
	ctx context.Context,
	client api.Client,
	export *Export) (int, error) {

	if export.Paths != nil && len(*export.Paths) == 0 {
		return 0, errors.New("no path set")
	}

	var resp Export

	if err := client.Post(
		ctx,
		exportsPath,
		"",
		nil,
		nil,
		export,
		&resp); err != nil {

		return 0, err
	}

	return resp.ID, nil
}

// ExportUpdate PUTs an Export object to the Isilon server.
func ExportUpdate(
	ctx context.Context,
	client api.Client,
	export *Export) error {

	return client.Put(
		ctx,
		exportsPath,
		strconv.Itoa(export.ID),
		nil,
		nil,
		export,
		nil)
}

// ExportDelete DELETEs an Export object on the Isilon server.
func ExportDelete(
	ctx context.Context,
	client api.Client,
	id int) error {

	return client.Delete(
		ctx,
		exportsPath,
		strconv.Itoa(id),
		nil,
		nil,
		nil)
}

// SetExportClients sets an Export's clients property.
func SetExportClients(
	ctx context.Context,
	client api.Client,
	id int,
	addrs ...string) error {

	return ExportUpdate(ctx, client, &Export{ID: id, Clients: &addrs})
}

// SetExportRootClients sets an Export's root_clients property.
func SetExportRootClients(
	ctx context.Context,
	client api.Client,
	id int,
	addrs ...string) error {

	return ExportUpdate(ctx, client, &Export{ID: id, RootClients: &addrs})
}

// Unexport is an alias for ExportDelete.
func Unexport(
	ctx context.Context,
	client api.Client,
	id int) error {

	return ExportDelete(ctx, client, id)
}
