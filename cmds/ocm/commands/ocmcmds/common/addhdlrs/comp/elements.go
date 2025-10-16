package comp

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	. "github.com/mandelsoft/goutils/finalizer"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/compatattr"
	"ocm.software/ocm/api/utils/errkind"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/refs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/rscs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/srcs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

const (
	ComponentVersionTag = common.ComponentVersionTag
)

type ResourceSpecHandler struct {
	rschandler *rscs.ResourceSpecHandler
	srchandler *srcs.ResourceSpecHandler
	refhandler *refs.ResourceSpecHandler
	version    string
	schema     *schemaoption.Option
}

var (
	_ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)
	_ options.Options            = (*ResourceSpecHandler)(nil)
)

func New(opts ...ocm.ModificationOption) *ResourceSpecHandler {
	return &ResourceSpecHandler{
		rschandler: rscs.New(opts...),
		srchandler: srcs.New(),
		refhandler: refs.New(),
		schema:     schemaoption.New(compdesc.DefaultSchemeVersion),
	}
}

func (h *ResourceSpecHandler) AsOptionSet() options.OptionSet {
	return options.OptionSet{h.rschandler.AsOptionSet(), h.srchandler.AsOptionSet(), h.refhandler.AsOptionSet(), h.schema}
}

func (h *ResourceSpecHandler) AddFlags(fs *pflag.FlagSet) {
	h.rschandler.AddFlags(fs)
	h.srchandler.AddFlags(fs)
	h.refhandler.AddFlags(fs)
	fs.StringVarP(&h.version, "version", "v", "", "default version for components")
	h.schema.AddFlags(fs)
}

func (h *ResourceSpecHandler) WithCLIOptions(opts ...options.Options) *ResourceSpecHandler {
	h.rschandler.WithCLIOptions(opts...)
	h.srchandler.WithCLIOptions(opts...)
	h.refhandler.WithCLIOptions(opts...)
	return h
}

func (*ResourceSpecHandler) Key() string {
	return "component"
}

func (*ResourceSpecHandler) RequireInputs() bool {
	return false
}

func (h *ResourceSpecHandler) Decode(data []byte) (addhdlrs.ElementSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	if desc.Version == "" {
		desc.Version = h.version
	}
	return &desc, nil
}

func (*ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r addhdlrs.Element, acc compdesc.AccessSpec) error {
	return fmt.Errorf("not supported for components")
}

func (h *ResourceSpecHandler) Add(ctx clictx.Context, ictx inputs.Context, elem addhdlrs.Element, repo ocm.Repository) (err error) {
	var final Finalizer
	defer final.FinalizeWithErrorPropagation(&err)

	r, ok := elem.Spec().(*ResourceSpec)
	if !ok {
		return fmt.Errorf("element spec is not a valid resource spec, failed to assert type %T to ResourceSpec", elem.Spec())
	}
	comp, err := repo.LookupComponent(r.Name)
	if err != nil {
		return errors.ErrNotFound(errkind.KIND_COMPONENT, r.Name)
	}
	final.Close(comp)

	cv, err := comp.NewVersion(r.Version, true)
	if err != nil {
		return errors.Wrapf(err, "%s:%s", r.Name, r.Version)
	}
	final.Close(cv)

	cd := cv.GetDescriptor()

	opts := h.srchandler.AsOptionSet()[0].(*addhdlrs.Options)
	if !opts.Replace {
		cd.Resources = nil
		cd.Sources = nil
		cd.References = nil
	}

	schema := h.schema.Schema
	if r.Meta.ConfiguredVersion != "" {
		schema = r.Meta.ConfiguredVersion
	}
	if schema != "" {
		if compdesc.DefaultSchemes[schema] == nil {
			return errors.ErrUnknown(errkind.KIND_SCHEMAVERSION, schema)
		}
		cd.Metadata.ConfiguredVersion = schema
	}

	cd.Labels = r.Labels
	cd.Provider = r.Provider
	if !compatattr.Get(ctx) {
		cd.CreationTime = metav1.NewTimestampP()
	}

	err = handle(ctx, ictx, elem.Source(), cv, r.Sources, h.srchandler)
	if err != nil {
		return err
	}
	err = handle(ctx, ictx, elem.Source(), cv, r.Resources, h.rschandler)
	if err != nil {
		return err
	}

	if len(r.References) > 0 && len(r.OldReferences) > 0 {
		return fmt.Errorf("only field references or componentReferences (deprecated) is possible")
	}
	err = handle(ctx, ictx, elem.Source(), cv, r.References, h.refhandler)
	if err != nil {
		return err
	}
	err = handle(ctx, ictx, elem.Source(), cv, r.OldReferences, h.refhandler)
	if err != nil {
		return err
	}
	return comp.AddVersion(cv)
}

func handle[T addhdlrs.ElementSpec](ctx clictx.Context, ictx inputs.Context, si addhdlrs.SourceInfo, cv ocm.ComponentVersionAccess, specs []T, h common.ResourceSpecHandler) error {
	key := utils.Plural(h.Key(), 0)
	elems, err := addhdlrs.MapSpecsToElems(ctx, ictx, si.Sub(key), specs, h)
	if err != nil {
		return errors.Wrapf(err, key)
	}
	return common.ProcessElements(ictx, cv, elems, h)
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	// Meta enabled to specify information for the serialization
	Meta compdesc.Metadata `json:"meta"`

	metav1.ObjectMeta `json:",inline"`
	// Sources defines sources that produced the component
	Sources []*srcs.ResourceSpec `json:"sources"`
	// References references component dependencies that can be resolved in the current context.
	References []*refs.ResourceSpec `json:"references"`
	// OldReferences references component dependencies that can be resolved in the current context.
	// Deprecated: use field References.
	OldReferences []*refs.ResourceSpec `json:"componentReferences"`
	// Resources defines all resources that are created by the component and by a third party.
	Resources []*rscs.ResourceSpec `json:"resources"`
}

var _ addhdlrs.ElementSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) GetRawIdentity() metav1.Identity {
	return metav1.NewIdentity(r.Name, metav1.SystemIdentityVersion, r.Version)
}

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("component %s:%s", r.Name, r.Version)
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *addhdlrs.ResourceInput) error {
	cd := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: r.ObjectMeta,
		},
	}
	return compdesc.Validate(cd)
}
