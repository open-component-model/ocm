package rscs

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	compdescv2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/skipdigestoption"
	"ocm.software/ocm/cmds/ocm/common/options"
)

const (
	ComponentVersionTag = common.ComponentVersionTag
)

type ResourceSpecHandler struct {
	addhdlrs.ResourceSpecHandlerBase
	opts *ocm.ModificationOptions
}

var (
	_ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)
	_ options.Options            = (*ResourceSpecHandler)(nil)
)

func New(opts ...ocm.ModificationOption) *ResourceSpecHandler {
	h := &ResourceSpecHandler{ResourceSpecHandlerBase: addhdlrs.NewBase(options.OptionSet{skipdigestoption.New()})}
	if len(opts) > 0 {
		h.opts = ocm.NewModificationOptions(opts...)
	}
	return h
}

func (h *ResourceSpecHandler) WithCLIOptions(opts ...options.Options) *ResourceSpecHandler {
	return &ResourceSpecHandler{
		h.ResourceSpecHandlerBase.WithCLIOptions(opts...),
		h.opts,
	}
}

func (h *ResourceSpecHandler) getModOpts() []ocm.ModificationOption {
	opts := options.FindOptions[ocm.ModificationOption](h.AsOptionSet())
	if h.opts != nil {
		opts = append(opts, h.opts)
	}
	return opts
}

func (*ResourceSpecHandler) Key() string {
	return "resource"
}

func (*ResourceSpecHandler) RequireInputs() bool {
	return true
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
	if spec.Relation == metav1.LocalRelation {
		if vers == "" || vers == ComponentVersionTag {
			vers = v.GetVersion()
		}
	}
	if vers == ComponentVersionTag {
		vers = v.GetVersion()
	}
	if vers == "" {
		return errors.Newf("resource %q (%s): missing version", spec.Name, r.Source())
	}

	meta := &compdesc.ResourceMeta{
		ElementMeta: compdesc.ElementMeta{
			Name:          spec.Name,
			Version:       vers,
			ExtraIdentity: spec.ExtraIdentity,
			Labels:        spec.Labels,
		},
		Type:       spec.Type,
		Relation:   spec.Relation,
		SourceRefs: compdescv2.ConvertSourcerefsTo(spec.SourceRefs),
	}
	opts := h.getModOpts()
	if spec.Options.SkipDigestGeneration {
		opts = append(opts, ocm.SkipDigest()) //nolint:staticcheck // skip digest still used for tests
	}
	/*
		if ocm.IsIntermediate(v.Repository().GetSpecification()) {
			opts = append(opts, ocm.ModifyElement())
		}
	*/
	return v.SetResource(meta, acc, opts...)
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	compdescv2.ElementMeta `json:",inline"`

	// Type describes the type of the object.
	Type string `json:"type"`

	// Relation describes the relation of the resource to the component.
	// Can be a local or external resource
	Relation metav1.ResourceRelation `json:"relation,omitempty"`

	// SourceRefs defines a list of source names.
	// These entries reference the sources defined in the
	// component.sources.
	SourceRefs []compdescv2.SourceRef `json:"srcRefs"`

	addhdlrs.ResourceInput `json:",inline"`

	// Options describes additional process related options
	// see ResourceOptions for more details.
	Options ResourceOptions `json:"options,omitempty"`
}

// ResourceOptions describes additional process related options
// which reflect the handling of the resource without describing it directly.
// Typical examples are any options that require specific changes in handling of the resource
// but are not reflected in the resource itself (outside of side effects)
type ResourceOptions struct {
	// SkipDigestGeneration omits the digest generation.
	SkipDigestGeneration bool `json:"skipDigestGeneration,omitempty"`
}

var _ addhdlrs.ElementSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) GetType() string {
	return r.Type
}

func (r *ResourceSpec) GetRawIdentity() metav1.Identity {
	return r.ElementMeta.GetRawIdentity()
}

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("resource %s: %s", r.Type, r.GetRawIdentity())
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *addhdlrs.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if r.Relation == "" {
		if input.Input != nil {
			r.Relation = metav1.LocalRelation
		}
		if r.Access != nil {
			r.Relation = metav1.ExternalRelation
		}
	}
	if r.Version == "" && r.Relation == metav1.LocalRelation {
		r.Version = ComponentVersionTag
	}
	rsc := compdescv2.Resource{
		ElementMeta: r.ElementMeta,
		Type:        r.Type,
		Relation:    r.Relation,
		SourceRefs:  r.SourceRefs,
	}
	if err := compdescv2.ValidateResource(fldPath, rsc, false); err != nil {
		allErrs = append(allErrs, err...)
	}
	return allErrs.ToAggregate()
}
