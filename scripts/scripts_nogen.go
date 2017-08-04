// +build !scripts_generated
// +build !agent
// +build !controller

package scripts

import "os"

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) { return nil, nil }

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte { return nil }

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) { return nil, nil }

// AssetNames returns the names of the assets.
func AssetNames() []string { return nil }

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) { return nil, nil }

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error { return nil }

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error { return nil }
