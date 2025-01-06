package featurehdlr

import (
	"sort"

	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/sliceutils"

	"ocm.software/ocm/api/datacontext/attrs/featuregatesattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/featuregates"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

func Elem(e interface{}) *Object {
	return e.(*Object)
}

////////////////////////////////////////////////////////////////////////////////

type Settings = featuregatesattr.FeatureGate

type Object struct {
	featuregates.FeatureGate `json:",inline"`
	Settings                 `json:",inline"`
}

func CompareObject(a, b output.Object) int {
	return Compare(a, b)
}

func (o *Object) AsManifest() interface{} {
	return o
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx ocm.Context
}

func NewTypeHandler(octx ocm.Context) utils.TypeHandler {
	h := &TypeHandler{
		octx: octx,
	}
	return h
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	result := []output.Object{}

	gates := featuregatesattr.Get(h.octx)
	list := sliceutils.AppendUnique(featuregates.DefaultRegistry().GetNames(), maputils.Keys(gates.Features)...)
	sort.Strings(list)

	for _, n := range list {
		var s *featuregatesattr.FeatureGate

		def := featuregates.DefaultRegistry().Get(n)
		if def != nil {
			s = def.GetSettings(h.octx)
		} else {
			def = &featuregates.FeatureGate{
				Name:        n,
				Short:       "<unknown>",
				Description: "",
				Enabled:     false,
			}
			s = gates.GetFeature(n)
		}

		o := &Object{
			FeatureGate: *def,
			Settings:    *s,
		}
		result = append(result, o)
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	def := featuregates.DefaultRegistry().Get(elemspec.String())

	if def == nil {
		def = &featuregates.FeatureGate{
			Name:        elemspec.String(),
			Short:       "<unknown>",
			Description: "",
			Enabled:     false,
		}
	}
	s := featuregatesattr.Get(h.octx).GetFeature(elemspec.String(), false)
	return []output.Object{
		&Object{
			FeatureGate: *def,
			Settings:    *s,
		},
	}, nil
}
