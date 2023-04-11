// Code generated by go-bindata. (@generated) DO NOT EDIT.

// Package jsonscheme generated by go-bindata.
// sources:
// ../../../../../../../resources/component-descriptor-v2-schema.yaml
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

var _ResourcesComponentDescriptorV2SchemaYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x1a\x5d\x6f\xdb\xba\xf5\x5d\xbf\xe2\xe0\xa6\x80\x9c\x26\x8a\x5b\x0f\x1d\x70\xfd\x12\x64\x1d\x06\x14\xdb\xbd\x19\xda\x6e\x0f\x4b\xbd\x82\x96\x8e\x6d\x76\x14\xe9\x91\x94\x1b\xb5\xb7\xff\x7d\x20\x29\xea\x5b\xb2\x1d\xa7\xed\x2e\xd0\xbc\x44\xa4\xce\x17\x0f\xcf\xb7\xfc\x84\x26\x73\x08\x37\x5a\x6f\xd5\x7c\x3a\x5d\x13\x99\x20\x47\x79\x15\x33\x91\x25\x53\x15\x6f\x30\x25\x6a\x1a\x8b\x74\x2b\x38\x72\x1d\x25\xa8\x62\x49\xb7\x5a\xc8\x68\x37\x0b\x83\x27\x0e\xa2\x46\xe1\x83\x12\x3c\x72\xbb\x57\x42\xae\xa7\x89\x24\x2b\x3d\x9d\x3d\x9b\x3d\x8b\x9e\xcf\x0a\x82\x61\xe0\xc9\x50\xc1\xe7\x10\xde\x6e\x91\xc3\x4b\xcf\x03\x7e\x11\x09\x32\xd8\xcd\xa0\x82\x5e\x51\x4e\x0d\xb0\x9a\x07\x00\x29\x6a\x62\xfe\x03\xe8\x7c\x8b\x73\x08\xc5\xf2\x03\xc6\x3a\xb4\x5b\x4d\xca\xa5\xe0\x50\x09\x6e\xf1\x13\xa2\x89\x43\x90\xf8\xdf\x8c\x4a\x4c\x1c\x45\x80\x08\x42\xc7\xf7\x9f\x28\x15\x15\xdc\x41\x6d\xa5\xd8\xa2\xd4\x14\x95\x87\x6b\x00\xf9\xcd\x52\x24\xa5\x25\xe5\xeb\x30\x08\x00\x18\x59\x22\x1b\x94\xb7\x87\x3d\x27\x29\x86\xd5\x72\x47\x58\x86\x43\x52\x18\xd8\x41\xe6\x6e\xd3\xe2\xcf\xe1\xf3\x17\xbf\x6e\x8b\xbc\x25\x5a\xa3\x34\xea\xfa\xf7\xee\xee\x59\xf4\xf3\xe2\xe2\x89\xc7\x55\x74\xcd\x29\x5f\x77\x38\x2c\x85\x60\x48\xb8\x3d\x5f\xa9\xe3\x5f\x4b\x59\x7a\xe4\x48\xc9\xfd\xdf\x90\xaf\xf5\x66\x0e\xb3\x17\x2f\x82\x16\xdf\x3b\x12\x7d\x5a\xdc\x45\x24\xfa\x64\xf8\x3f\x9d\xdc\x5d\x2d\x5a\x5b\xe7\x4f\xfd\xde\xe7\xd9\xe5\x97\xc9\xb4\xf1\xfa\x7d\x0f\xca\x7b\x83\x73\x6e\x8e\x12\x00\xd0\x04\xb9\xa6\x3a\xbf\xd1\x5a\xd2\x65\xa6\xf1\xaf\x98\x3b\x51\x53\xca\x4b\xb9\xfa\xa4\x32\xcc\x27\x77\xd1\xfb\x0b\x2f\x88\xdf\x3c\xbf\x76\xa4\x25\x32\x72\x8f\xc9\x1b\x4c\x77\x28\x1d\xcd\x33\xd0\xe4\x3f\xc8\x61\x25\x45\x0a\xca\xbe\x30\xbe\x00\x84\x27\x40\x92\x0f\x99\xd2\x98\x80\x16\x40\x18\x13\x1f\x81\x70\x10\xd6\x5e\x09\x03\x86\x24\xa1\x7c\x0d\xe1\x2e\xbc\x84\x94\x7c\x10\x32\x12\x9c\xe5\x97\x16\xd5\xae\xaf\x52\xca\x8b\x5d\xcf\x6b\x43\x15\xa4\x48\xb8\x02\xbd\x41\x58\x09\x43\xd5\x10\x71\xea\x57\x40\x24\x1a\x56\xc6\x10\x68\xd2\x94\x57\x79\x81\x9f\x5f\xcd\xae\xfe\x50\x7f\x8e\x56\x42\x5c\x2c\x89\x2c\xf6\x76\x75\x80\x5d\x1f\xc4\xf3\xab\x99\x7f\x2a\xc1\x6a\xf0\xe5\x63\x03\xad\xae\xec\xdd\xe2\x7a\xf2\xec\xb7\xbb\xe7\xd1\xcf\x8b\x77\xc9\xd3\xf3\xc9\xf5\xfc\xdd\x55\x7d\xe3\xfc\xba\x7f\x2b\x9a\x4c\xae\xe7\xd5\xe6\x6f\xef\x12\x7b\x47\x37\xd1\xbf\xa2\x85\x31\x67\xff\xec\x49\x1e\x08\x7c\xee\x39\x5e\x4c\xea\x2f\x2e\x2c\x91\xc6\x8e\x85\x2c\x5c\xa6\xeb\xfe\x1d\xd3\x1b\x0c\x05\x85\x77\xe7\xc6\x8f\xd4\x1c\x3e\xc3\x13\x89\xab\x39\x84\x67\xd3\x5a\x00\x9c\xf6\x99\x72\x08\x5f\x9c\x29\x6e\x85\xa2\x5a\xc8\xfc\xa5\xe0\x1a\xef\xf5\x31\x51\xc7\x40\x0d\x45\x19\x4b\x61\x24\xc4\x89\x98\xbe\xee\xe7\x4d\x18\xbb\x5d\x55\x5c\x7a\x4f\xd4\x11\xbb\x0a\x7e\x6d\x39\x0b\x59\x97\x44\xe1\x3f\x24\x0b\xab\x10\xd6\x11\xd9\xfc\x15\x60\xf5\xad\x81\x18\x69\xfe\x1a\x71\xec\x17\xb2\xdd\x36\x02\xdf\x28\x2a\x00\xf2\x2c\x9d\xc3\x5d\x98\x49\xf6\x77\xa2\x37\xe1\x25\x84\x6a\x43\x66\x2f\xfe\x18\x25\x74\x8d\x4a\x87\x8b\xa0\x45\xe7\x58\xca\x56\xc7\x6b\xaa\xb4\xcc\x0d\xf5\xdb\x97\xaf\xca\xe5\xc2\xdc\x01\x89\x63\x54\xea\xc0\xbc\x68\x34\x63\xa1\x60\x25\x64\x81\x8a\x0a\x26\x66\x85\xf7\x1a\xb9\xc9\x10\xea\x7c\x8f\xb1\x04\x00\x6b\xaa\x37\xd9\xf2\x66\x9c\xf7\xa8\xb5\xd9\xa5\x31\x81\xda\x85\xda\x9d\xd5\x83\xac\xb1\xad\x36\x27\x60\xa9\xfe\x82\xd1\x1e\x74\x63\xa5\xe3\x10\xb1\x48\x53\xaa\xc7\x7c\x82\x0b\x8e\xa7\xe8\xe5\xc4\x73\xff\x2a\x38\x3a\xc3\x50\x22\x93\x31\xfe\xb9\x74\xb8\x23\xc4\x31\xd5\x45\xb9\x28\xea\x86\x72\x6d\x28\x94\x0b\x67\x42\x0f\x2f\x52\xe0\x88\x60\x57\xa0\xe0\xbd\x96\xe4\x55\x01\x30\x3f\x92\x4e\x38\x54\x0c\x0d\x44\xa8\x5a\xce\x0c\x0f\xbf\x0e\x5b\xfa\xa9\x0e\x10\x91\x92\xe4\xd5\xc9\xa9\xc6\xb4\x11\xb7\x7a\x65\xb0\xb4\x3c\x52\xdd\xd9\xed\x9a\xe7\xb7\xab\x66\x90\xec\x25\xe2\xf0\xc2\xfd\x80\x75\xbf\x3e\x00\xdc\x54\xff\x1e\x38\x00\x70\x31\xef\xcd\x16\xe3\x23\x8c\x6d\x43\xd4\xe6\x86\xad\x85\xa4\x7a\x93\x56\x26\x28\x64\x4a\x18\x55\xc4\x30\xea\xbe\xb6\xc5\xed\x80\xd9\x35\x08\xb6\x2f\xc1\x5d\x94\x37\xd0\x5e\x26\xa3\x28\xae\xaa\xee\x87\x08\x5c\xe1\x4c\x74\x26\xf1\x48\x25\x90\x91\x13\x9a\x55\x8a\x09\x25\x6f\xbd\xe7\x75\xcf\x4c\x4e\x16\xde\x6d\x95\x7c\x2a\xa8\x66\x06\x79\xbb\x41\x07\xe4\xd2\x88\x58\xd9\xe2\xb3\x3c\x36\xd4\xba\x96\x51\xfd\x3c\x34\x1a\x39\x13\x2b\x97\x25\xbd\x23\x42\x50\xe3\xc0\x8e\xde\x9e\x38\x50\xd9\x75\xbd\x41\xaa\x9d\x63\x10\xb3\x61\x0f\xd6\x47\x94\x8c\x5f\xfb\x34\xb3\x37\x5f\x13\x93\x92\x50\x22\x8f\xd1\x36\x0e\x30\xa9\x5a\x72\x26\x62\xc2\xce\x8b\x30\x3f\x94\x3b\x7c\x00\x7c\x83\x0c\x63\x2d\xe4\x43\xe3\xe5\x57\x88\x68\xf5\x16\xf2\xb5\x3f\xe5\x43\xf5\x52\x52\x3a\xb4\xbb\x6e\x54\x7d\xf5\xae\x7b\xbc\xfb\xef\x69\x7a\x07\xcf\xd9\xcb\x62\x2c\x27\xc2\x19\x90\x58\x67\x84\xb1\x7c\x5e\x71\x8a\xac\xa3\x7d\x9c\x82\xda\x62\x4c\x09\x03\x89\x06\x3e\xb6\x4c\x7e\xbf\x69\xf4\xab\xe5\xc8\xb6\x47\x0b\x8e\xed\x1c\x59\xf0\xe2\x19\x63\x07\x24\xb9\xba\xf7\xdb\x6e\xcb\xb9\x5c\x15\x25\x8f\x2c\xbb\x3d\x01\x75\xf0\x24\xa8\xb0\x49\x38\xb3\xf8\xd6\xf1\x2b\x2a\x97\xc5\x24\x20\x53\x1a\x52\xa2\xe3\x4d\xcd\x19\x54\xa7\x7a\xeb\x56\xe0\xcc\x66\xbf\xda\x56\xbd\x58\xf8\x51\xd4\x95\xa7\x72\x81\xfb\x91\x2c\xd6\x11\xab\xfa\x0e\x77\x09\x07\x57\xf9\xd6\x04\x4c\x3b\x68\x9a\x36\xc9\x09\x2b\x1b\x9d\xdf\x6f\xe9\x29\x62\xfa\x27\x26\x0e\xaf\x3d\xad\x0e\xfe\x42\x19\xaa\x5c\x69\x4c\x8f\xc7\xbd\xed\x63\xf8\xb5\xa3\x87\x88\xe9\xab\x94\xac\x4f\x6a\x11\xed\x92\x1a\x2a\x65\xde\x7c\x94\xde\xb1\x3e\x6a\xf0\xf6\xd4\x64\xb3\x67\x18\x54\xa9\xf3\x84\x83\x31\x92\x7b\xbf\x3c\xed\x3c\x10\x16\x22\x85\x50\x8d\x01\x56\x43\x85\xed\x8d\x39\x40\xb3\xac\x30\x95\x6d\x4a\x38\x5d\xa1\xd2\xed\x92\xb6\xc5\xf4\x81\x75\xb3\xd3\x8c\x0b\xe0\xce\x51\x9c\x04\x0a\xb4\xd8\xc3\xb1\x6d\xa8\x5d\x76\x0e\xc2\xb3\xd2\x44\xae\x51\x63\x02\xb1\xe0\xba\x2c\x94\x06\xc9\x2b\xfa\x69\xf4\x2c\xe6\x3d\x50\x0e\xcb\x5c\xa3\xf2\x3c\x96\x46\xd9\x6d\xba\x3c\x4b\x97\xe6\x42\x03\x80\x41\x97\x3d\xc1\x5c\x56\x94\x61\x95\x2f\x4f\xb5\x98\x1e\x09\x2b\xeb\xf1\xac\x86\xf4\xe2\xdf\xd7\xd5\x01\x7a\x43\x34\x50\x65\xcf\x6e\xd4\x4f\xb9\x7d\xf7\x93\x79\xa9\x7e\x82\x84\x4a\x5b\x98\xe7\x83\xf7\xe1\xf5\x76\xfb\x48\xfe\xf5\x15\x14\x76\xdb\xf6\xb3\x71\xe3\x6c\x1a\xa6\xf5\x77\xf8\x48\xf5\xa6\x50\x4d\x9c\x49\x89\x5c\x43\xdf\x47\xbc\x31\x2d\xf9\xd0\xfa\xba\xa8\x8c\x4e\xf9\xf6\x56\xef\x02\xfa\x94\xf8\xa3\x46\xda\x9f\x4b\xec\x65\x7c\xfb\xc2\x64\xa8\xb8\xa8\xa5\xdd\x6f\x93\xec\x03\x80\x6a\x3e\x76\x82\xc3\x66\x7e\x40\x7e\x62\x7a\x37\xc2\x94\xd7\x91\x8d\x0c\xc3\x03\x80\x35\x72\x94\x34\xfe\x8e\x83\xec\x42\x02\x37\xcb\x2e\x16\x3f\x3c\xfb\xff\xc0\xb3\xab\x8b\x71\xfb\xdf\xd7\xb1\x1b\x86\xfa\xad\x8a\xf8\x32\x33\x1d\x3c\xae\x3a\x7a\x3e\xd5\xb5\xd3\xce\xe7\x52\x55\x7b\xb9\x95\x62\x47\x93\xea\x46\x23\x08\x1b\x43\x86\xe6\xcc\xab\xac\xe7\x55\x83\x7e\x03\x63\x9f\xed\x1f\x3e\xf2\x3a\xc1\x30\x63\x89\xb6\x31\x7e\x4b\xf7\x8c\x71\x01\x56\x42\xa6\x44\xcf\x21\x21\x1a\x23\x4d\xcb\x91\x71\x57\x6d\x47\x9b\x6a\xa7\xd5\x1d\xeb\x61\x3b\x1f\xc4\x43\x38\xf3\x25\x0d\xcb\x2f\xe1\x23\x82\xe0\x2c\x2f\x7e\x04\x62\x2b\x7f\xc1\xbd\xb0\xfe\x1a\xbf\xd7\xf7\xa5\xc2\x02\x1e\x69\xc6\xd1\xfa\x00\x59\x5e\x6a\xd7\x0c\x1f\x87\x61\x97\x70\x35\x60\x79\xe8\xc9\x0e\xbf\xfb\xfa\x5c\x30\x3c\xd0\x58\x1a\xf5\xea\x41\x48\xad\x4c\x68\xc3\x51\xbf\x4a\xe1\xf3\x97\x20\x08\x5a\xb1\xa9\x1e\x78\x22\x08\x53\x74\x3f\x87\xab\x07\x87\x30\x68\xba\x7e\xf5\xb3\xbb\x5e\x81\x3c\x89\x56\x4c\x1c\xbf\xa0\xb0\xfe\x29\xa8\x59\x5f\xd4\x2e\xa4\x71\x19\xe3\x9f\x57\xc2\xff\x05\x00\x00\xff\xff\x99\x92\x92\xbd\xd0\x28\x00\x00")

func ResourcesComponentDescriptorV2SchemaYamlBytes() ([]byte, error) {
	return bindataRead(
		_ResourcesComponentDescriptorV2SchemaYaml,
		"../../../../../../../resources/component-descriptor-v2-schema.yaml",
	)
}

func ResourcesComponentDescriptorV2SchemaYaml() (*asset, error) {
	bytes, err := ResourcesComponentDescriptorV2SchemaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../../../../../../resources/component-descriptor-v2-schema.yaml", size: 10448, mode: os.FileMode(436), modTime: time.Unix(1681215048, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
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
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
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
	"../../../../../../../resources/component-descriptor-v2-schema.yaml": ResourcesComponentDescriptorV2SchemaYaml,
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
// AssetDir("foo.txt") and AssetDir("nonexistent") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
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
									"component-descriptor-v2-schema.yaml": {ResourcesComponentDescriptorV2SchemaYaml, map[string]*bintree{}},
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
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
