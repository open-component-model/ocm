// Code generated by go-bindata. (@generated) DO NOT EDIT.

//Package jsonscheme generated by go-bindata.// sources:
// ../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml
package jsonscheme

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _ResourcesComponentDescriptorOcmV3SchemaYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xe4\x19\x5f\x6f\xdb\xb6\xf3\xdd\x9f\xe2\x80\x04\x90\xdd\x44\x76\x92\xa2\x0f\xd5\x4b\x50\xb4\x2f\x3f\xfc\xb6\x75\x58\x8b\x3d\x2c\xf5\x02\x46\x3a\xd9\x4c\x25\x52\x23\x69\x27\x5e\x9b\xef\x3e\x90\x94\x28\x4a\x96\x6c\xcb\x49\x1f\x86\xbd\x24\xe2\xf1\xee\x78\xbc\xff\x47\x9f\xd2\x24\x82\x60\xa9\x54\x21\xa3\xd9\x6c\x41\x44\x82\x0c\xc5\x34\xce\xf8\x2a\x99\xc9\x78\x89\x39\x91\xb3\x98\xe7\x05\x67\xc8\x54\x98\xa0\x8c\x05\x2d\x14\x17\x21\x8f\xf3\x70\xfd\x9a\x64\xc5\x92\x5c\x06\xa3\x53\x8b\xeb\xf1\xba\x97\x9c\x85\x16\x3a\xe5\x62\x31\x4b\x04\x49\xd5\xec\xea\xe2\xea\x22\xbc\xbc\x2a\x59\x07\xa3\x8a\x21\xe5\x2c\x82\xe0\xe3\xfb\x9f\xe1\x7d\x75\x18\x7c\x70\x87\xc1\xfa\x35\x54\x14\xa7\x09\xa6\x32\x1a\x01\xe4\xa8\x88\xfe\x0f\xa0\x36\x05\x46\x10\xf0\xbb\x7b\x8c\x55\x60\x40\x4d\xbe\xee\x02\xb0\x46\x21\x29\x67\x86\x38\x21\x8a\x58\x6c\x81\x7f\xad\xa8\xc0\xc4\xb2\x03\x08\x21\x60\x24\xc7\xa0\x5e\x96\x74\x16\x42\x92\x84\x6a\xce\x24\xfb\x55\xf0\x02\x85\xa2\x28\x23\x48\x49\x26\xd1\xec\x17\x35\xb4\xe4\xa0\xb9\x55\xdf\x00\xa7\x02\xd3\x08\x82\x93\x99\xb9\x4b\xad\xde\x5f\xbc\x33\xcb\x03\x7b\x89\x04\x66\xe4\x11\x93\x4f\x98\xaf\x51\x54\x44\x19\xb9\xc3\x4c\xf6\xd2\xd8\xed\x0a\xb9\x10\x7c\x4d\x13\x14\xbd\xe8\x15\x42\x45\x10\x0b\x24\xfa\xda\x9f\xa9\x7f\x19\xab\x7c\xa9\x04\x65\x0b\x07\x4c\xb9\xc8\x89\x8a\x20\x21\x0a\x43\x45\x73\x1c\x19\x83\x89\x05\xf6\x5a\x6c\x5b\x69\x24\x5b\x70\x41\xd5\x32\xaf\x0f\x2b\x88\x52\x28\xb4\x49\xff\xbc\x21\xe1\xdf\x73\xfd\xe7\x22\x7c\x3b\xbb\x0d\xe7\x67\xa7\x4e\x4e\xce\x52\xba\x88\xe0\x1b\x3c\x1d\x60\x2e\x5f\x67\xa5\x58\x44\x08\xb2\xb1\xdc\xa8\xc2\xdc\x09\xd4\xa5\xce\xa0\x62\xd1\x7b\xb1\x9d\xce\xb5\xcf\x55\x3a\xb4\x4b\xe2\x18\x65\xbf\x91\xed\xb6\x73\x23\x92\xad\x30\x82\x6f\x4f\x7d\x6e\xe5\x69\x74\x7d\x73\x11\xbe\xf5\xf4\x28\xe9\x82\x51\xb6\x68\x0b\x13\xdc\x71\x9e\x21\x61\x15\x9a\x67\xd6\x0e\x71\xcc\xee\xfe\xb0\x19\x69\xb3\x79\x61\xd0\xd0\xa6\xbd\xbe\x65\x92\x93\xc7\x9f\x90\x2d\xd4\x32\x82\xab\x37\x6f\x46\x9d\x4e\x11\x5a\xaf\x98\xbf\x1a\xdf\x4c\xe7\x2d\xd0\xe4\x55\x05\xfb\x76\x75\xfe\x34\x9e\x35\xb6\x6f\x3b\x48\x6e\x35\xcd\x44\x6b\x65\x04\x40\x13\x64\x8a\xaa\xcd\x3b\xa5\x04\xbd\x5b\x29\xfc\x3f\x6e\xac\xa8\x39\x65\x4e\xae\x2e\xa9\xf4\xe1\xe3\x9b\xf0\xf6\xac\x12\xa4\x02\x4e\xae\x2d\xeb\x46\x40\x5b\x9e\x27\xa0\xc8\x57\x64\x90\x0a\x9e\x83\x34\x1b\x3a\x95\x02\x61\x09\x90\xe4\x7e\x25\x15\x26\xa0\x38\x90\x2c\xe3\x0f\x40\x18\xf0\xc2\xea\x17\x32\x24\x09\x65\x0b\x08\xd6\xc1\x39\xe4\xe4\x5e\xe7\x6b\x96\x6d\xce\x0d\xa9\x59\x4f\x73\xca\x4a\x68\x75\xd6\x92\x4a\xc8\x91\x30\x09\x6a\x89\x90\x72\xcd\x55\x33\xb1\xea\x97\x40\x04\xea\xa3\xb4\x4f\xd1\xa4\x29\xaf\xac\x04\xbe\x9c\x5e\x4d\x5f\xfb\xdf\x61\xca\xf9\xd9\x1d\x11\x25\x6c\xed\x23\xac\xbb\x30\x2e\xa7\x57\xd5\x97\x43\xf3\xf0\xdd\x67\x83\xcc\x57\xf6\x7a\x7e\x3d\xbe\xf8\x7e\x73\x19\xbe\x9d\x7f\x49\x5e\x4d\xc6\xd7\xd1\x97\xa9\x0f\x98\x5c\x77\x83\xc2\xf1\xf8\x3a\xaa\x81\xdf\xbf\x24\xc6\x46\xef\xc2\x3f\xc2\xb9\x8e\x8c\xea\xbb\x62\x79\x20\xf2\xa4\x3a\xf1\x6c\xec\x6f\x9c\x19\x26\x0d\x88\xc1\x2c\xa3\xaf\xe5\xf9\x5d\xae\xb7\x2f\x93\x6e\x74\x1c\x49\x9d\x06\x5b\x21\xd9\xe5\xc4\x01\x3c\x59\x27\x2c\xb8\xa4\x8a\x8b\xcd\x7b\xce\x14\x3e\xaa\x21\x59\x4d\x63\xf5\x65\x35\xc3\xa1\x9d\x48\xbc\xdb\xf9\x49\x6d\x6f\x39\xbf\x23\x12\x0d\x96\xae\x33\x25\x29\x4a\x18\xeb\x15\x3e\x2a\x64\x3a\xc5\xc9\xc9\x1e\x41\x47\x00\x92\xaf\x44\x8c\x1f\x30\xa5\xcc\x64\xa6\x01\xb7\xd5\x69\xda\x2d\xca\xac\xea\xd6\x9a\x83\x5b\x58\xf9\x06\x64\xfb\x46\xbe\xeb\x48\xa9\x9d\xf6\x2b\x91\xf1\x51\x09\xf2\xbf\x12\xa1\x37\x29\x6f\x71\x78\x56\xd7\xb1\xd3\xb6\x16\x38\xa8\x31\x39\xb0\xc0\x8d\x00\x12\xba\x40\xa9\x3e\x15\x18\x0f\xb0\xdc\x92\xc8\xe5\xbb\xaa\xb5\xa8\xed\xa9\x3b\x96\x8c\x4a\xd3\xe1\x6c\x6f\x9b\x3a\x7a\x64\xf3\xd7\x38\x70\x67\x69\xef\x16\x62\x27\x89\x2d\xf0\xdd\x18\x23\x00\xdd\x7b\x49\x45\xf2\xa2\xad\x24\xab\xa3\x1e\x89\x77\x31\x2d\x41\x47\xf6\x80\xba\xa7\x20\x6a\x25\x70\xa0\xd1\xc8\x0e\x8b\xe8\x55\x8e\x09\x25\x9f\xab\xb0\x1b\x6e\xa3\x8e\x5e\x73\xa0\xb2\x2d\xc8\xc9\x51\x63\x35\x73\xd7\xe7\x25\x5a\x24\x9b\xc0\x78\x6a\x8a\xad\x53\x8b\x3d\x22\xe8\xb5\xa7\x43\x3c\x36\x55\xd9\x90\x71\x4b\xc7\xef\x45\x06\x9b\x0e\x85\xd8\xf3\x7a\x83\xb9\x8e\x60\xbf\xeb\xf4\x6e\xd8\x41\xd3\xf0\xa1\xc0\x73\x48\xe3\xe8\xbd\x64\x8d\x50\x30\xe9\x83\xa1\xee\xa0\x3e\x1c\x93\x44\xfa\x74\x7a\x84\x86\xb6\x72\x6e\x07\xce\x33\xd3\xfa\x00\x23\x38\xb5\xb8\x39\xdc\xea\xa7\xbf\x36\x0f\xac\x8d\x2d\x07\x14\x58\x16\x61\x7b\xca\x8b\xb8\xe1\x0f\x9b\xaf\x07\x3b\x73\xeb\x76\x5b\x55\xd2\x9b\x34\xa1\x3d\x6d\x76\x1c\xd0\x76\x58\xdb\xc4\x88\xf8\x37\x4c\x0f\x6c\x9d\x08\x08\x4c\x51\x20\x8b\xd1\x4c\x0e\x30\xae\x1f\x77\x32\x1e\x93\x6c\x52\x36\x45\xc7\xbe\x74\x54\x3e\xf8\x09\x33\x8c\x15\xef\x7f\x5d\xe8\x75\xd6\x03\x7b\x05\xd3\xad\x96\x57\x39\xf6\xf2\xee\xee\x87\x3e\x03\x75\xba\xd2\xf3\x9f\x87\x3a\x46\xdf\x43\xfd\x78\x57\x03\x09\x27\x40\x62\xb5\x22\x59\xb6\x89\xea\x33\x42\x53\x78\x1e\x66\x20\x0b\x8c\x29\xc9\xb4\x93\x2a\x41\x63\x2d\xb2\xfc\xb7\xf4\x9c\x83\x1a\xca\x76\xd8\x72\x86\x1f\x53\x3f\xcc\xc2\x4a\x71\x6c\x95\x65\x41\x63\x63\x77\xaa\x74\xf1\xbd\x7f\x84\xd8\x35\xc2\x54\x6c\xe4\xd0\xe7\x48\x38\x31\xf4\x26\x72\x6b\x2e\xe7\xe5\x2c\xbf\x92\x0a\x72\xa2\xe2\xa5\xe7\xe8\x72\x2b\x21\x7b\x73\x9b\x59\x6a\x7d\x2b\xe7\xcc\x06\xe4\xbf\x29\xfd\x67\x67\x18\x9b\x67\x9f\x9d\xc3\x2d\x9b\xba\x40\x58\x65\xef\xd5\x1f\xb2\x55\x1e\xc1\x4d\x60\x4c\x1d\x9c\x43\xa0\x07\x5d\xc1\x48\x16\xcc\x8f\x09\x89\x41\x8f\x88\x3f\x2e\x7e\x9a\x6f\xd0\x43\x1f\x4f\x5f\xa6\x5d\x3d\x6e\x54\xd5\x85\x77\x57\xcb\xd8\x0c\x77\x93\x6a\x53\x1a\x1b\x5b\x57\x4d\x7f\xcc\x99\x42\xa6\xf4\xd2\x2b\x45\x95\xff\xaa\x63\xef\x58\x26\x81\x67\xfb\x69\x2b\xb1\xd5\x1e\x5b\x96\xd0\x67\x9f\xe0\x38\xb5\xbb\xa5\x17\xe0\xbc\x2d\xfd\xa8\x65\x24\xdf\xb3\x74\x8a\x2b\xe8\xef\x75\x11\x0f\x21\xf8\x4a\x59\x52\x7e\xfa\x3f\x16\x85\xd6\x98\xc1\xa8\xa9\xf8\x9a\xbc\xf7\xe5\xba\x8c\x60\x08\x78\x9c\x4f\x5b\xbf\xb4\xb9\x1f\xd2\xce\xed\xb6\xe4\xa9\x7a\x20\x02\xeb\x0d\xd0\x61\xae\x65\xea\xe5\x1f\x73\x26\x55\x04\x81\x6b\xdc\xbd\xfb\x54\x37\xb0\xc4\x5b\x4f\xf4\xf6\x6a\x5b\x8f\x7f\x47\xfd\x2a\xb2\xc5\x45\x17\xa9\x78\x25\x04\x32\x95\x6d\xce\xe1\x01\x81\xb3\x6c\x53\x3e\x5a\x9b\x42\xc5\x19\x36\xc2\xa9\xed\x89\x65\x43\xed\xe6\xbe\xa3\xe4\x72\xd4\x41\x6b\xf2\x3b\x8a\x5b\xf7\x8c\x14\x8c\xfe\x09\x00\x00\xff\xff\x90\xde\x0b\x5c\x50\x1d\x00\x00")

func ResourcesComponentDescriptorOcmV3SchemaYamlBytes() ([]byte, error) {
	return bindataRead(
		_ResourcesComponentDescriptorOcmV3SchemaYaml,
		"../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml",
	)
}

func ResourcesComponentDescriptorOcmV3SchemaYaml() (*asset, error) {
	bytes, err := ResourcesComponentDescriptorOcmV3SchemaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml": ResourcesComponentDescriptorOcmV3SchemaYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"..": {nil, map[string]*bintree{
		"..": {nil, map[string]*bintree{
			"..": {nil, map[string]*bintree{
				"..": {nil, map[string]*bintree{
					"..": {nil, map[string]*bintree{
						"..": {nil, map[string]*bintree{
							"..": {nil, map[string]*bintree{
								"resources": {nil, map[string]*bintree{
									"component-descriptor-ocm-v3-schema.yaml": {ResourcesComponentDescriptorOcmV3SchemaYaml, map[string]*bintree{}},
								}},
							}},
						}},
					}},
				}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
