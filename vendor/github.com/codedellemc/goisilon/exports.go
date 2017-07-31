package goisilon

import (
	"golang.org/x/net/context"

	api "github.com/codedellemc/goisilon/api/v2"
)

type ExportList []*api.Export
type Export *api.Export
type UserMapping *api.UserMapping

// GetExports returns a list of all exports on the cluster
func (c *Client) GetExports(ctx context.Context) (ExportList, error) {
	return api.ExportsList(ctx, c.API)
}

// GetExportByID returns an export with the provided ID.
func (c *Client) GetExportByID(ctx context.Context, id int) (Export, error) {
	return api.ExportInspect(ctx, c.API, id)
}

// GetExportByName returns the first export with a path for the provided
// volume name.
func (c *Client) GetExportByName(
	ctx context.Context, name string) (Export, error) {

	exports, err := api.ExportsList(ctx, c.API)
	if err != nil {
		return nil, err
	}
	path := c.API.VolumePath(name)
	for _, ex := range exports {
		for _, p := range *ex.Paths {
			if p == path {
				return ex, nil
			}
		}
	}
	return nil, nil
}

// Export the volume with a given name on the cluster
func (c *Client) Export(ctx context.Context, name string) (int, error) {

	ok, id, err := c.IsExported(ctx, name)
	if err != nil {
		return 0, err
	}
	if ok {
		return id, nil
	}

	paths := []string{c.API.VolumePath(name)}

	return api.ExportCreate(
		ctx, c.API,
		&api.Export{Paths: &paths})
}

// GetRootMapping returns the root mapping for an Export.
func (c *Client) GetRootMapping(
	ctx context.Context, name string) (UserMapping, error) {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	return ex.MapRoot, nil
}

// GetRootMappingByID returns the root mapping for an Export.
func (c *Client) GetRootMappingByID(
	ctx context.Context, id int) (UserMapping, error) {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	return ex.MapRoot, nil
}

