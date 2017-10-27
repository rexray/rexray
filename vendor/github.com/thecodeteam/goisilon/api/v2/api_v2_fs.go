package v2

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"

	"github.com/thecodeteam/goisilon/api"
	"context"
)

// ContainerChild is a child object of a container.
type ContainerChild struct {
	Name  *string   `json:"name,omitempty"`
	Path  *string   `json:"container_path,omitempty"`
	Type  *string   `json:"type,omitempty"`
	Owner *string   `json:"owner,omitempty"`
	Group *string   `json:"group,omitempty"`
	Mode  *FileMode `json:"mode,omitempty"`
	Size  *int      `json:"size,omitempty"`
}

type resumeableContainerChildList struct {
	Children []*ContainerChild `json:"children,omitempty"`
	Resume   string            `json:"resume,omitempty"`
}

// ContainerChildList is a list of a container's children.
type ContainerChildList []*ContainerChild

// MarshalJSON marshals a ContainerChildList to JSON.
func (l ContainerChildList) MarshalJSON() ([]byte, error) {
	containers := struct {
		Children []*ContainerChild `json:"children,omitempty"`
	}{l}
	return json.Marshal(containers)
}

// UnmarshalJSON unmarshals a ContainerChildList from JSON.
func (l *ContainerChildList) UnmarshalJSON(text []byte) error {
	containers := struct {
		Children []*ContainerChild `json:"children,omitempty"`
	}{}
	if err := json.Unmarshal(text, &containers); err != nil {
		return err
	}
	*l = containers.Children
	return nil
}

// ContainerQuery is used to query a container.
type ContainerQuery struct {
	Result []string             `json:"result,omitempty"`
	Scope  *ContainerQueryScope `json:"scope,omitempty"`
}

// ContainerQueryScope is the query's scope.
type ContainerQueryScope struct {
	Logic      string        `json:"logic,omitempty"`
	Conditions []interface{} `json:"conditions,omitempty"`
}

// ContainerQueryScopeCondition is the query's condition.
type ContainerQueryScopeCondition struct {
	Operator string `json:"operator,omitempty"`
	Attr     string `json:"attr,omitempty"`
	Value    string `json:"value,omitempty"`
}

var containerChildrenGetAllDetail = []string{
	"type",
	"container_path",
	"size",
	"mode",
	"owner",
	"group",
	"name",
}

var containerQueryAll = &ContainerQuery{
	Result: containerChildrenGetAllDetail,
}

var containerQueryAllSubDirs = &ContainerQuery{
	Result: containerChildrenGetAllDetail,
	Scope: &ContainerQueryScope{
		Logic: "and",
		Conditions: []interface{}{
			&ContainerQueryScopeCondition{
				Operator: "=",
				Attr:     "type",
				Value:    "container",
			},
		},
	},
}

var (
	trueByteArr      = []byte("true")
	falseByteArr     = []byte("false")
	overwriteByteArr = []byte("overwrite")
	recursiveByteArr = []byte("recursive")
	typeByteArr      = []byte("type")
	queryByteArr     = []byte("query")
	limitByteArr     = []byte("limit")
	resumeByteArr    = []byte("resume")
	maxDepthByteArr  = []byte("max-depth")
	sortQS           = [][]byte{[]byte("sort")}
	detailQS         = [][]byte{[]byte("detail")}
	sortDirAsc       = [][]byte{[]byte("dir"), []byte("ASC")}
	sortDirDesc      = [][]byte{[]byte("dir"), []byte("DESC")}
)

func to2DByteArray(sa []string) [][]byte {
	ba := make([][]byte, len(sa))
	for i := 0; i < len(sa); i++ {
		ba[i] = []byte(sa[i])
	}
	return ba
}

// ContainerChildrenGetQuery queries a container for children regardless of
// ACLs preventing traversal.
func ContainerChildrenGetQuery(
	ctx context.Context,
	client api.Client,
	containerPath string,
	limit, maxDepth int,
	objectType, sortDir string,
	sort, detail []string) (<-chan *ContainerChild, <-chan error) {

	var (
		ec  = make(chan error)
		cc  = make(chan *ContainerChild)
		wg  = &sync.WaitGroup{}
		rnp = realNamespacePath(client)
		qs  = api.OrderedValues{
			{queryByteArr},
			{limitByteArr, []byte(fmt.Sprintf("%d", limit))},
			{maxDepthByteArr, []byte(fmt.Sprintf("%d", maxDepth))},
		}
	)

	if objectType != "" {
		qs.Set(typeByteArr, []byte(objectType))
	}
	if len(sort) > 0 {
		if strings.EqualFold(sortDir, "asc") {
			qs = append(qs, sortDirAsc)
		} else {
			qs = append(qs, sortDirDesc)
		}
		qs = append(qs, append(sortQS, to2DByteArray(sort)...))
	}
	if len(detail) > 0 {
		qs = append(qs, append(detailQS, to2DByteArray(detail)...))
	}

	go func() {
		for {
			var resp resumeableContainerChildList
			if err := client.Get(
				ctx,
				rnp,
				containerPath,
				qs,
				nil,
				&resp); err != nil {
				ec <- err
				close(ec)
				close(cc)
				return
			}
			wg.Add(1)
			go func(resp *resumeableContainerChildList) {
				defer wg.Done()
				for _, c := range resp.Children {
					cc <- c
				}
			}(&resp)
			if resp.Resume == "" {
				break
			}
			qs.Set(resumeByteArr, []byte(resp.Resume))
		}
		wg.Wait()
		close(ec)
		close(cc)
	}()
	return cc, ec
}

