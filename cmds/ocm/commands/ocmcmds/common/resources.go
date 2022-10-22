// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	_ "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/clisupport"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ResourceInput struct {
	Access *runtime.UnstructuredTypedObject `json:"access"`
	// Input  *inputs.BlobInput                `json:"input,omitempty"`
	Input *inputs.GenericInputSpec `json:"input,omitempty"`
}

type ResourceSpecHandler interface {
	RequireInputs() bool
	Decode(data []byte) (ResourceSpec, error)
	Set(v ocm.ComponentVersionAccess, r Resource, acc compdesc.AccessSpec) error
}

type ResourceSpec interface {
	GetName() string
	Info() string
	Validate(ctx clictx.Context, input *ResourceInput) error
}

type Resource interface {
	Source() string
	Spec() ResourceSpec
	Input() *ResourceInput
}

type resource struct {
	path   string
	source string
	spec   ResourceSpec
	input  *ResourceInput
}

func (r *resource) Source() string {
	return r.source
}

func (r *resource) Spec() ResourceSpec {
	return r.spec
}

func (r *resource) Input() *ResourceInput {
	return r.input
}

func NewResource(spec ResourceSpec, input *ResourceInput, path string, indices ...int) *resource {
	id := path
	for _, i := range indices {
		id += fmt.Sprintf("[%d]", i)
	}
	return &resource{
		path:   path,
		source: id,
		spec:   spec,
		input:  input,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpecifications interface {
	Origin() string
	Get() (string, error)
}

type ResourceSpecificationsFile struct {
	filesystem vfs.FileSystem
	path       string
}

func NewResourceSpecificationsFile(path string, fss ...vfs.FileSystem) ResourceSpecifications {
	return &ResourceSpecificationsFile{
		filesystem: accessio.FileSystem(fss...),
		path:       path,
	}
}

func (r *ResourceSpecificationsFile) Get() (string, error) {
	data, err := vfs.ReadFile(r.filesystem, r.path)
	if err != nil {
		return "", errors.Wrapf(err, "cannot read resource file %q", r.path)
	}
	return string(data), nil
}

func (r *ResourceSpecificationsFile) Origin() string {
	return r.path
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpecificationsProvider interface {
	AddFlags(fs *pflag.FlagSet)
	Complete() error
	Resources() ([]ResourceSpecifications, error)
	Description() string
	IsSpecified() bool
}

////////////////////////////////////////////////////////////////////////////////

type ResourceMetaDataSpecificationsProvider struct {
	typename string
	meta     string
	name     string
	version  string
}

func NewResourceMetaDataSpecificationsProvider(name string) ResourceMetaDataSpecificationsProvider {
	return ResourceMetaDataSpecificationsProvider{typename: name}
}

func (a *ResourceMetaDataSpecificationsProvider) ElementType() string {
	return a.typename
}

func (a *ResourceMetaDataSpecificationsProvider) IsSpecified() bool {
	return a.meta != "" || a.name != "" || a.version != ""
}

func (a *ResourceMetaDataSpecificationsProvider) Description() string {
	return fmt.Sprintf(`
It is possible to describe a single %s via command line options.
The meta data of this element is described by the argument of option <code>--%s</code>,
which must be a YAML or JSON string.
Alternatively, the <em>name</em> and <em>version</em> can be specified with the
options <code>--name</code> and <code>--version</code>. Explicitly specified options
override values specified by the <code>--%s</code> option.
(Note: Go templates are not supported for YAML-based option values. Besides
this restriction, the finally composed element description is still processd
by the selected templater.) 
`, a.typename, a.typename, a.typename)
}

func (a *ResourceMetaDataSpecificationsProvider) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&a.meta, a.typename, "", "", fmt.Sprintf("%s meta data (yaml)", a.typename))
	fs.StringVarP(&a.name, "name", "", "", fmt.Sprintf("%s name", a.typename))
	fs.StringVarP(&a.version, "version", "", "", fmt.Sprintf("%s version", a.typename))
}

func (a *ResourceMetaDataSpecificationsProvider) Complete() error {
	if !a.IsSpecified() {
		return nil
	}
	if a.meta != "" {
		if err := a.CheckData("meta data", a.meta); err != nil {
			return err
		}
	}
	return nil
}

func (a *ResourceMetaDataSpecificationsProvider) Origin() string {
	return a.typename + " (by options)"
}

func (a *ResourceMetaDataSpecificationsProvider) CheckData(n string, v string) error {
	if v == "" {
		return nil
	}
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(v), &data); err != nil {
		return errors.Wrapf(err, "%s %s is no valid yaml", a.typename, n)
	}
	return nil
}