// EnableRootMapping enables the root mapping for an Export.
func (c *Client) EnableRootMapping(
	ctx context.Context, name, user string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapRoot: ex.MapRoot}

	setUserMapping(
		nex,
		user,
		true,
		func(e Export) UserMapping { return e.MapRoot },
		func(e Export, m UserMapping) { e.MapRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// EnableRootMappingByID enables the root mapping for an Export.
func (c *Client) EnableRootMappingByID(
	ctx context.Context, id int, user string) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapRoot: ex.MapRoot}

	setUserMapping(
		nex,
		user,
		true,
		func(e Export) UserMapping { return e.MapRoot },
		func(e Export, m UserMapping) { e.MapRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// DisableRootMapping disables the root mapping for an Export.
func (c *Client) DisableRootMapping(
	ctx context.Context, name string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapRoot: ex.MapRoot}

	setUserMapping(
		nex,
		"nobody",
		false,
		func(e Export) UserMapping { return e.MapRoot },
		func(e Export, m UserMapping) { e.MapRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// DisableRootMappingbyID disables the root mapping for an Export.
func (c *Client) DisableRootMappingByID(
	ctx context.Context, id int) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapRoot: ex.MapRoot}

	setUserMapping(
		nex,
		"nobody",
		false,
		func(e Export) UserMapping { return e.MapRoot },
		func(e Export, m UserMapping) { e.MapRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// GetNonRootMapping returns the map_non_root mapping for an Export.
func (c *Client) GetNonRootMapping(
	ctx context.Context, name string) (UserMapping, error) {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	return ex.MapNonRoot, nil
}

// GetNonRootMappingByID returns the map_non_root mapping for an Export.
func (c *Client) GetNonRootMappingByID(
	ctx context.Context, id int) (UserMapping, error) {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	return ex.MapNonRoot, nil
}

// EnableNonRootMapping enables the map_non_root mapping for an Export.
func (c *Client) EnableNonRootMapping(
	ctx context.Context, name, user string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapNonRoot: ex.MapNonRoot}

	setUserMapping(
		nex,
		user,
		true,
		func(e Export) UserMapping { return e.MapNonRoot },
		func(e Export, m UserMapping) { e.MapNonRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// EnableNonRootMappingByID enables the map_non_root mapping for an Export.
func (c *Client) EnableNonRootMappingByID(
	ctx context.Context, id int, user string) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapNonRoot: ex.MapNonRoot}

	setUserMapping(
		nex,
		user,
		true,
		func(e Export) UserMapping { return e.MapNonRoot },
		func(e Export, m UserMapping) { e.MapNonRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// DisableNonRootMapping disables the map_non_root mapping for an Export.
func (c *Client) DisableNonRootMapping(
	ctx context.Context, name string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapNonRoot: ex.MapNonRoot}

	setUserMapping(
		nex,
		"nobody",
		false,
		func(e Export) UserMapping { return e.MapNonRoot },
		func(e Export, m UserMapping) { e.MapNonRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// DisableNonRootMappingByID disables the map_non_root mapping for an Export.
func (c *Client) DisableNonRootMappingByID(
	ctx context.Context, id int) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapNonRoot: ex.MapNonRoot}

	setUserMapping(
		nex,
		"nobody",
		false,
		func(e Export) UserMapping { return e.MapNonRoot },
		func(e Export, m UserMapping) { e.MapNonRoot = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// GetFailureMapping returns the map_failure mapping for an Export.
func (c *Client) GetFailureMapping(
	ctx context.Context, name string) (UserMapping, error) {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	return ex.MapFailure, nil
}

// GetFailureMappingByID returns the map_failure mapping for an Export.
func (c *Client) GetFailureMappingByID(
	ctx context.Context, id int) (UserMapping, error) {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	return ex.MapFailure, nil
}

// EnableFailureMapping enables the map_failure mapping for an Export.
func (c *Client) EnableFailureMapping(
	ctx context.Context, name, user string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapFailure: ex.MapFailure}

	setUserMapping(
		nex,
		user,
		true,
		func(e Export) UserMapping { return e.MapFailure },
		func(e Export, m UserMapping) { e.MapFailure = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// EnableFailureMappingByID enables the map_failure mapping for an Export.
func (c *Client) EnableFailureMappingByID(
	ctx context.Context, id int, user string) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapFailure: ex.MapFailure}

	setUserMapping(
		nex,
		user,
		true,
		func(e Export) UserMapping { return e.MapFailure },
		func(e Export, m UserMapping) { e.MapFailure = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// DisableFailureMapping disables the map_failure mapping for an Export.
func (c *Client) DisableFailureMapping(
	ctx context.Context, name string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapFailure: ex.MapFailure}

	setUserMapping(
		nex,
		"nobody",
		false,
		func(e Export) UserMapping { return e.MapFailure },
		func(e Export, m UserMapping) { e.MapFailure = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

// DisableFailureMappingByID disables the map_failure mapping for an Export.
func (c *Client) DisableFailureMappingByID(
	ctx context.Context, id int) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	nex := &api.Export{ID: ex.ID, MapFailure: ex.MapFailure}

	setUserMapping(
		nex,
		"nobody",
		false,
		func(e Export) UserMapping { return e.MapFailure },
		func(e Export, m UserMapping) { e.MapFailure = m })

	return api.ExportUpdate(ctx, c.API, nex)
}

func setUserMapping(
	ex Export,
	user string,
	enabled bool,
	getMapping func(Export) UserMapping,
	setMapping func(Export, UserMapping)) {

	m := getMapping(ex)
	if m == nil {
		m = &api.UserMapping{
			User: &api.Persona{
				ID: &api.PersonaID{
					ID:   user,
					Type: api.PersonaIDTypeUser,
				},
			},
		}
		setMapping(ex, m)
		return
	}

	if m.Enabled != nil || !enabled {
		m.Enabled = &enabled
	}

	if m.User == nil {
		m.User = &api.Persona{
			ID: &api.PersonaID{
				ID:   user,
				Type: api.PersonaIDTypeUser,
			},
		}
		return
	}

	u := m.User
	if u.ID != nil {
		u.ID.ID = user
		return
	}

	u.Name = &user
}

// GetExportClients returns an Export's clients property.
func (c *Client) GetExportClients(
	ctx context.Context, name string) ([]string, error) {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	if ex.Clients == nil {
		return nil, nil
	}
	return *ex.Clients, nil
}

// GetExportClientsByID returns an Export's clients property.
func (c *Client) GetExportClientsByID(
	ctx context.Context, id int) ([]string, error) {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	if ex.Clients == nil {
		return nil, nil
	}
	return *ex.Clients, nil
}

// AddExportClients adds to the Export's clients property.
func (c *Client) AddExportClients(
	ctx context.Context, name string, clients ...string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}
	addClients := ex.Clients
	if addClients == nil {
		addClients = &clients
	} else {
		*addClients = append(*addClients, clients...)
	}
	return api.ExportUpdate(
		ctx, c.API, &api.Export{ID: ex.ID, Clients: addClients})
}

// AddExportClientsByID adds to the Export's clients property.
func (c *Client) AddExportClientsByID(
	ctx context.Context, id int, clients ...string) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}
	addClients := ex.Clients
	if addClients == nil {
		addClients = &clients
	} else {
		*addClients = append(*addClients, clients...)
	}
	return api.ExportUpdate(
		ctx, c.API, &api.Export{ID: ex.ID, Clients: addClients})
}

// SetExportClients sets the Export's clients property.
func (c *Client) SetExportClients(
	ctx context.Context, name string, clients ...string) error {

	ok, id, err := c.IsExported(ctx, name)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return api.ExportUpdate(ctx, c.API, &api.Export{ID: id, Clients: &clients})
}

// SetExportClientsByID sets the Export's clients property.
func (c *Client) SetExportClientsByID(
	ctx context.Context, id int, clients ...string) error {

	return api.ExportUpdate(ctx, c.API, &api.Export{ID: id, Clients: &clients})
}

// ClearExportClients sets the Export's clients property to nil.
func (c *Client) ClearExportClients(
	ctx context.Context, name string) error {

	return c.SetExportClients(ctx, name, []string{}...)
}

// ClearExportClientsByID sets the Export's clients property to nil.
func (c *Client) ClearExportClientsByID(
	ctx context.Context, id int) error {

	return c.SetExportClientsByID(ctx, id, []string{}...)
}

// GetExportRootClients returns an Export's root_clients property.
func (c *Client) GetExportRootClients(
	ctx context.Context, name string) ([]string, error) {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	if ex.RootClients == nil {
		return nil, nil
	}
	return *ex.RootClients, nil
}

// GetExportRootClientsByID returns an Export's clients property.
func (c *Client) GetExportRootClientsByID(
	ctx context.Context, id int) ([]string, error) {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}
	if ex.RootClients == nil {
		return nil, nil
	}
	return *ex.RootClients, nil
}

// AddExportRootClients adds to the Export's root_clients property.
func (c *Client) AddExportRootClients(
	ctx context.Context, name string, clients ...string) error {

	ex, err := c.GetExportByName(ctx, name)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}
	addClients := ex.RootClients
	if addClients == nil {
		addClients = &clients
	} else {
		*addClients = append(*addClients, clients...)
	}
	return api.ExportUpdate(
		ctx, c.API, &api.Export{ID: ex.ID, RootClients: addClients})
}

// AddExportRootClientsByID adds to the Export's root_clients property.
func (c *Client) AddExportRootClientsByID(
	ctx context.Context, id int, clients ...string) error {

	ex, err := c.GetExportByID(ctx, id)
	if err != nil {
		return err
	}
	if ex == nil {
		return nil
	}
	addClients := ex.RootClients
	if addClients == nil {
		addClients = &clients
	} else {
		*addClients = append(*addClients, clients...)
	}
	return api.ExportUpdate(
		ctx, c.API, &api.Export{ID: ex.ID, RootClients: addClients})
}

// SetExportRootClients sets the Export's root_clients property.
func (c *Client) SetExportRootClients(
	ctx context.Context, name string, clients ...string) error {

	ok, id, err := c.IsExported(ctx, name)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return api.ExportUpdate(
		ctx, c.API, &api.Export{ID: id, RootClients: &clients})
}

// SetExportRootClientsByID sets the Export's clients property.
func (c *Client) SetExportRootClientsByID(
	ctx context.Context, id int, clients ...string) error {

	return api.ExportUpdate(
		ctx, c.API, &api.Export{ID: id, RootClients: &clients})
}

// ClearExportRootClients sets the Export's root_clients property to nil.
func (c *Client) ClearExportRootClients(
	ctx context.Context, name string) error {

	return c.SetExportRootClients(ctx, name, []string{}...)
}

// ClearExportRootClientsByID sets the Export's clients property to nil.
func (c *Client) ClearExportRootClientsByID(
	ctx context.Context, id int) error {

	return c.SetExportRootClientsByID(ctx, id, []string{}...)
}

// Stop exporting a given volume from the cluster
func (c *Client) Unexport(
	ctx context.Context, name string) error {

	ok, id, err := c.IsExported(ctx, name)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return c.UnexportByID(ctx, id)
}

// UnexportByID unexports an Export by its ID.
func (c *Client) UnexportByID(
	ctx context.Context, id int) error {

	return api.Unexport(ctx, c.API, id)
}

// IsExported returns a flag and export ID if the provided volume name is
// already exported.
func (c *Client) IsExported(
	ctx context.Context, name string) (bool, int, error) {

	export, err := c.GetExportByName(ctx, name)
	if err != nil {
		return false, 0, err
	}
	if export == nil {
		return false, 0, nil
	}
	return true, export.ID, nil
}
