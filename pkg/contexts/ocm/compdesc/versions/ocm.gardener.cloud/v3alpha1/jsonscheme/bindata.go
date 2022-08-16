// Code generated by go-bindata. (@generated) DO NOT EDIT.

// Package jsonscheme generated by go-bindata.
// sources:
// ../../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml
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

var _ResourcesComponentDescriptorOcmV3SchemaYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x1a\xef\x6f\xdb\xb8\xf5\xbb\xfe\x8a\x87\x4b\x01\x39\x4d\x64\x37\x29\x3a\xe0\xfc\x25\xc8\x7a\x18\x50\x6c\x77\x39\xb4\xb7\x7d\x58\xea\x15\xb4\xf4\x6c\xb3\x47\x91\x1e\x49\xb9\x71\x7b\xfd\xdf\x07\x92\xa2\x44\xc9\x92\x7f\x26\xdd\x86\xbb\x7c\x89\x49\xbd\x5f\x7c\x7c\xbf\xa5\x67\x34\x1b\x43\xbc\xd0\x7a\xa9\xc6\xa3\xd1\x9c\xc8\x0c\x39\xca\x61\xca\x44\x91\x8d\x54\xba\xc0\x9c\xa8\x51\x2a\xf2\xa5\xe0\xc8\x75\x92\xa1\x4a\x25\x5d\x6a\x21\x13\x91\xe6\xc9\xea\x25\x61\xcb\x05\xb9\x8a\xa3\x67\x0e\x36\xa0\xf5\x51\x09\x9e\xb8\xdd\xa1\x90\xf3\x51\x26\xc9\x4c\x8f\xae\x5f\x5c\xbf\x48\xae\xae\x4b\xd2\x71\xe4\x09\x52\xc1\xc7\x10\xdf\xbd\xfe\x11\x5e\x7b\x66\xf0\x43\xc5\x0c\x56\x2f\xa1\xc6\x98\x51\x4e\x0d\x82\x1a\x47\x00\x39\x6a\x62\xfe\x03\xe8\xf5\x12\xc7\x10\x8b\xe9\x47\x4c\x75\x6c\xb7\x9a\xd4\xab\x63\xc0\x0a\xa5\xa2\x82\x5b\xe4\x8c\x68\xe2\xa0\x25\xfe\xbb\xa0\x12\x33\x47\x0e\x20\x81\x98\x93\x1c\xe3\x7a\x59\xe2\xb9\x1d\x92\x65\x56\x0c\xc2\x7e\x96\x62\x89\x52\x53\x54\x63\x98\x11\xa6\xd0\x3e\x5f\xd6\xbb\x25\x05\x43\xcd\xff\x06\x78\x26\x71\x36\x86\xf8\x6c\x14\x9c\xa8\x56\xf5\x4f\x01\xe7\x92\xed\x0e\x54\x89\x8c\x3c\x60\xf6\x0e\xf3\x15\x4a\x8f\xca\xc8\x14\x99\xda\x81\xe9\x80\x3c\xca\x52\x8a\x15\xcd\x50\xee\x40\xf2\x60\x71\x14\x35\xd9\x94\xf7\x40\xa4\x24\x6b\x47\x93\x6a\xcc\x2b\x19\xfa\x25\x88\x3d\xa1\xde\xfb\xdc\xe3\x86\x08\x2b\xca\xf5\x2e\xfd\x3b\xfa\x4a\x4b\xca\xe7\x5e\xd1\x06\x7b\x0c\x5f\xbe\xf6\x29\x7e\x49\xb4\x46\x69\x8c\xe9\x5f\xab\xfb\x17\xc9\xf7\x93\x8b\x67\x9e\xb9\xa2\x73\x4e\x74\x21\x37\x38\xc4\x53\x21\x18\x92\x3d\xac\x26\x02\x68\xdc\x7f\x43\x0f\x4e\x50\x47\x24\x27\x0f\x7f\x43\x3e\xd7\x8b\x31\x5c\xbf\x7a\x15\xb5\x24\xbb\x27\xc9\xe7\xc9\x7d\x42\x92\xcf\x46\xc2\xe7\x83\xfb\xe1\xa4\xb5\x75\xfe\xdc\xef\x7d\xb9\xbe\xfc\x3a\x18\x35\x1e\x7f\xe8\x40\xf9\x60\x70\xce\xcd\x61\x23\x00\x9a\x21\xd7\x54\xaf\x6f\xb5\x96\x74\x5a\x68\xfc\x2b\xae\x9d\xa8\x39\xe5\x95\x5c\x5d\x52\x19\xe6\x83\xfb\xe4\xc3\x85\x17\xc4\x6f\x9e\xdf\x38\xd2\x0d\x1b\x76\x34\xcf\x40\x93\x5f\x91\xc3\x4c\x8a\x1c\x94\x7d\x60\xe2\x09\x10\x9e\x01\xc9\x3e\x16\x4a\x63\x06\x5a\x00\x61\x4c\x7c\x02\xc2\x41\x2c\x9d\x7e\x81\x21\xc9\x28\x9f\x43\xbc\x8a\x2f\x21\x27\x1f\x4d\xd0\xe2\x6c\x7d\x69\x51\xed\x7a\x98\x53\x5e\xee\x7a\x5e\x0b\xaa\x20\x47\xc2\x15\xe8\x05\xc2\x4c\x18\xaa\x86\x88\x53\xbf\x02\x22\xd1\xb0\x32\xa6\x42\xb3\xa6\xbc\xca\x0b\x7c\x35\xbc\x1e\xbe\x0c\x7f\x27\x33\x21\x2e\xa6\x44\x96\x7b\xab\x10\x60\xd5\x05\x71\x35\xbc\xf6\xbf\x2a\xb0\x00\xbe\xfa\xd9\x40\x0b\x95\xbd\x9a\xdc\x0c\x5e\xfc\x76\x7f\x95\x7c\x3f\x79\x9f\x3d\x3f\x1f\xdc\x8c\xdf\x0f\xc3\x8d\xf3\x9b\xee\xad\x64\x30\xb8\x19\xd7\x9b\xbf\xbd\xcf\xec\x1d\xdd\x26\xff\x4c\x26\xc6\xe0\xfd\x6f\x4f\x72\x4f\xe0\x73\xcf\xf1\x62\x10\x3e\xb8\xb0\x44\x1a\x3b\x16\xb2\x74\xaa\x96\xe5\x77\x99\x5e\x6f\xa8\x28\xbd\x7f\x6d\xfc\x48\x8d\xe1\x4b\x77\xdc\xe9\x32\xe5\x18\xbe\x3a\x53\x5c\x0a\x45\xb5\x90\xeb\xd7\x82\x6b\x7c\xd0\x87\x44\x25\x03\xd5\x17\x85\x2c\x85\x76\x8c\x08\xce\x28\x52\xfa\xb6\x9b\x37\x61\xec\x6e\x56\x73\xe9\xc9\x02\x2d\xd4\x3a\x38\xb6\xe5\x2c\x65\x9d\x12\x85\x7f\x97\x2c\xae\x83\xdc\x86\xc8\xe6\xaf\x04\x0b\xb7\x3a\x63\x93\xfb\x6b\xc4\xb1\x1f\xc9\x72\x49\xf9\x7c\x4f\x54\x00\xe4\x45\x3e\x86\xfb\xb8\x90\xec\x67\xa2\x17\xf1\x25\xc4\x6a\x41\xae\x5f\xfd\x29\xc9\xe8\x1c\x95\x8e\x27\x51\x8b\xce\xa1\x94\xad\x8e\xe7\x54\x69\xb9\x36\xd4\xef\x5e\xbf\xa9\x96\x13\x73\x07\x24\x4d\x51\xa9\x3d\xeb\x0a\xa3\x19\x0b\x05\x33\x21\x4b\x54\x54\x30\x30\x2b\x7c\xd0\xc8\x4d\x0e\x51\xe7\x3b\x8c\x25\x02\x98\x53\xbd\x28\xa6\xb7\xdb\x79\x6f\xb5\x36\xbb\x34\x26\x10\x5c\xa8\xdd\x99\x1d\x65\x8d\x6d\xb5\x39\x01\x2b\xf5\x97\x8c\x76\xa0\x1b\x2b\xdd\x0e\x91\x8a\x3c\xa7\x7a\x9b\x4f\x70\xc1\xf1\x14\xbd\x9c\x78\xee\x9f\x04\x47\x67\x18\x4a\x14\x32\xc5\x1f\x2a\x87\x3b\x40\x1c\x53\x7d\x54\x8b\xb2\xb2\xa8\xd6\x86\x42\xb5\x70\x26\x74\x40\x11\xb3\x21\xf8\xfe\xc1\xae\x44\xc1\x07\x2d\xc9\x9b\x12\x60\x47\xe5\xb7\x41\xe7\x11\xea\xd4\x43\xcd\xb0\xb2\xc1\x23\x0a\xdc\xd0\xb9\xed\x9a\xaf\xef\x66\xcd\xa0\xd8\x49\xc5\xe1\xc5\xbb\x01\x43\x3f\xde\x03\xdc\x74\x4c\x1e\x38\x02\x70\x31\xee\xdd\x12\xd3\x03\x8c\x6b\x41\xd4\xe2\x96\xcd\x85\xa4\x7a\x91\xd7\x26\x27\x64\x4e\x18\x55\xc4\x30\xda\x7c\x6c\xcb\xdd\x23\x7b\x99\x06\xc3\xad\x45\x75\xb7\x10\x7b\xd4\xe1\xdd\x10\x51\x50\x6a\x1f\xa8\x24\xb2\x45\x03\x66\x95\x63\x46\xc9\x2f\xde\x13\x0f\xd7\x09\x39\xf9\x70\x6e\xab\x92\xa3\x86\x6a\x66\x9c\x5f\x16\xe8\x80\x5c\xda\x11\x33\x5b\xac\x56\x6a\x81\xa0\x0b\xda\xaa\xbf\x63\xa3\x97\x33\xd1\x6a\x59\xd1\x3b\x52\x6f\x3b\xfb\x32\xc7\x6f\x87\x93\xd7\x7e\xb3\xa5\x25\xeb\xc4\x6c\xd8\x93\xf5\x41\x25\xd3\xb7\x3e\x6d\xed\xcc\xff\xc4\xa4\x38\x94\xc8\x53\xb4\x8d\x08\x0c\xea\x81\x09\x13\x29\x61\xe7\x65\xda\xe8\xcb\x45\x3e\xa0\xbe\x43\x86\xa9\x16\xbb\x3a\xef\xde\xf8\x7b\x50\x2c\xb4\x25\x6e\x29\xf6\xb1\x07\xad\xce\xb9\x6f\x7b\xde\x39\xde\x38\x7d\xb0\xd2\xd1\x35\xf7\x9e\xbf\x53\x84\x6d\x49\x15\xce\x80\xa4\xba\x20\x8c\xad\xc7\x35\xa7\xc4\x7a\xde\xa7\x11\xa8\x25\xa6\x94\x30\x90\x68\xe0\x53\xcb\xe4\xff\x37\x0f\x1f\x91\x4e\xdb\xce\x29\x38\xb6\xd3\x69\xa9\x50\x5e\x30\xb6\x47\x3e\x0c\x1d\xd9\x5a\xa9\xf3\x9e\x3a\x20\x1e\x58\x91\x7b\x02\xea\xd0\x31\x1f\x9c\x59\x7c\xeb\xc3\x35\x95\xcb\x72\x48\x50\x28\x0d\x39\xd1\xe9\x22\x70\x03\xb5\x51\xd8\x6d\x16\xe7\xcc\x26\xc2\x60\x2b\xac\x2b\xfe\xa8\xf7\xaa\x53\xb9\x18\xac\x36\xa0\x82\xc1\x22\xb4\x87\x8b\xbd\x42\x38\x62\x75\x4b\xe2\x2e\x61\xef\x8a\xd3\x9a\x80\xe9\x14\x4d\x3f\x27\x39\x61\xff\xd3\xf5\xa7\x48\xe9\x9f\x99\xd8\xbf\x00\xb5\xa7\xfb\x0b\x65\xa8\xd6\x4a\x63\x7e\x38\xee\x5d\x17\xc3\xa7\x8e\x0b\x22\xa5\x6f\x72\x32\x3f\xa9\x2f\xb4\x4b\x6a\xa8\xbc\xf5\x99\xed\x51\x1a\xc6\x70\xbe\xe0\x2d\xa5\xc9\x66\xc7\x04\xa8\x56\xe7\x09\x07\x63\x64\xed\x3d\xee\xb4\xf3\x40\x5c\x8a\x14\x43\xdd\xfb\xcf\xfa\xaa\xd3\x5b\x73\x80\x66\xa9\x60\xca\xd3\x9c\x70\x3a\x43\xa5\xdb\x75\x69\x8b\xe9\x91\xc5\xaf\xd3\x8c\x0b\xcd\xce\x51\x9c\x04\x0a\xb4\xd8\xc1\xb1\x6d\xa8\x9b\xec\x1c\x84\x67\xa5\x89\x9c\xa3\xc6\x0c\x52\xc1\x75\x55\xfc\xf4\x92\x57\xf4\xf3\xd6\xb3\x98\xe7\x40\x39\x4c\xd7\x1a\x95\xe7\x31\x35\xca\x6e\xd3\xe5\x45\x3e\xf5\x2f\x5c\xfa\x5c\xf6\x04\x73\x99\x51\x86\x75\x26\x3c\xd5\x62\x3a\x24\xac\xad\xc7\xb3\xea\xd3\x8b\x7f\x1e\xaa\x03\xf4\x82\x68\xa0\xca\x9e\xdd\xa8\x9f\x72\xfb\xec\x3b\xf3\x50\x7d\x07\x19\x95\xb6\x7a\x5e\xf7\xde\x87\xd7\xdb\xdd\x23\xf9\xd7\x13\x28\xec\xae\xed\x67\xdb\x8d\xb3\x69\x98\xd6\xdf\xe1\x13\xd5\x8b\x52\x35\x69\x21\x25\x72\x5d\x17\x28\x50\xbf\xc0\xdd\xa6\x25\x1f\x5a\xdf\x96\x35\xcf\x29\x2f\xe4\xc2\xca\xbe\x4b\x89\x7f\x54\x3f\xbb\x73\x89\xbd\x8c\xc7\x2c\x39\xfa\xca\x86\x20\xa1\x7e\x9b\x34\x1e\x01\xd4\xe3\xaf\x13\x5c\xb1\xf0\xf3\xee\x13\x13\xb7\x11\xa6\x52\x74\xb1\x65\xb6\x1d\x01\xcc\x91\xa3\xa4\xe9\x7f\x71\x2e\x5d\x4a\xe0\x46\xd3\xe5\xe2\x5b\xfb\xec\xe3\x8c\x7b\x7e\x67\x3e\x5d\x5f\x9c\xdb\x7f\x2a\x97\x6e\x98\xe8\xb7\x2a\xcc\x9b\x1f\x90\x1c\x6a\x81\x4f\x62\x4f\x87\x4e\xc6\xd4\xb6\xc1\x72\x33\x05\xdb\xf9\xcf\x8c\xa6\xb6\xa1\xf4\x99\xb8\xac\x0c\xcd\x32\x98\x92\x79\xf3\xd2\xc7\x9e\xb4\x9c\x40\x3c\x52\x4b\xdc\x7a\x95\x15\xbc\xaf\x73\x85\xfb\x23\xf1\x91\xcd\xce\xaa\x1e\xe8\x1c\x4e\x7f\xa3\x53\xde\xf2\x1a\xbc\x1e\x1a\xc5\xfb\x20\xb4\x4b\x9e\xbd\x90\x5a\x21\x37\x8e\xa2\x96\xb9\x84\x96\x6e\xe2\xe6\x92\xfe\xa3\x8e\xad\x09\xc4\xbf\x52\x9e\x95\x3f\xc3\x6f\xd1\x12\x67\x56\x71\xd4\x34\x81\x1a\xbd\x61\x9b\xa1\xa9\x07\x0d\x5b\x3e\x6c\x7d\xce\x57\x7d\xad\x67\x8b\x4b\xc3\xba\x97\x4c\x2a\xb8\xd2\x63\x88\xab\x8f\xf1\x02\xb1\xbd\xa0\x0e\xb9\x53\x2f\x06\x24\xee\xfa\x86\x62\xbf\x4f\xc4\x5a\xd7\xdc\x7f\x63\x1b\xdf\x49\xc4\x70\xe6\x8b\x5e\xb6\xbe\x84\x4f\x08\x82\xb3\x75\xf9\x6d\x90\xed\x0d\x05\xc7\x86\x7f\x77\xbb\x46\xf9\x12\xa1\x7a\x31\x70\xc2\xa7\x6d\x15\x8d\xf8\x3f\x01\x00\x00\xff\xff\x1e\x1d\x2e\xd1\x6c\x29\x00\x00")

func ResourcesComponentDescriptorOcmV3SchemaYamlBytes() ([]byte, error) {
	return bindataRead(
		_ResourcesComponentDescriptorOcmV3SchemaYaml,
		"../../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml",
	)
}

func ResourcesComponentDescriptorOcmV3SchemaYaml() (*asset, error) {
	bytes, err := ResourcesComponentDescriptorOcmV3SchemaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml", size: 10604, mode: os.FileMode(436), modTime: time.Unix(1659967843, 0)}
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
	"../../../../../../../../resources/component-descriptor-ocm-v3-schema.yaml": ResourcesComponentDescriptorOcmV3SchemaYaml,
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
								"..": &bintree{nil, map[string]*bintree{
									"resources": &bintree{nil, map[string]*bintree{
										"component-descriptor-ocm-v3-schema.yaml": &bintree{ResourcesComponentDescriptorOcmV3SchemaYaml, map[string]*bintree{}},
									}},
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