func (a *ResourceMetaDataSpecificationsProvider) ParsedMeta() (map[string]interface{}, error) {
	data := map[string]interface{}{}
	if a.IsSpecified() {
		if a.meta != "" {
			err := yaml.Unmarshal([]byte(a.meta), &data)
			if err != nil {
				return nil, err
			}
		}
		if a.name != "" {
			data["name"] = a.name
		}
		if a.version != "" {
			data["version"] = a.version
		}
	}
	return data, nil
}

////////////////////////////////////////////////////////////////////////////////

type ContentResourceSpecificationsProvider struct {
	ResourceMetaDataSpecificationsProvider
	ctx           clictx.Context
	DefaultType   string
	rtype         string
	access        string
	inputOptions  clisupport.ConfigOptions
	accessOptions clisupport.ConfigOptions
}

var _ ResourceSpecificationsProvider = (*ContentResourceSpecificationsProvider)(nil)
var _ ResourceSpecifications = (*ContentResourceSpecificationsProvider)(nil)

func NewContentResourceSpecificationProvider(ctx clictx.Context, name string, deftype ...string) ResourceSpecificationsProvider {
	def := ""
	if len(deftype) > 0 {
		def = deftype[0]
	}
	return &ContentResourceSpecificationsProvider{
		ResourceMetaDataSpecificationsProvider: NewResourceMetaDataSpecificationsProvider(name),
		DefaultType:                            def,
		ctx:                                    ctx,
	}
}