// ContainerChildrenGetAll GETs all descendent children of a container.
func ContainerChildrenGetAll(
	ctx context.Context,
	client api.Client,
	containerPath string) ([]*ContainerChild, error) {

	var children []*ContainerChild

	cc, ec := ContainerChildrenGetQuery(
		ctx, client, containerPath,
		2, -1, "", "", nil, containerChildrenGetAllDetail)

	done := make(chan int)

	go func() {
		defer close(done)
		for c := range cc {
			children = append(children, c)
		}
	}()

	for {
		select {
		case <-done:
			return children, nil
		case e := <-ec:
			if e != nil {
				return nil, e
			}
		}
	}
}

// ContainerChildrenMapAll GETs all descendent children of a container and
// returns a map with the children's paths as the key.
func ContainerChildrenMapAll(
	ctx context.Context,
	client api.Client,
	containerPath string) (map[string]*ContainerChild, error) {

	resp, err := ContainerChildrenGetAll(ctx, client, containerPath)
	if err != nil {
		return nil, err
	}

	children := map[string]*ContainerChild{}
	for _, v := range resp {
		children[path.Join(*v.Path, *v.Name)] = v
	}

	return children, nil
}

// ContainerChildrenPostQuery queries a container for children with additional
// traversal options and matching capabilities, but is subject to ACLs that
// may prevent traversal.
func ContainerChildrenPostQuery(
	ctx context.Context,
	client api.Client,
	containerPath string,
	limit, maxDepth int,
	query *ContainerQuery) ([]*ContainerChild, error) {

	var resp ContainerChildList

	if err := client.Post(
		ctx,
		realNamespacePath(client),
		containerPath,
		api.OrderedValues{
			{queryByteArr},
			{limitByteArr, []byte(fmt.Sprintf("%d", limit))},
			{maxDepthByteArr, []byte(fmt.Sprintf("%d", maxDepth))},
		},
		nil,
		query,
		&resp); err != nil {

		return nil, err
	}

	return resp, nil
}

var (
	contCreateTTQueryString = api.OrderedValues{
		{recursiveByteArr, trueByteArr},
	}
	contCreateFTQueryString = api.OrderedValues{
		{overwriteByteArr, falseByteArr},
		{recursiveByteArr, trueByteArr},
	}
	contCreateFFQueryString = api.OrderedValues{
		{overwriteByteArr, falseByteArr},
	}
)

// ContainerCreateDir creates a directory as a child object of a container.
func ContainerCreateDir(
	ctx context.Context,
	client api.Client,
	containerPath, dirName string,
	fileMode FileMode,
	overwrite, recursive bool) error {

	var params api.OrderedValues
	if overwrite && recursive {
		params = contCreateTTQueryString
	} else if !overwrite && !recursive {
		params = contCreateFFQueryString
	} else if !overwrite && recursive {
		params = contCreateFTQueryString
	}

	return client.Put(
		ctx,
		realNamespacePath(client),
		path.Join(containerPath, dirName),
		params,
		map[string]string{
			"x-isi-ifs-target-type":    "container",
			"x-isi-ifs-access-control": fileMode.String(),
		},
		nil,
		nil)
}

var (
	overwriteFalseQueryString = api.NewOrderedValues([][]string{
		{"overwrite", "false"},
	})
	recursiveTrueQueryString = api.NewOrderedValues([][]string{
		{"recursive", "true"},
	})
)

// ContainerCreateFile creates a file as a child object of a container.
func ContainerCreateFile(
	ctx context.Context,
	client api.Client,
	containerPath, fileName string,
	fileSize int,
	fileMode FileMode,
	fileHndl io.ReadCloser,
	overwrite bool) error {

	var params api.OrderedValues
	if !overwrite {
		params = overwriteFalseQueryString
	}

	return client.Put(
		ctx,
		realNamespacePath(client),
		path.Join(containerPath, fileName),
		params,
		map[string]string{
			"x-isi-ifs-target-type":    "object",
			"x-isi-ifs-access-control": fileMode.String(),
			"Content-Length":           fmt.Sprintf("%d", fileSize),
		},
		fileHndl,
		nil)
}

// ContainerChildDelete deletes a child of a container.
func ContainerChildDelete(
	ctx context.Context,
	client api.Client,
	childPath string,
	recursive bool) error {

	var params api.OrderedValues
	if recursive {
		params = recursiveTrueQueryString
	}

	return client.Delete(
		ctx,
		realNamespacePath(client),
		childPath,
		params,
		nil,
		nil)
}
