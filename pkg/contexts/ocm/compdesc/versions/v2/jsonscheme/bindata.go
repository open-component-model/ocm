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

var _ResourcesComponentDescriptorV2SchemaYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x1a\x6d\x6f\xdb\xb8\xf9\xbb\x7e\xc5\x83\x4b\x01\x39\x4d\x14\xb7\x1e\x3a\xe0\xfc\x25\xc8\x7a\xd8\x50\x6c\xbb\x0c\xed\x6d\x1f\x96\x7a\x05\x2d\x3d\xb6\xd9\x51\xa4\x47\x52\x6e\xd4\x5e\xff\xfb\x40\x52\xd4\xbb\x64\x3b\x4e\xdb\x3b\xa0\xf9\x12\x91\x7a\xde\xf9\xbc\x52\x7e\x42\x93\x39\x84\x1b\xad\xb7\x6a\x3e\x9d\xae\x89\x4c\x90\xa3\xbc\x8a\x99\xc8\x92\xa9\x8a\x37\x98\x12\x35\x8d\x45\xba\x15\x1c\xb9\x8e\x12\x54\xb1\xa4\x5b\x2d\x64\xb4\x9b\x85\xc1\x13\x07\x51\xa3\xf0\x5e\x09\x1e\xb9\xdd\x2b\x21\xd7\xd3\x44\x92\x95\x9e\xce\x9e\xcd\x9e\x45\xcf\x67\x05\xc1\x30\xf0\x64\xa8\xe0\x73\x08\xff\x52\x70\x85\x97\x9e\x0f\xfc\x54\xf2\x81\xdd\x0c\x2a\xb4\x15\xe5\xd4\x60\xa9\x79\x00\x90\xa2\x26\xe6\x3f\x80\xce\xb7\x38\x87\x50\x2c\xdf\x63\xac\x43\xbb\xd5\x64\x51\x6a\x00\x95\x06\x16\x3f\x21\x9a\x38\x04\x89\xff\xcb\xa8\xc4\xc4\x51\x04\x88\x20\x74\x7c\xff\x85\x52\x51\xc1\x1d\xd4\x56\x8a\x2d\x4a\x4d\x51\x79\xb8\x06\x90\xdf\x2c\x45\x52\x5a\x52\xbe\x0e\x83\x00\x80\x91\x25\xb2\x41\x79\x7b\xd8\x73\x92\x62\x58\x2d\x77\x84\x65\x38\x24\x85\x81\x1d\x64\xee\x36\x2d\xfe\x1c\x3e\x7d\xf6\xeb\xb6\xc8\x5b\xa2\x35\x4a\x63\xae\xff\xec\xee\x9e\x45\x3f\x2e\x2e\x9e\x78\x5c\x45\xd7\x9c\xe8\x4c\x76\x79\x2c\x85\x60\x48\xb8\xd5\xb0\xb4\xf2\xcf\xa5\x34\x3d\x92\xa4\xe4\xfe\x6f\xc8\xd7\x7a\x33\x87\xd9\x8b\x17\x41\x8b\xf3\x1d\x89\x3e\x2e\xee\x22\x12\x7d\x34\x12\x3c\x9d\xdc\x5d\x2d\x5a\x5b\xe7\x4f\xfd\xde\xa7\xd9\xe5\xe7\xc9\xb4\xf1\xfa\x5d\x0f\xca\x3b\x83\x73\x6e\x94\x09\x00\x68\x82\x5c\x53\x9d\xdf\x68\x2d\xe9\x32\xd3\xf8\x57\xcc\x9d\xa8\x29\xe5\xa5\x5c\x7d\x52\x19\xe6\x93\xbb\xe8\xdd\x85\x17\xc4\x6f\x9e\x5f\x3b\xd2\x12\x19\xb9\xc7\xe4\x0d\xa6\x3b\x94\x8e\xe6\x19\x68\xf2\x5f\xe4\xb0\x92\x22\x05\x65\x5f\x98\xb0\x00\xc2\x13\x20\xc9\xfb\x4c\x69\x4c\x40\x0b\x20\x8c\x89\x0f\x40\x38\x08\xeb\xb1\x84\x01\x43\x92\x50\xbe\x86\x70\x17\x5e\x42\x4a\xde\x0b\x19\x09\xce\xf2\x4b\x8b\x6a\xd7\x57\x29\xe5\xc5\xae\xe7\xb5\xa1\x0a\x52\x24\x5c\x81\xde\x20\xac\x84\xa1\x6a\x88\x38\xf3\x2b\x20\x12\x0d\x2b\xe3\x0a\x34\x69\xca\xab\xbc\xc0\xcf\xaf\x66\x57\x7f\xa8\x3f\x47\x2b\x21\x2e\x96\x44\x16\x7b\xbb\x3a\xc0\xae\x0f\xe2\xf9\xd5\xcc\x3f\x95\x60\x35\xf8\xf2\xb1\x81\x56\x37\xf6\x6e\x71\x3d\x79\xf6\xeb\xdd\xf3\xe8\xc7\xc5\xdb\xe4\xe9\xf9\xe4\x7a\xfe\xf6\xaa\xbe\x71\x7e\xdd\xbf\x15\x4d\x26\xd7\xf3\x6a\xf3\xd7\xb7\x89\x3d\xa3\x9b\xe8\xdf\xd1\xc2\x38\xb4\x7f\xf6\x24\x0f\x04\x3e\xf7\x1c\x2f\x26\xf5\x17\x17\x96\x48\x63\xc7\x42\x16\x41\xd3\x4d\x00\x1d\xd7\x1b\x4c\x06\x45\x7c\xe7\x26\x8e\xd4\x1c\x3e\xc1\x13\x89\xab\x39\x84\x67\xd3\x5a\x0a\x9c\xf6\xb9\x72\x08\x9f\x9d\x2b\x6e\x85\xa2\x5a\xc8\xfc\xa5\xe0\x1a\xef\xf5\x31\x79\xc7\x40\x0d\xe5\x19\x4b\x61\x24\xc9\x89\x98\xbe\xee\xe7\x4d\x18\xbb\x5d\x55\x5c\x7a\x35\xea\x88\x5d\xa5\xbf\xb6\x9c\x85\xac\x4b\xa2\xf0\x9f\x92\x85\x55\x12\xeb\x88\x6c\xfe\x0a\xb0\xfa\xd6\x40\x96\x34\x7f\x8d\x3c\xf6\x77\xb2\xdd\x52\xbe\x3e\x10\x15\x00\x79\x96\xce\xe1\x2e\xcc\x24\xfb\x07\xd1\x9b\xf0\x12\x42\xb5\x21\xb3\x17\x7f\x8c\x12\xba\x46\xa5\xc3\x45\xd0\xa2\x73\x2c\x65\x6b\xe3\x35\x55\x5a\xe6\x86\xfa\xed\xcb\x57\xe5\x72\x61\xce\x80\xc4\x31\x2a\x75\x60\x65\x34\x96\xb1\x50\xb0\x12\xb2\x40\x45\x05\x13\xb3\xc2\x7b\x8d\xdc\xd4\x08\x75\xbe\xc7\x59\x02\x80\x35\xd5\x9b\x6c\x79\x33\xce\x7b\xd4\xdb\xec\xd2\xb8\x40\xed\x40\xed\xce\xea\x41\xde\xd8\x36\x9b\x13\xb0\x34\x7f\xc1\x68\x0f\xba\xf1\xd2\x71\x88\x58\xa4\x29\xd5\x63\x31\xc1\x05\xc7\x53\xec\x72\xa2\xde\x3f\x0b\x8e\xce\x31\x94\xc8\x64\x8c\x3f\x95\x01\x77\x84\x38\xa6\xbf\x28\x17\x45\xe7\x50\xae\x0d\x85\x72\xe1\x5c\xe8\xe1\x6d\x0a\x1c\x91\xec\x0a\x14\xbc\xd7\x92\xbc\x2a\x00\xe6\x47\xd2\x09\x87\xda\xa1\x81\x0c\x55\xab\x99\xe1\xe1\xc7\x61\x9b\x3f\xd5\x01\x22\x52\x92\xbc\xd2\x9c\x6a\x4c\x1b\x79\xab\x57\x06\x4b\xcb\x23\xd5\x83\xdd\xae\x79\x7e\xbb\x6a\x26\xc9\x5e\x22\x0e\x2f\xdc\x0f\x58\x8f\xeb\x03\xc0\xcd\x20\xe0\x81\x03\x00\x97\xf3\xde\x6c\x31\x3e\xc2\xd9\x36\x44\x6d\x6e\xd8\x5a\x48\xaa\x37\x69\xe5\x82\x42\xa6\x84\x51\x45\x0c\xa3\xee\x6b\xdb\xde\x0e\xb8\x5d\x83\x60\xfb\x10\xdc\x41\x79\x07\xed\x65\x32\x8a\xe2\xfa\xea\x7e\x88\xa0\xd6\x3a\x1f\x69\x04\x32\xa2\xa1\x59\xa5\x98\x50\xf2\x8b\x8f\xbc\xae\xce\xe4\x64\xe1\xdd\x56\xc9\xa7\x82\x6a\x56\x90\x5f\x36\xe8\x80\x5c\x19\x11\x2b\xdb\x7c\x96\x6a\x43\x6d\x6e\x19\xb5\xcf\x43\xb3\x91\x73\xb1\x72\x59\xd2\x3b\x22\x05\x35\x14\x76\xf4\xf6\xe4\x81\xca\xaf\x47\x46\xa4\x5e\xcc\x86\x3f\xd8\x18\x51\x32\x7e\xed\xcb\xcc\xde\x7a\x4d\x4c\x49\x42\x89\x3c\x46\x3b\x38\xc0\xa4\x9a\xce\x99\x88\x09\x3b\x2f\xd2\xfc\x50\xed\xf0\x09\xf0\x0d\x32\x8c\xb5\x90\x0f\xcd\x97\x5f\x20\xa3\xd5\x47\xc8\xd7\x5e\xcb\x87\xda\xa5\xa4\x74\xe8\x7c\xdd\xe8\xfa\xea\x73\xf7\xf8\xfc\xdf\x33\xf4\x0e\xea\xd9\xcb\x62\xac\x26\xc2\x19\x90\x58\x67\x84\xb1\x7c\x5e\x71\x8a\x6c\xa0\x7d\x98\x82\xda\x62\x4c\x09\x03\x89\x06\x3e\xb6\x4c\x7e\xbf\x65\xf4\x8b\xd5\xc8\x76\x44\x0b\x8e\xed\x1a\x59\xf0\xe2\x19\x63\x07\x14\xb9\x7a\xf4\xdb\x69\xcb\x85\x5c\x95\x25\x8f\x6c\xbb\x3d\x01\x75\xf0\x5d\x50\xe1\x93\x70\x66\xf1\x6d\xe0\x57\x54\x2e\x8b\x9b\x80\x4c\x69\x48\x89\x8e\x37\xb5\x60\x50\x9d\xee\xad\xdb\x81\x33\x5b\xfd\x6a\x5b\xf5\x66\xe1\x7b\x53\x57\x6a\xe5\x12\xf7\x23\x79\xac\x23\x56\xcd\x1d\xee\x10\x0e\xee\xf2\xad\x0b\x98\x71\xd0\x0c\x6d\x92\x13\x56\x0e\x3a\xbf\xdf\xd6\x53\xc4\xf4\x4f\x4c\x1c\xde\x7b\x5a\x1b\xfc\x99\x32\x54\xb9\xd2\x98\x1e\x8f\x7b\xdb\xc7\xf0\x4b\x67\x0f\x11\xd3\x57\x29\x59\x9f\x34\x22\xda\x25\x35\x54\xca\xba\xf9\x28\xb3\x63\xfd\xaa\xc1\xfb\x53\x93\xcd\x9e\xcb\xa0\xca\x9c\x27\x28\xc6\x48\xee\xe3\xf2\x34\x7d\x20\x2c\x44\x0a\xa1\xba\x06\x58\x0d\x35\xb6\x37\x46\x81\x66\x5b\x61\x3a\xdb\x94\x70\xba\x42\xa5\xdb\x2d\x6d\x8b\xe9\x03\xfb\x66\x67\x19\x97\xc0\x5d\xa0\x38\x09\x14\x68\xb1\x87\x63\xdb\x51\xbb\xec\x1c\x84\x67\xa5\x89\x5c\xa3\xc6\x04\x62\xc1\x75\xd9\x28\x0d\x92\x57\xf4\xe3\xa8\x2e\xe6\x3d\x50\x0e\xcb\x5c\xa3\xf2\x3c\x96\xc6\xd8\x6d\xba\x3c\x4b\x97\xe6\x40\x03\x80\xc1\x90\x3d\xc1\x5d\x56\x94\x61\x55\x2f\x4f\xf5\x98\x1e\x09\x2b\xef\xf1\xac\x86\xec\xe2\xdf\xd7\xcd\x01\x7a\x43\x34\x50\x65\x75\x37\xe6\xa7\xdc\xbe\xfb\xc1\xbc\x54\x3f\x40\x42\xa5\x6d\xcc\xf3\xc1\xf3\xf0\x76\xbb\x7d\xa4\xf8\xfa\x02\x06\xbb\x6d\xc7\xd9\xb8\x73\x36\x1d\xd3\xc6\x3b\x7c\xa0\x7a\x53\x98\x26\xce\xa4\x44\xae\xa1\xef\x33\xde\x98\x95\x7c\x6a\x7d\x5d\x74\x46\xa7\x7c\x7d\xab\x4f\x01\x7d\x46\xfc\xde\x23\xed\xaf\x25\xf6\x30\xbe\x7e\x63\x32\xd4\x5c\xd4\xca\xee\xd7\x29\xf6\x01\x40\x75\x3f\x76\x42\xc0\x66\xfe\x82\xfc\xc4\xf2\x6e\x84\x29\x8f\x23\x1b\xb9\x0c\x0f\x00\xd6\xc8\x51\xd2\xf8\x1b\x5e\x64\x17\x12\xb8\xbb\xec\x62\xf1\x3d\xb2\x7f\x03\x91\x5d\x1d\x8c\xdb\xff\xb6\x81\xdd\x70\xd4\xaf\xd5\xc4\x97\x95\xe9\xe0\xeb\xaa\xa3\xef\xa7\xba\x7e\xda\xf9\x5c\xaa\x6a\x2f\xb7\x52\xec\x68\x52\x9d\x68\x04\x61\xe3\x92\xa1\x79\xe7\x55\xf6\xf3\xaa\x41\xbf\x81\xb1\xcf\xf7\x0f\xbf\xf2\x3a\xc1\x31\xbb\x3a\x1f\xed\x67\x9d\x39\x75\x6c\x00\xed\x7c\xcd\x0e\xe1\xcc\xf7\x23\x2c\xbf\x84\x0f\x08\x82\xb3\xbc\xf8\x05\x87\x6d\xdb\x05\xf7\x97\xd3\xfe\x0c\xbe\xd5\xc7\xa1\xe2\xf8\x1e\xe9\x82\xa2\xf5\xf5\xd0\xe3\xf7\xf8\xd0\xe3\x30\xec\x12\xae\x9c\xe0\xa1\x9a\x1d\x7e\xf6\xf5\x4b\xbd\xf0\x40\x67\x69\x34\x9b\x07\x21\xb5\xca\x98\xcd\x25\xfd\x26\x85\x4f\x9f\x83\x20\x68\x25\x96\x7a\xd6\x88\x20\x4c\xd1\xfd\x9a\xad\x1e\xd9\x61\xd0\x8c\xdb\xea\x57\x73\xbd\x02\x79\x12\xad\x84\x36\x7e\x40\x61\xfd\x3b\x4e\xb3\x39\xa8\x1d\x48\xe3\x30\xc6\xbf\x8d\x84\xff\x0f\x00\x00\xff\xff\xe8\xa1\xfb\x44\x98\x28\x00\x00")

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

	info := bindataFileInfo{name: "../../../../../../../resources/component-descriptor-v2-schema.yaml", size: 10392, mode: os.FileMode(436), modTime: time.Unix(1665847130, 0)}
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
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
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
	"..": &bintree{nil, map[string]*bintree{
		"..": &bintree{nil, map[string]*bintree{
			"..": &bintree{nil, map[string]*bintree{
				"..": &bintree{nil, map[string]*bintree{
					"..": &bintree{nil, map[string]*bintree{
						"..": &bintree{nil, map[string]*bintree{
							"..": &bintree{nil, map[string]*bintree{
								"resources": &bintree{nil, map[string]*bintree{
									"component-descriptor-v2-schema.yaml": &bintree{ResourcesComponentDescriptorV2SchemaYaml, map[string]*bintree{}},
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
