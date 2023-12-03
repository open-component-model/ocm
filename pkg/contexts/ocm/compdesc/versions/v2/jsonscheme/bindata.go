// Code generated by go-bindata. (@generated) DO NOT EDIT.

//Package jsonscheme generated by go-bindata.// sources:
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

var _ResourcesComponentDescriptorV2SchemaYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x5a\x5f\x6f\x1b\xb9\x11\x7f\xd7\xa7\x18\x9c\x03\x50\x8e\xbd\x96\xa3\x22\x05\xa2\x17\xc3\xcd\xa1\x40\xd0\xde\xf9\x90\xa4\x7d\xa8\xe3\x06\xd4\xee\x48\x62\xca\x25\x55\x92\x52\xac\xcb\xf9\xbb\x1f\x48\x2e\xf7\xff\xae\xfe\xd9\xb9\x0b\x10\x3f\x24\x22\x77\x38\x33\x1c\xfe\x66\x38\x33\xbb\xcf\x58\x32\x01\xb2\x30\x66\xa9\x27\xa3\xd1\x9c\xaa\x04\x05\xaa\x8b\x98\xcb\x55\x32\xd2\xf1\x02\x53\xaa\x47\xb1\x4c\x97\x52\xa0\x30\x51\x82\x3a\x56\x6c\x69\xa4\x8a\xd6\x63\x32\x78\xe6\x29\x4a\x1c\x3e\x69\x29\x22\x3f\x7b\x21\xd5\x7c\x94\x28\x3a\x33\xa3\xf1\xe5\xf8\x32\x7a\x31\xce\x18\x92\x41\x60\xc3\xa4\x98\x00\xb9\x59\xa2\x80\xd7\x41\x06\xfc\x24\x13\xe4\xb0\x1e\x43\x41\x3d\x63\x82\x59\x62\x3d\x19\x00\xa4\x68\xa8\xfd\x1f\xc0\x6c\x96\x38\x01\x22\xa7\x9f\x30\x36\xc4\x4d\x55\x39\xe7\x8a\x43\xa1\xb8\x5b\x9f\x50\x43\xfd\x02\x85\xff\x5f\x31\x85\x89\xe7\x08\x10\x01\xf1\x72\xff\x8d\x4a\x33\x29\x3c\xd5\x52\xc9\x25\x2a\xc3\x50\x07\xba\x0a\x51\x98\xcc\x55\xd2\x46\x31\x31\x27\x03\xa7\xae\x9a\x63\xa7\xbe\x4d\xc6\x94\xcf\xa5\x62\x66\x91\x16\x4c\x97\xd4\x18\x54\x76\x43\xff\xbd\xa5\xd1\xaf\x77\xf6\x9f\xcb\xe8\xd5\xe8\x63\x74\x77\xf6\x8c\x64\x64\xb1\x14\x33\x36\x9f\xc0\x17\x78\x70\x33\x34\x49\x9c\xcd\x28\xff\xa5\x90\x01\x33\xca\x35\x0e\x00\x38\x9d\x22\xef\xd4\xaa\xc5\x28\x82\xa6\x48\x8a\xe1\x9a\xf2\x15\x76\x6d\xc1\xd2\x76\x9a\xc4\x4f\xba\xf5\x13\xf8\xf2\x10\xc6\x75\x43\x96\xf6\xbc\xbe\xbd\x8c\x5e\x95\x76\xaa\xd9\x5c\x30\x31\x6f\x48\x98\x4a\xc9\x91\x8a\x40\x56\x32\xbc\xfd\x7b\xa6\x70\x36\x01\x72\x32\x2a\xc1\x69\xe4\x68\xdc\x31\xe5\x50\xf9\x39\x57\xbe\x45\xf1\x94\xde\xff\x13\xc5\xdc\x2c\x26\x30\x7e\xf9\x72\xd0\x7a\x38\x91\x3f\x9d\xbb\xe7\xc3\xdb\x8b\xbb\xda\xd4\xe9\xf3\x30\xf7\x65\x7c\xfe\x30\x1c\x55\x1e\x7f\x6c\x59\xf2\xd1\xae\x39\xb5\x7b\x1f\x00\xb0\x04\x85\x61\x66\x73\x6d\x8c\x62\xd3\x95\xc1\x7f\xe0\xc6\xab\x9a\x32\x91\xeb\xd5\xa6\x95\x15\x3e\xbc\x8d\x3e\x9e\x05\x45\xc2\xe4\xe9\x95\x67\xad\x90\xd3\x7b\x4c\xde\x61\xba\x46\xe5\x79\x9e\x80\xa1\xff\x43\x01\x33\x25\x53\xd0\xee\x81\x75\x69\xa0\x22\x01\x9a\x7c\x5a\x69\x83\x09\x18\x09\x94\x73\xf9\x19\xa8\x00\xb9\xf4\x78\x03\x8e\x34\x61\x62\x0e\x64\x4d\xce\x21\xa5\x9f\xa4\x8a\xa4\xe0\x9b\x73\xb7\xd4\x8d\x2f\x52\x26\xb2\xd9\x20\x6b\xc1\x34\xa4\x48\x85\x06\xb3\x40\x98\x49\xcb\xd5\x32\xf1\xe6\xd7\x40\x15\x5a\x51\x16\x39\x2c\xa9\xea\xab\x83\xc2\x2f\x2e\xc6\x17\x7f\x29\xff\x8e\x66\x52\x9e\x4d\xa9\xca\xe6\xd6\x65\x82\x75\x1b\xc5\x8b\x8b\x71\xf8\x95\x93\x95\xe8\xf3\x9f\x95\x65\x65\x63\xaf\xef\xae\x86\x97\xbf\xdd\xbe\x88\x5e\xdd\x7d\x48\x9e\x9f\x0e\xaf\x26\x1f\x2e\xca\x13\xa7\x57\xed\x53\xd1\x70\x78\x35\x29\x26\x7f\xfb\x90\xb8\x33\xba\x8e\xfe\x13\xdd\x59\xfc\x87\xdf\x81\xe5\x8e\xc4\xa7\x41\xe2\xd9\xb0\xfc\xe0\xcc\x31\xa9\xcc\x38\xca\xcc\xc7\x9a\x51\xac\x01\xbd\x6d\x11\x6d\x63\xfd\x48\xdb\x70\xd4\xea\x78\x6d\x50\x26\xf0\xe0\xa1\xb8\x94\x9a\x19\xa9\x36\xaf\xa5\x30\x78\x6f\xf6\x09\x53\x96\xaa\x2b\x2c\x39\x0e\x3d\x91\x5a\xc6\xec\x6d\xbb\x6c\xca\xf9\xcd\xac\x90\xd2\xba\xa3\x86\xda\x45\xb4\xac\xeb\x99\xe9\x3a\xa5\x1a\xff\xa5\x38\x29\x62\x5e\x43\x65\xfb\x97\x91\x95\xa7\x3a\x82\xaa\xbf\x06\x4a\x71\xec\x27\xba\x5c\x56\x22\x65\xef\x52\x00\x14\xab\x74\x02\xb7\x64\xa5\xf8\x2f\xd4\x2c\xc8\x39\x10\xbd\xa0\xe3\x97\x7f\x8d\x12\x36\x47\x6d\xc8\xdd\xa0\xc6\x67\x5f\xce\xce\xc6\x73\xa6\x8d\xda\x58\xee\x37\xaf\xdf\xe4\xc3\x3b\x7b\x06\x34\x8e\x51\xeb\x1d\xaf\x77\x6b\x19\x47\x05\x33\xa9\xb2\xa5\xa8\x61\x68\x47\x78\x6f\x50\xd8\x2b\x45\x9f\x6e\x01\xcb\x00\x60\xce\xcc\x62\x35\xbd\xee\x97\xdd\x8b\x36\x37\xb4\x10\x28\x1d\xa8\x9b\x99\x1d\x84\xc6\xba\xd9\xbc\x82\xb9\xf9\x33\x41\x5b\x96\x5b\x94\xf6\x53\xc4\x32\x4d\x99\xe9\xf3\x09\x21\x05\x1e\x63\x97\x23\xf7\xfd\xb3\x14\xe8\x81\xa1\xe5\x4a\xc5\xf8\x63\xee\x70\x7b\xa8\x63\xd3\x91\x7c\x90\x25\x1a\xf9\xd8\x72\xc8\x07\x1e\x42\x87\x67\x35\x1d\x59\x46\x6b\xb0\xcb\x96\xe0\xbd\x51\xf4\x4d\x46\xb0\x25\x5b\x69\xf0\x21\x5d\xd9\x53\x47\x84\x2a\xdd\x99\x64\xf7\xe3\x70\xb9\xa2\x6e\x10\x51\xa5\xe8\xa6\xd8\x39\x33\x98\x56\xe2\x56\xab\x0e\x8e\x57\x58\x54\x76\x76\x37\x16\x9b\x9b\x59\x35\x48\xb6\x32\xf1\xeb\xc8\x76\xc2\xb2\x5f\xef\x40\x6e\x8b\x98\x40\x3c\x00\xf0\x31\xef\xdd\x12\xe3\x3d\xc0\xb6\xa0\x7a\x71\x1d\x52\xf8\x02\x82\x52\xa5\x94\x33\x4d\xad\xa0\xe6\x63\x97\x0d\x77\xc0\xae\xc2\xb0\x7e\x08\xfe\xa0\x02\x40\x5b\x85\xf4\x2e\xf1\x69\x78\x3b\xc5\xc0\x67\xda\xd4\xac\x14\xee\x69\x04\xda\xb3\x43\x3b\x4a\x31\x61\xf4\x7d\xf0\xbc\x9d\x6a\xa0\x3d\x95\xf7\x53\xb9\x9c\x82\xaa\x7a\x83\xbc\x5f\xa0\x27\xf2\xd7\x88\x9c\xb9\xe4\x33\xdf\x36\x94\xca\x9c\x5e\xfb\x1c\x1a\x8d\x3c\xc4\xf2\x61\xce\xcf\x67\x1d\xbd\x15\xdc\xae\x21\xaa\x62\x10\x2f\x6f\x4b\x9c\x28\x70\x5f\xae\xb8\x4a\xfb\xec\x5c\x59\xc1\x4b\x1e\x61\x58\x8a\xda\xd0\x74\xd9\x8d\x33\x81\xb6\x98\xf8\xf1\x10\x7f\xcb\xcd\x79\x80\x35\x1a\x61\xb3\x85\xe6\x51\xe2\xf3\xde\x66\xcf\x6d\x92\xb7\x45\xbc\x71\xf6\xb9\x85\x3b\xaf\xbd\x6d\x96\x6a\xd5\xae\x92\x57\x3e\xc2\xbd\x73\x20\x12\x15\x66\x89\x40\xd9\x1c\x70\xe4\xa5\x54\x87\x9f\xb3\xbf\x56\xf1\xdb\x90\x40\x6d\xcd\x44\xa9\x4d\xb6\x50\xa1\x88\xd1\x95\xc4\x30\x2c\x7a\x66\x5c\xc6\x94\x9f\x66\x09\x4c\x57\x56\x14\xa0\xf3\x0e\x39\xc6\x46\xaa\x43\x91\xf6\x04\x77\x75\xb9\x39\xf2\x36\xec\xf2\x50\xbb\xe4\x9c\x76\x6d\x34\xb5\xe2\x2e\x02\xb2\xee\x6f\xcf\xb5\xb4\x73\xf6\x83\x76\x5f\xb6\x07\x27\x40\x63\xb3\xa2\x9c\x6f\x26\x85\xa4\xc8\x5d\x21\x9f\x47\xa0\x97\x18\x33\xca\x2d\x56\x8d\x62\xb1\x13\xf2\xed\x26\x88\x4f\x96\xfd\xd5\x23\x80\x14\x58\xcf\xfe\x32\x59\x62\xc5\xf9\x0e\xe9\x5b\x2d\x80\x86\x50\x51\xdc\xff\x7b\x16\x94\x81\x81\xde\xb9\x29\x9a\x61\x12\x4e\xdc\x7a\xe7\xf8\x05\x97\xf3\xac\xc7\xb5\xd2\x06\x52\x6a\xe2\x45\xc9\x19\x74\xa3\x2e\x69\xd6\x96\xdc\xe5\x75\xa5\xa9\x72\x1a\xfc\xbd\x5c\xc9\x77\xe5\x03\xf7\x23\x21\xd6\x33\x2b\x6e\x1f\x7f\x08\x3b\xd7\xaf\x0e\x02\xe4\x1c\x08\xde\x1b\x54\x82\xf2\xbc\x84\xff\x76\x8b\x2a\x19\xb3\xbf\x71\xb9\x7b\x55\xe5\x6c\xf0\x77\xc6\x51\x6f\xb4\xc1\x74\xff\xb5\x37\x6d\x02\x9f\x3a\x7a\xc8\x98\xbd\x49\xe9\xfc\xa8\xe6\x87\x1b\x32\xcb\x25\xbf\x37\x1f\xa5\x2b\x52\x6e\xa2\x05\x3c\x55\xc5\x6c\x69\x73\x16\xe6\x3c\x62\x63\x9c\x6e\x82\x5f\x1e\xb7\x1f\x20\x99\x4a\x04\x8a\x06\xd7\xac\xab\x64\xbb\xb6\x1b\xa8\xa6\x15\xb6\x66\x4b\xa9\x60\x33\xd4\xa6\x5e\xac\xd5\x84\x1e\x58\x11\x7a\xcb\xf8\x00\xee\x1d\xc5\x6b\xa0\xc1\xc8\x2d\x12\xeb\x40\x6d\x8a\xf3\x14\x41\x94\xa1\x6a\x8e\x06\x13\x88\xa5\x30\x79\xa2\xd4\xc9\x5e\xb3\x5f\x7b\xf7\x62\x9f\x03\x13\x30\xdd\x18\xd4\x41\xc6\xd4\x1a\xbb\xce\x57\xac\xd2\xa9\x3d\xd0\x01\x40\xa7\xcb\x1e\x01\x97\x19\xe3\x58\xdc\x97\xc7\x22\xa6\x45\xc3\x02\x3d\x41\x54\x97\x5d\xc2\xf3\xb2\x39\xc0\x2c\xa8\x01\xa6\xdd\xde\xad\xf9\x99\x70\xcf\x7e\xb0\x0f\xf5\x0f\x90\x30\xe5\x12\xf3\x4d\xe7\x79\x04\xbb\xdd\x3c\x92\x7f\x3d\x81\xc1\x6e\xea\x7e\xd6\x0f\xce\x2a\x30\x9d\xbf\xc3\x67\x66\x16\x99\x69\xe2\x95\x52\x28\x0c\xb4\xbd\x65\xef\xb3\x52\x08\xad\x6f\xb3\xcc\x68\x1f\x1b\x75\x64\x5c\x9d\x46\xfc\x9e\x23\x6d\xbf\x4b\xdc\x61\x7c\xfd\xc4\xa4\x2b\xb9\x28\x5d\xbb\x5f\xe7\xb2\x1f\x00\x14\x9d\xdf\x23\x1c\x76\x15\x5e\xfd\x1c\x79\xbd\x5b\x65\xf2\xe3\x58\xf5\xbc\xe6\x19\x00\xcc\x51\xa0\x62\xf1\x1f\xf8\x8a\x26\xd3\xc0\xbf\xa5\xc9\x06\xdf\x3d\xfb\x4f\xe0\xd9\xc5\xc1\xf8\xf9\x3f\xd6\xb1\x2b\x40\xfd\x5a\x49\x7c\x7e\x33\xed\xdc\xae\xda\xbb\x3f\xd5\xc4\x69\xe3\x43\x00\x5d\x7a\xb8\x54\x72\xcd\x92\xe2\x44\x23\x20\x95\x26\x43\xb5\xe7\x95\xe7\xf3\xba\xc2\xbf\xb2\xe2\x4f\xd1\xcd\x8d\x15\xba\xc2\xf8\x3d\x6b\xba\xdd\x6d\x40\xe8\x79\x76\x8c\xc5\x47\x04\x33\xa9\x52\x6a\x26\x90\x50\x83\x91\x61\x79\xbf\xba\x69\xc2\xbd\x61\xdb\x28\x7b\xfb\xea\xd9\xc6\x67\x1f\x04\x4e\x42\x7a\xc3\x37\xe7\xf0\x19\x41\x0a\xbe\xc9\x3e\x75\x72\x55\x80\x14\x41\xd9\x70\xa4\x5b\x1c\xf3\xc9\xdc\x2f\x43\xc3\x23\xf5\x3b\x6a\xaf\xd9\xf3\x03\x6e\x42\xf2\x71\x04\x36\x19\xd7\x5b\xfd\x4f\x79\xf6\xe5\x1e\x21\xd9\x11\x2c\x95\xdc\x75\xa7\x45\xb5\x5b\xd1\x85\xa6\x76\x93\xc2\x97\x87\xc1\x60\x50\x8b\x53\xe5\x20\x14\x01\x49\xd1\x7f\xbb\x5a\x0e\x14\x64\x50\x0d\x03\xc5\x37\xb2\x1d\x9f\x3d\x7a\x16\xb5\xf8\xd8\x7f\x40\xa4\xfc\xc2\xb3\x9a\x6b\x94\x0e\xa4\x72\x18\xfd\x2f\x09\x49\xed\xcd\xdf\x11\x3c\xdb\x5f\x96\x91\xdf\x03\x00\x00\xff\xff\x9e\xf8\x98\x8b\xde\x2c\x00\x00")

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

	info := bindataFileInfo{name: "../../../../../../../resources/component-descriptor-v2-schema.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
