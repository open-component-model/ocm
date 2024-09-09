package pluginhdlr

import (
	"slices"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

func Elem(e interface{}) *Object {
	return e.(*Object)
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	Signatures []string                      `json:"signatures,omitempty"`
	Element    *compdesc.ComponentDescriptor `json:"descriptor"`
}

func (o *Object) AsManifest() interface{} {
	return o
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx  clictx.OCM
	store signing.VerifiedStore
}

func NewTypeHandler(octx clictx.OCM, path string, fss ...vfs.FileSystem) (utils.TypeHandler, error) {
	fs := general.OptionalDefaulted(vfsattr.Get(octx.Context()), fss...)
	store, err := signing.NewVerifiedStore(path, fs)
	if err != nil {
		return nil, err
	}
	return &TypeHandler{
		octx:  octx,
		store: store,
	}, nil
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	result := []output.Object{}

	for _, nv := range h.store.Entries() {
		e := h.store.GetEntry(nv)
		result = append(result, &Object{slices.Clone(e.Signatures), e.Descriptor.Descriptor().Copy()})
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	comp, err := ocm.ParseComp(elemspec.String())
	if err != nil {
		return nil, err
	}
	if comp.IsVersion() {
		e := h.store.GetEntry(comp.NameVersion())
		if e == nil {
			return nil, errors.ErrNotFound(ocm.KIND_COMPONENTVERSION, elemspec.String())
		}
		return []output.Object{&Object{slices.Clone(e.Signatures), e.Descriptor.Descriptor().Copy()}}, nil
	}

	var objs []output.Object
	for _, nv := range h.store.Entries() {
		if nv.GetName() == comp.Component {
			e := h.store.GetEntry(nv)
			objs = append(objs, &Object{slices.Clone(e.Signatures), e.Descriptor.Descriptor().Copy()})
		}
	}
	if len(objs) == 0 {
		return nil, errors.ErrNotFound(ocm.KIND_COMPONENT, elemspec.String())
	}
	return objs, nil
}
