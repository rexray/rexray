package executors

import (
	"encoding/json"
	"regexp"

	// depend upon this tool with a nil import in order to preserve it
	// in the dependency list
	_ "github.com/jteeuwen/go-bindata"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

var (
	executors = map[string]*ExecutorInfoEx{}
	pathRX    = regexp.MustCompile(`^lsx-(.+?)(?:.exe)?$`)
)

func init() {
	for path, bdFunc := range _bindata {
		bd, err := bdFunc()

		if err != nil {
			panic(err)
		}

		executors[path] = &ExecutorInfoEx{
			ExecutorInfo: types.ExecutorInfo{
				Name:         path,
				MD5Checksum:  bd.info.MD5Checksum(),
				Size:         bd.info.Size(),
				LastModified: bd.info.ModTime().Unix(),
			},
		}
	}
}

// ExecutorInfos returns a channel on which all executor information can be
// received.
func ExecutorInfos() <-chan *ExecutorInfoEx {
	c := make(chan *ExecutorInfoEx)
	go func() {
		for _, v := range executors {
			c <- v
		}
		close(c)
	}()
	return c
}

// ExecutorInfoInspect returns the executor info for the provided name.
func ExecutorInfoInspect(name string, data bool) (*ExecutorInfoEx, error) {
	ei, ok := executors[name]
	if !ok {
		return nil, utils.NewNotFoundError(name)
	}
	if !data {
		return ei, nil
	}

	bd, ok := _bindata[ei.Name]
	if !ok {
		return nil, utils.NewNotFoundError(name)
	}
	a, err := bd()
	if err != nil {
		return nil, err
	}
	ei.Data = a.bytes
	return ei, nil
}

// ExecutorInfoEx is an extension of ExecutorInfo
type ExecutorInfoEx struct {
	types.ExecutorInfo
	Data []byte `json:"-"`
}

// MarshalJSON marshals the ExecutorInfoEx to JSON.
func (i *ExecutorInfoEx) MarshalJSON() ([]byte, error) {
	return json.Marshal(&i.ExecutorInfo)
}
