package refs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ResourceSpecHandler struct {
	addhdlrs.ResourceSpecHandlerBase
}

var (
	_ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)
	_ options.Options            = (*ResourceSpecHandler)(nil)
)

func New(opts ...options.Options) *ResourceSpecHandler {
	return &ResourceSpecHandler{addhdlrs.NewBase(opts...)}
}

func (*ResourceSpecHandler) Key() string {
	return "reference"
}

func (h *ResourceSpecHandler) WithCLIOptions(opts ...options.Options) *ResourceSpecHandler {
	return &ResourceSpecHandler{
		h.ResourceSpecHandlerBase.WithCLIOptions(opts...),
	}
}

func (*ResourceSpecHandler) Decode(data []byte) (addhdlrs.ElementSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func (h *ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r addhdlrs.Element, acc compdesc.AccessSpec) error {
	spec, ok := r.Spec().(*ResourceSpec)
	if !ok {
		return fmt.Errorf("element spec is not a valid resource spec, failed to assert type %T to ResourceSpec", r.Spec())
	}
	vers := spec.Version
	if vers == "" {
		vers = v.GetVersion()
	}
	meta := &compdesc.ComponentReference{
		ElementMeta: compdesc.ElementMeta{
			Name:          spec.Name,
			Version:       vers,
			ExtraIdentity: spec.ExtraIdentity,
			Labels:        spec.Labels,
		},
		ComponentName: spec.ComponentName,
	}
	return v.SetReference(meta, h.GetTargetOpts()...)
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	compdescv2.ElementMeta `json:",inline"`
	// ComponentName describes the remote name of the referenced object
	ComponentName string `json:"componentName"`
}

var _ addhdlrs.ElementSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) GetRawIdentity() metav1.Identity {
	return r.ElementMeta.GetRawIdentity()
}

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("reference %s: %s", r.ComponentName, r.GetRawIdentity())
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *addhdlrs.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	ref := compdescv2.ComponentReference{
		ElementMeta:   r.ElementMeta,
		ComponentName: r.ComponentName,
	}
	if err := compdescv2.ValidateComponentReference(fldPath, ref); err != nil {
		allErrs = append(allErrs, err...)
	}
	return allErrs.ToAggregate()
}
