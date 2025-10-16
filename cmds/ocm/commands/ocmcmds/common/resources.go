package common

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	utils2 "ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	_ "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/dryrunoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/fileoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/templateroption"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

const ComponentVersionTag = "<componentversion>"

type ResourceSpecHandler interface {
	addhdlrs.ElementSpecHandler
	Set(v ocm.ComponentVersionAccess, r addhdlrs.Element, acc compdesc.AccessSpec) error
}

func CheckHint(v ocm.ComponentVersionAccess, elem addhdlrs.Element, acc compdesc.AccessSpec) error {
	err := checkHint(v, "source", elem, compdesc.SourceArtifacts, acc)
	if err != nil {
		return err
	}
	return checkHint(v, "resource", elem, compdesc.ResourceArtifacts, acc)
}

func checkHint(v ocm.ComponentVersionAccess, typ string, elem addhdlrs.Element, artacc compdesc.ArtifactAccess, acc compdesc.AccessSpec) error {
	spec, err := v.GetContext().AccessSpecForSpec(acc)
	if err != nil {
		return err
	}
	local, ok := spec.(*localblob.AccessSpec)
	if !ok {
		return nil
	}
	if local.ReferenceName == "" {
		return nil
	}
	elemid := elem.Spec().GetRawIdentity()
	if elemid[v1.SystemIdentityVersion] == ComponentVersionTag {
		elemid[v1.SystemIdentityVersion] = v.GetVersion()
	}
	accessor := artacc(v.GetDescriptor())
	for i := 0; i < accessor.Len(); i++ {
		a := accessor.GetArtifact(i)
		other, err := v.GetContext().AccessSpecForSpec(a.GetAccess())
		if err != nil {
			continue
		}
		if elemid.Equals(a.GetMeta().GetRawIdentity()) {
			continue
		}
		olocal, ok := other.(*localblob.AccessSpec)
		if !ok {
			continue
		}
		if olocal.ReferenceName != local.ReferenceName {
			continue
		}
		if mime.BaseType(local.MediaType) == mime.BaseType(olocal.MediaType) {
			return fmt.Errorf("reference name (hint) %q with base media type %s already used for %s %s:%s",
				local.ReferenceName, mime.BaseType(local.MediaType), typ, a.GetMeta().GetName(), a.GetMeta().GetVersion())
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type ElementFileSource struct {
	filesystem vfs.FileSystem
	path       addhdlrs.SourceInfo
}

func NewElementFileSource(path string, fss ...vfs.FileSystem) addhdlrs.ElementSource {
	return &ElementFileSource{
		filesystem: utils2.FileSystem(fss...),
		path:       addhdlrs.NewSourceInfo(path),
	}
}

func (r *ElementFileSource) Get() (string, error) {
	data, err := vfs.ReadFile(r.filesystem, r.path.Origin())
	if err != nil {
		return "", errors.Wrapf(err, "cannot read resource file %q", r.path)
	}
	return string(data), nil
}

func (r *ElementFileSource) Origin() addhdlrs.SourceInfo {
	return r.path
}

////////////////////////////////////////////////////////////////////////////////

type ElementSpecificationsProvider interface {
	AddFlags(fs *pflag.FlagSet)
	Complete() error
	Resources() ([]addhdlrs.ElementSource, error)
	Description() string
	IsSpecified() bool
}

////////////////////////////////////////////////////////////////////////////////

type ElementMetaDataSpecificationsProvider struct {
	typename     string
	metaProvider flagsets.ConfigTypeOptionSetConfigProvider
	metaOptions  flagsets.ConfigOptions
}

func NewElementMetaDataSpecificationsProvider(name string, adder flagsets.ConfigAdder, types ...flagsets.ConfigOptionType) *ElementMetaDataSpecificationsProvider {
	meta := flagsets.NewPlainConfigProvider(name, flagsets.ComposedAdder(addMeta(name), adder),
		sliceutils.CopyAppend(types,
			flagsets.NewYAMLOptionType(name, fmt.Sprintf("%s meta data (yaml)", name)),
			flagsets.NewStringOptionType("name", fmt.Sprintf("%s name", name)),
			flagsets.NewStringOptionType("version", fmt.Sprintf("%s version", name)),
			flagsets.NewStringMapOptionType("extra", fmt.Sprintf("%s extra identity", name)),
			flagsets.NewValueMapOptionType("label", fmt.Sprintf("%s label (leading * indicates signature relevant, optional version separated by @)", name)),
		)...,
	)
	meta.AddGroups(cases.Title(language.English).String(fmt.Sprintf("%s meta data options", name)))
	a := &ElementMetaDataSpecificationsProvider{
		typename:     name,
		metaProvider: meta,
	}
	a.metaOptions = a.metaProvider.CreateOptions()
	return a
}

func addMeta(typename string) flagsets.ConfigAdder {
	return func(opts flagsets.ConfigOptions, config flagsets.Config) error {
		if o, ok := opts.GetValue(typename); ok {
			for k, v := range o.(flagsets.Config) {
				config[k] = v
			}
		}

		flagsets.AddFieldByOption(opts, "name", config)
		flagsets.AddFieldByOption(opts, "version", config)
		flagsets.AddFieldByOption(opts, "extra", config, "extraIdentity")
		if err := flagsets.AddFieldByMappedOption(opts, "label", config, MapLabelSpecs, "labels"); err != nil {
			return err
		}
		return nil
	}
}

func (a *ElementMetaDataSpecificationsProvider) ElementType() string {
	return a.typename
}

func (a *ElementMetaDataSpecificationsProvider) IsSpecified() bool {
	return a.metaOptions.Changed()
}

func (a *ElementMetaDataSpecificationsProvider) Description() string {
	return fmt.Sprintf(`
It is possible to describe a single %s via command line options.
The meta data of this element is described by the argument of option <code>--%s</code>,
which must be a YAML or JSON string.
Alternatively, the <em>name</em> and <em>version</em> can be specified with the
options <code>--name</code> and <code>--version</code>. With the option <code>--extra</code>
it is possible to add extra identity attributes. Explicitly specified options
override values specified by the <code>--%s</code> option.
(Note: Go templates are not supported for YAML-based option values. Besides
this restriction, the finally composed element description is still processed
by the selected template engine.)
`, a.typename, a.typename, a.typename)
}

func (a *ElementMetaDataSpecificationsProvider) AddFlags(fs *pflag.FlagSet) {
	a.metaOptions.AddFlags(fs)
}

func (a *ElementMetaDataSpecificationsProvider) Complete() error {
	return nil
}

func (a *ElementMetaDataSpecificationsProvider) Origin() addhdlrs.SourceInfo {
	return addhdlrs.NewSourceInfo(a.typename + " (by options)")
}

func (a *ElementMetaDataSpecificationsProvider) ParsedMeta() (flagsets.Config, error) {
	return a.metaProvider.GetConfigFor(a.metaOptions)
}

////////////////////////////////////////////////////////////////////////////////

type ContentResourceSpecificationsProvider struct {
	*ElementMetaDataSpecificationsProvider
	ctx         clictx.Context
	DefaultType string

	accprov flagsets.ConfigTypeOptionSetConfigProvider
	shared  flagsets.ConfigOptionTypeSet
	options flagsets.ConfigOptions

	contentFlags []string
}

var (
	_ ElementSpecificationsProvider = (*ContentResourceSpecificationsProvider)(nil)
	_ addhdlrs.ElementSource        = (*ContentResourceSpecificationsProvider)(nil)
)

func NewContentResourceSpecificationProvider(ctx clictx.Context, name string, adder flagsets.ConfigAdder, deftype string, types ...flagsets.ConfigOptionType) *ContentResourceSpecificationsProvider {
	a := &ContentResourceSpecificationsProvider{
		DefaultType: deftype,
		ctx:         ctx,
		ElementMetaDataSpecificationsProvider: NewElementMetaDataSpecificationsProvider(name, flagsets.ComposedAdder(addContentMeta, adder),
			sliceutils.CopyAppend(types,
				flagsets.NewStringOptionType("type", fmt.Sprintf("%s type", name)),
			)...,
		),
	}
	return a
}

func addContentMeta(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOption(opts, "type", config)
	return nil
}

func (a *ContentResourceSpecificationsProvider) Description() string {
	return a.ElementMetaDataSpecificationsProvider.Description() + fmt.Sprintf(`
The %s type can be specified with the option <code>--type</code>. Therefore, the
minimal required meta data for elements can be completely specified by dedicated
options and don't need the YAML option.

To describe the content of this element one of the options <code>--access</code> or
<code>--input</code> must be given. They take a YAML or JSON value describing an
attribute set, also. The structure of those values is similar to the <code>access</code>
or <code>input</code> fields of the description file format.
`, a.typename)
}

func (a *ContentResourceSpecificationsProvider) AddFlags(fs *pflag.FlagSet) {
	a.ElementMetaDataSpecificationsProvider.AddFlags(fs)

	a.accprov = a.ctx.OCMContext().AccessMethods().CreateConfigTypeSetConfigProvider()
	inptypes := inputs.For(a.ctx).ConfigTypeSetConfigProvider()

	set := flagsets.NewConfigOptionTypeSet("resources")
	set.AddAll(a.accprov)
	dup, err := set.AddAll(inptypes)
	if err != nil {
		logging.Logger().LogError(err, "composing resources flags")
	}
	a.shared = dup
	a.options = set.CreateOptions()
	a.options.AddTypeSetGroupsToOptions(a.accprov)
	a.options.AddTypeSetGroupsToOptions(inptypes)
	a.options.AddFlags(fs)
	a.contentFlags = nil
	for _, t := range []flagsets.ConfigOptionType{inptypes.GetPlainOptionType(), inptypes.GetTypeOptionType(), a.accprov.GetPlainOptionType(), a.accprov.GetTypeOptionType()} {
		if t != nil {
			a.contentFlags = append(a.contentFlags, t.GetName())
		}
	}
}

func (a *ContentResourceSpecificationsProvider) IsSpecified() bool {
	return a.ElementMetaDataSpecificationsProvider.IsSpecified() || a.options.Changed()
}

func (a *ContentResourceSpecificationsProvider) Complete() error {
	if !a.IsSpecified() {
		return nil
	}
	if err := a.ElementMetaDataSpecificationsProvider.Complete(); err != nil {
		return err
	}

	unique := a.options.FilterBy(flagsets.Not(a.shared.HasOptionType))
	aopts := unique.FilterBy(a.accprov.HasOptionType)
	iopts := unique.FilterBy(inputs.For(a.ctx).ConfigTypeSetConfigProvider().HasOptionType)

	if !a.options.Changed(a.contentFlags...) {
		return fmt.Errorf("one of %v is required", flagsets.AddPrefix("--", a.contentFlags...))
	}
	if aopts.Changed() && iopts.Changed() {
		return fmt.Errorf("either input or access specification is possible")
	}
	return nil
}

func (a *ContentResourceSpecificationsProvider) apply(p flagsets.ConfigTypeOptionSetConfigProvider, data flagsets.Config) error {
	if p.IsExplicitlySelected(a.options) {
		ac, err := p.GetConfigFor(a.options)
		if err != nil {
			return errors.Wrapf(err, "%s specification", p.GetName())
		}
		if ac != nil {
			data[p.GetName()] = ac
		}
	}
	return nil
}

func (a *ContentResourceSpecificationsProvider) ParsedMeta() (flagsets.Config, error) {
	data, err := a.ElementMetaDataSpecificationsProvider.ParsedMeta()
	if err != nil {
		return nil, err
	}
	if data["type"] == nil && a.DefaultType != "" {
		data["type"] = a.DefaultType
	}

	if data["type"] == nil {
		return nil, fmt.Errorf("resource type is required")
	}
	return data, err
}

func (a *ContentResourceSpecificationsProvider) Get() (string, error) {
	data, err := a.ParsedMeta()
	if err != nil {
		return "", err
	}

	err = a.apply(a.accprov, data)
	if err != nil {
		return "", err
	}
	err = a.apply(inputs.For(a.ctx).ConfigTypeSetConfigProvider(), data)
	if err != nil {
		return "", err
	}

	//nolint:errchkjson // We don't care about this error.
	r, _ := json.Marshal(data)
	return string(r), nil
}

func (a *ContentResourceSpecificationsProvider) Resources() ([]addhdlrs.ElementSource, error) {
	if !a.IsSpecified() {
		return nil, nil
	}
	return []addhdlrs.ElementSource{a}, nil
}

////////////////////////////////////////////////////////////////////////////////

type ResourceAdderCommand struct {
	utils.BaseCommand

	Adder ElementSpecificationsProvider

	Resources []addhdlrs.ElementSource
	Envs      []string

	Archive string

	Handler ResourceSpecHandler
}

func NewResourceAdderCommand(ctx clictx.Context, h ResourceSpecHandler, provider ElementSpecificationsProvider, opts ...options.Options) ResourceAdderCommand {
	if o, ok := h.(options.Options); ok {
		opts = append(opts, o)
	}
	return ResourceAdderCommand{
		BaseCommand: utils.NewBaseCommand(ctx, sliceutils.CopyAppend[options.Options](opts,
			//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
			fileoption.NewCompArch(),
			dryrunoption.New(fmt.Sprintf("evaluate and print %s specifications", h.Key()), true),
			templateroption.New(""),
		)...),
		Adder:   provider,
		Handler: h,
	}
}

func (o *ResourceAdderCommand) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	if o.Adder != nil {
		o.Adder.AddFlags(fs)
	}
}

func (o *ResourceAdderCommand) Complete(args []string) error {
	err := o.OptionSet.ProcessOnOptions(options.CompleteOptionsWithCLIContext(o.Context))
	if err != nil {
		return err
	}

	o.Archive, args = fileoption.From(o).GetPath(args, o.Context.FileSystem())

	if o.Adder != nil {
		err := o.Adder.Complete()
		if err != nil {
			return err
		}

		rsc, err := o.Adder.Resources()
		if err != nil {
			return err
		}
		o.Resources = append(o.Resources, rsc...)
	}

	t := templateroption.From(o)
	err = t.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	paths := t.FilterSettings(args...)
	for _, p := range paths {
		o.Resources = append(o.Resources, NewElementFileSource(p, o.FileSystem()))
	}

	if len(o.Resources) == 0 {
		return fmt.Errorf("no specifications given")
	}
	return nil
}

func (o *ResourceAdderCommand) ProcessResourceDescriptions() error {
	fs := o.Context.FileSystem()
	printer := common.NewPrinter(o.Context.StdOut())
	elems, ictx, err := addhdlrs.ProcessDescriptions(o.Context, printer, templateroption.From(o).Options, o.Handler, o.Resources)
	if err != nil {
		return err
	}

	dr := dryrunoption.From(o)
	if dr.DryRun {
		return addhdlrs.PrintElements(printer, elems, dr.Outfile, o.Context.FileSystem())
	}

	// FIXME: use CommonTransportFormat archives to store OCM components
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	obj, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_WRITABLE, o.Archive, 0, accessio.PathFileSystem(fs))
	if err != nil {
		return err
	}
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	defer obj.Close()
	return ProcessElements(ictx, obj, elems, o.Handler)
}

func IsVersionSet(vers string) bool {
	return vers != "" && vers != ComponentVersionTag
}

// ProcessElements add a list of evaluated elements to a component version.
func ProcessElements(ictx inputs.Context, cv ocm.ComponentVersionAccess, elems []addhdlrs.Element, h ResourceSpecHandler) error {
	var err error
	for _, elem := range elems {
		ictx := ictx.Section("adding %s...", elem.Spec().Info())
		if h.RequireInputs() {
			if elem.Input().Input != nil {
				var acc ocm.AccessSpec
				// Local Blob
				info := inputs.InputResourceInfo{
					ComponentVersion: common.VersionedElementKey(cv),
					ElementName:      elem.Spec().GetName(),
					InputFilePath:    general.OptionalDefaulted(elem.Source().Origin(), elem.Input().SourceFile),
				}
				blob, hint, berr := elem.Input().Input.GetBlob(ictx, info)
				if berr != nil {
					return errors.Wrapf(berr, "cannot get %s blob for %q(%s)", h.Key(), elem.Spec().GetName(), elem.Source())
				}
				if iv := elem.Input().Input.GetInputVersion(ictx); iv != "" && !IsVersionSet(elem.Spec().GetVersion()) {
					elem.Spec().SetVersion(iv)
				}
				acc, err = cv.AddBlob(blob, elem.Type(), hint, nil)
				blob.Close()
				if err == nil {
					err = CheckHint(cv, elem, acc)
					if err == nil {
						err = h.Set(cv, elem, acc)
					}
				}
			} else {
				acc := elem.Input().Access
				err = CheckHint(cv, elem, acc)
				if err == nil {
					err = h.Set(cv, elem, acc)
				}
			}
		} else {
			err = h.Set(cv, elem, nil)
		}
		if err != nil {
			return errors.Wrapf(err, "cannot add %s %q(%s)", h.Key(), elem.Spec().GetName(), elem.Source())
		}
	}
	return nil
}
