package comparch

import (
	"reflect"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/api/utils/accessobj"
)

type StateHandler struct{}

var _ accessobj.StateHandler = &StateHandler{}

func NewStateHandler(fs vfs.FileSystem) accessobj.StateHandler {
	return &StateHandler{}
}

func (i StateHandler) Initial() interface{} {
	return compdesc.New("", "")
}

func (i StateHandler) Encode(d interface{}) ([]byte, error) {
	return compdesc.Encode(d.(*compdesc.ComponentDescriptor))
}

func (i StateHandler) Decode(data []byte) (interface{}, error) {
	return compdesc.Decode(data)
}

func (i StateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