func (a *ContentResourceSpecificationsProvider) Description() string {
	return a.ResourceMetaDataSpecificationsProvider.Description() + fmt.Sprintf(`
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
	a.ResourceMetaDataSpecificationsProvider.AddFlags(fs)
	fs.StringVarP(&a.rtype, "type", "", "", fmt.Sprintf("%s type", a.typename))
	fs.StringVarP(&a.access, "access", "", "", "access specification")

	a.inputOptions = inputs.For(a.ctx).CreateOptions()
	a.inputOptions.AddFlags(fs)
}

func (a *ContentResourceSpecificationsProvider) IsSpecified() bool {
	return a.ResourceMetaDataSpecificationsProvider.IsSpecified() || a.rtype != "" || a.inputOptions.Changed() || a.access != ""
}

func (a *ContentResourceSpecificationsProvider) Complete() error {
	if !a.IsSpecified() {
		return nil
	}
	if err := a.ResourceMetaDataSpecificationsProvider.Complete(); err != nil {
		return err
	}
	if a.access != "" && a.inputOptions.Changed() {
		return fmt.Errorf("either --input or --access is possible")
	}
	if a.access == "" && !a.inputOptions.Changed() {
		return fmt.Errorf("either --input, --inputType or --access is required")
	}

	if err := a.CheckData("access", a.access); err != nil {
		return err
	}
	return nil
}

func (a *ContentResourceSpecificationsProvider) Resources() ([]ResourceSpecifications, error) {
	if !a.IsSpecified() {
		return nil, nil
	}
	return []ResourceSpecifications{a}, nil
}

func (a *ContentResourceSpecificationsProvider) Get() (string, error) {
	data, err := a.ParsedMeta()
	if err != nil {
		return "", err
	}

	if a.rtype != "" {
		data["type"] = a.rtype
	}

	if data["type"] == nil && a.DefaultType != "" {
		data["type"] = a.DefaultType
	}

	if a.access != "" {
		var access map[string]interface{}
		yaml.Unmarshal([]byte(a.access), &access)
		data["access"] = access
	}

	in, err := inputs.For(a.ctx).GetConfigFor(a.inputOptions)
	if err != nil {
		return "", errors.Wrapf(err, "input specification")
	}
	if in != nil {
		data["input"] = in
	}

	r, err := json.Marshal(data)
	return string(r), nil
}

////////////////////////////////////////////////////////////////////////////////

type ResourceAdderCommand struct {
	utils.BaseCommand

	Templating template.Options
	Adder      ResourceSpecificationsProvider

	Archive   string
	Resources []ResourceSpecifications
	Envs      []string
}

func (o *ResourceAdderCommand) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	o.Templating.AddFlags(fs)
	if o.Adder != nil {
		o.Adder.AddFlags(fs)
	}
}

func (o *ResourceAdderCommand) Complete(args []string) error {
	o.Archive = args[0]
	o.Templating.Complete(o.Context.FileSystem())

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

	err := o.Templating.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	paths := o.Templating.FilterSettings(args[1:]...)
	for _, p := range paths {
		o.Resources = append(o.Resources, NewResourceSpecificationsFile(p, o.FileSystem()))
	}

	if len(o.Resources) == 0 {
		return fmt.Errorf("no specifications given")
	}
	return nil
}

func (o *ResourceAdderCommand) ProcessResourceDescriptions(listkey string, h ResourceSpecHandler) error {
	fs := o.Context.FileSystem()
	printer := common.NewPrinter(o.Context.StdOut())
	ictx := inputs.NewContext(o.Context, printer)

	resources := []*resource{}
	for _, source := range o.Resources {
		tmp, err := determineResources(printer, o.Context, ictx, o.Templating, listkey, h, source)
		if err != nil {
			return errors.Wrapf(err, "%s", source.Origin())
		}
		resources = append(resources, tmp...)
	}

	printer.Printf("found %d %s\n", len(resources), listkey)

	obj, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_WRITABLE, o.Archive, 0, accessio.PathFileSystem(fs))
	if err != nil {
		return err
	}
	defer obj.Close()

	for _, r := range resources {
		ictx := ictx.Section("adding %s...", r.Spec().Info())
		if h.RequireInputs() {
			if r.input.Input != nil {
				var acc ocm.AccessSpec
				// Local Blob
				blob, hint, berr := r.input.Input.GetBlob(ictx, common.VersionedElementKey(obj), r.path)
				if berr != nil {
					return errors.Wrapf(berr, "cannot get resource blob for %q(%s)", r.spec.GetName(), r.source)
				}
				acc, err = obj.AddBlob(blob, hint, nil)
				if err == nil {
					err = h.Set(obj, r, acc)
				}
				blob.Close()
			} else {
				err = h.Set(obj, r, compdesc.GenericAccessSpec(r.input.Access))
			}
		} else {
			err = h.Set(obj, r, nil)
		}
		if err != nil {
			return errors.Wrapf(err, "cannot add resource %q(%s)", r.spec.GetName(), r.source)
		}
	}
	return nil
}

func determineResources(printer common.Printer, ctx clictx.Context, ictx inputs.Context, templ template.Options, listkey string, h ResourceSpecHandler, source ResourceSpecifications) ([]*resource, error) {
	resources := []*resource{}
	origin := source.Origin()

	printer.Printf("processing %s...\n", origin)
	r, err := source.Get()
	if err != nil {
		return nil, err
	}
	parsed, err := templ.Execute(string(r))
	if err != nil {
		return nil, errors.Wrapf(err, "error during variable substitution")
	}

	// sigs parser has no multi document stream parsing
	// but yaml.v3 does not recognize json tagged fields.
	// Therefore, we first use the v3 parser to parse the multi doc,
	// marshal it again and finally unmarshal it with the sigs parser.
	decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(parsed)))
	i := 0
	for {
		var tmp map[string]interface{}

		i++
		err := decoder.Decode(&tmp)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
			break
		}
		printer.Printf("  processing document %d...\n", i)
		if (tmp["input"] != nil || tmp["access"] != nil) && !h.RequireInputs() {
			return nil, errors.Newf("invalid spec %d: no input or access possible for %s", i, listkey)
		}

		var list []json.RawMessage
		if reslist, ok := tmp[listkey]; ok {
			if len(tmp) != 1 {
				return nil, errors.Newf("invalid spec %d: either a list or a single spec possible", i)
			}
			l, ok := reslist.([]interface{})
			if !ok {
				return nil, errors.Newf("invalid spec %d: invalid resource list", i)
			}
			for j, e := range l {
				// cannot use json here, because yaml generates a map[interface{}]interface{}
				data, err := yaml.Marshal(e)
				if err != nil {
					return nil, errors.Newf("invalid spec %d[%d]: %s", i, j+1, err.Error())
				}
				list = append(list, data)
			}
		} else {
			if len(tmp) == 0 {
				return nil, errors.Newf("invalid spec %d: empty", i)
			}
			data, err := yaml.Marshal(tmp)
			if err != nil {
				return nil, err
			}
			list = append(list, data)
		}

		for j, d := range list {
			printer.Printf("    processing index %d\n", j+1)
			var input *ResourceInput
			r, err := DecodeResource(d, h)
			if err != nil {
				return nil, errors.Newf("invalid spec %d[%d]: %s", i, j+1, err)
			}

			if h.RequireInputs() {
				input, err = DecodeInput(d, ctx)
				if err != nil {
					return nil, errors.Newf("invalid spec %d[%d]: %s", i, j+1, err)
				}
				if err = Validate(input, ictx, origin); err != nil {
					return nil, errors.Wrapf(err, "invalid spec %d[%d]", i, j+1)
				}
			}

			if err = r.Validate(ctx, input); err != nil {
				return nil, errors.Wrapf(err, "invalid spec %d[%d]", i, j+1)
			}

			resources = append(resources, NewResource(r, input, origin, i, j+1))
		}
	}
	return resources, nil
}

func DecodeResource(data []byte, h ResourceSpecHandler) (ResourceSpec, error) {
	result, err := h.Decode(data)
	if err != nil {
		return nil, err
	}
	accepted, err := runtime.DefaultJSONEncoding.Marshal(result)
	if err != nil {
		return nil, err
	}
	var plainAccepted interface{}
	err = runtime.DefaultJSONEncoding.Unmarshal(accepted, &plainAccepted)
	if err != nil {
		return nil, err
	}
	var plainOrig map[string]interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &plainOrig)
	if err != nil {
		return nil, err
	}
	delete(plainOrig, "input")
	err = utils.CheckForUnknown(nil, plainOrig, plainAccepted).ToAggregate()
	return result, err
}

func DecodeInput(data []byte, ctx clictx.Context) (*ResourceInput, error) {
	var input ResourceInput
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}
	_, err = input.Input.Evaluate(inputs.For(ctx))
	if err != nil {
		return nil, err
	}
	accepted, err := runtime.DefaultJSONEncoding.Marshal(input.Input)
	if err != nil {
		return nil, err
	}
	var plainAccepted interface{}
	err = runtime.DefaultJSONEncoding.Unmarshal(accepted, &plainAccepted)
	if err != nil {
		return nil, err
	}
	var plainOrig map[string]interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &plainOrig)
	if err != nil {
		return nil, err
	}
	var fldPath *field.Path
	err = utils.CheckForUnknown(fldPath.Child("input"), plainOrig["input"], plainAccepted).ToAggregate()
	return &input, err
}

func Validate(r *ResourceInput, ctx inputs.Context, inputFilePath string) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if r.Input != nil && r.Access != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath, "only either input or access might be specified"))
	} else {
		if r.Input == nil && r.Access == nil {
			allErrs = append(allErrs, field.Forbidden(fldPath, "either input or access must be specified"))
		}
		if r.Access != nil {
			if r.Access.GetType() == "" {
				allErrs = append(allErrs, field.Required(fldPath.Child("access", "type"), "type of access required"))
			} else {
				acc, err := r.Access.Evaluate(ctx.OCMContext().AccessMethods())
				if err != nil {
					if errors.IsErrUnknown(err) {
						//nolint: errorlint // No way I can untagle this.
						err.(errors.Kinded).SetKind(errors.KIND_ACCESSMETHOD)
					}
					raw, _ := r.Access.GetRaw()
					allErrs = append(allErrs, field.Invalid(fldPath.Child("access"), string(raw), err.Error()))
				} else if acc.(ocm.AccessSpec).IsLocal(ctx.OCMContext()) {
					kind := runtime.ObjectVersionedType(r.Access.ObjectType).GetKind()
					allErrs = append(allErrs, field.Invalid(fldPath.Child("access", "type"), kind, "local access no possible"))
				}
			}
		}
		if r.Input != nil {
			if err := r.Input.Validate(fldPath.Child("input"), ctx, inputFilePath); err != nil {
				allErrs = append(allErrs, err...)
			}
		}
	}
	return allErrs.ToAggregate()
}
