package plugindirattr

import (
	"fmt"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/modern-go/reflect2"
	"ocm.software/ocm/api/datacontext"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ATTR_KEY   = "github.com/mandelsoft/ocm/plugindir"
	ATTR_SHORT = "plugindir"

	DEFAULT_PLUGIN_DIR = utils.DEFAULT_OCM_CONFIG_DIR + "/plugins"
)

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

func DefaultDir(fs vfs.FileSystem) string {
	home, _ := os.UserHomeDir() // use home if provided
	if home != "" {
		dir := filepath.Join(home, DEFAULT_PLUGIN_DIR)
		if ok, err := vfs.DirExists(fs, dir); ok && err == nil {
			return dir
		}
	}
	return ""
}

type AttributeType struct{}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*plugin directory*
Directory to look for OCM plugin executables.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(string); !ok {
		return nil, fmt.Errorf("directory path required")
	}
	return marshaller.Marshal(v)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value string
	err := unmarshaller.Unmarshal(data, &value)
	return value, err
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) string {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if reflect2.IsNil(a) {
		return DefaultDir(osfs.New())
	}
	return a.(string)
}

func Set(ctx datacontext.Context, path string) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, path)
}
