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

type Resources interface {
	Origin() string
	Get(printer common.Printer, topts template.Options) (string, error)
}

type ResourcesFile struct {
	filesystem vfs.FileSystem
	path       string
}

func NewResourcesFile(path string, fss ...vfs.FileSystem) Resources {
	return &ResourcesFile{
		filesystem: accessio.FileSystem(fss...),
		path:       path,
	}
}

func (r *ResourcesFile) Get(printer common.Printer, topts template.Options) (string, error) {
	printer.Printf("processing %s...\n", r.path)
	data, err := vfs.ReadFile(r.filesystem, r.path)
	if err != nil {
		return "", errors.Wrapf(err, "cannot read resource file %q", r.path)
	}

	parsed, err := topts.Execute(string(data))
	if err != nil {
		return "", errors.Wrapf(err, "error during variable substitution for %q", r.path)
	}
	return parsed, nil
}

func (r *ResourcesFile) Origin() string {
	return r.path
}

////////////////////////////////////////////////////////////////////////////////

type AdderOptions interface {
	AddFlags(fs *pflag.FlagSet)
	Complete() error
	GetResources() ([]Resources, error)
	Description() string
}

type ResourceAdder struct {
	typename string
	resource string
	input    string
	access   string
}

var _ AdderOptions = (*ResourceAdder)(nil)
var _ Resources = (*ResourceAdder)(nil)

func NewResourceAdder(name string) AdderOptions {
	return &ResourceAdder{typename: name}
}

func (a *ResourceAdder) Description() string {
	return fmt.Sprintf(`
It is possible to describe a single %s via command line options.
This requires the option <code>--resource</code> and one of the options
<code>--access</code> or <code>--input</code>. All three options require
a yaml or json value describing an attribute set. This is similar
to the one supported for the specification via yaml file.
`, a.typename)
}

func (a *ResourceAdder) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&a.resource, a.typename, "", "", fmt.Sprintf("%s meta data (yaml)", a.typename))
	fs.StringVarP(&a.input, "input", "", "", "input specification")
	fs.StringVarP(&a.access, "access", "", "", "access specification")
}

func (a *ResourceAdder) check(n string, v string) error {
	if v == "" {
		return nil
	}
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(v), &data); err != nil {
		return errors.Wrapf(err, "%s %s is no valid yaml", a.typename, n)
	}
	return nil
}

func (a *ResourceAdder) Complete() error {
	if a.resource == "" && a.input == "" && a.access == "" {
		return nil
	}
	if a.resource == "" {
		return fmt.Errorf("%s meta data is missing (--%s)", a.typename, a.typename)
	}
	if a.access != "" && a.input != "" {
		return fmt.Errorf("either --input or --access is possible")
	}
	if a.access == "" && a.input == "" {
		return fmt.Errorf("either --input or --access is required")
	}

	if err := a.check("meta data", a.resource); err != nil {
		return err
	}
	if err := a.check("input", a.input); err != nil {
		return err
	}
	if err := a.check("access", a.access); err != nil {
		return err
	}
	return nil
}

func (a *ResourceAdder) GetResources() ([]Resources, error) {
	if a.resource == "" {
		return nil, nil
	}
	return []Resources{a}, nil
}

func (a *ResourceAdder) Origin() string {
	return a.typename + " (by options)"
}

func (a *ResourceAdder) Get(printer common.Printer, topts template.Options) (string, error) {
	printer.Printf("processing %s...\n", a.Origin())

	var data map[string]interface{}

	yaml.Unmarshal([]byte(a.resource), &data)

	if a.input != "" {
		var input map[string]interface{}
		yaml.Unmarshal([]byte(a.input), &input)
		data["input"] = input
	}
	if a.access != "" {
		var access map[string]interface{}
		yaml.Unmarshal([]byte(a.access), &access)
		data["access"] = access
	}
	r, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrapf(err, "cannot marshal option based %s specification", a.typename)
	}

	parsed, err := topts.Execute(string(r))
	if err != nil {
		return "", errors.Wrapf(err, "error during variable substitution for %q", a.Origin())
	}

	return parsed, nil
}

////////////////////////////////////////////////////////////////////////////////

type ResourceAdderCommand struct {
	utils.BaseCommand

	Templating template.Options
	Adder      AdderOptions

	Archive   string
	Resources []Resources
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

		rsc, err := o.Adder.GetResources()
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
		o.Resources = append(o.Resources, NewResourcesFile(p, o.FileSystem()))
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
		origin := source.Origin()
		parsed, err := source.Get(printer, o.Templating)
		if err != nil {
			return err
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
					return err
				}
				break
			}
			printer.Printf("  processing document %d...\n", i)
			if (tmp["input"] != nil || tmp["access"] != nil) && !h.RequireInputs() {
				return errors.Newf("invalid spec %d in %q: no input or access possible for %s", i, origin, listkey)
			}

			var list []json.RawMessage
			if reslist, ok := tmp[listkey]; ok {
				if len(tmp) != 1 {
					return errors.Newf("invalid spec %d in %q: either a list or a single spec possible", i, origin)
				}
				l, ok := reslist.([]interface{})
				if !ok {
					return errors.Newf("invalid spec %d in %q: invalid resource list", i, origin)
				}
				for j, e := range l {
					// cannot use json here, because yaml generates a map[interface{}]interface{}
					data, err := yaml.Marshal(e)
					if err != nil {
						return errors.Newf("invalid spec %d[%d] in %q: %s", i, j+1, origin, err.Error())
					}
					list = append(list, data)
				}
			} else {
				if len(tmp) == 0 {
					return errors.Newf("invalid spec %d in %q: empty", i, origin)
				}
				data, err := yaml.Marshal(tmp)
				if err != nil {
					return err
				}
				list = append(list, data)
			}

			for j, d := range list {
				printer.Printf("    processing index %d\n", j+1)
				var input *ResourceInput
				r, err := DecodeResource(d, h)
				if err != nil {
					return errors.Newf("invalid spec %d[%d] in %q: %s", i, j+1, origin, err)
				}

				if h.RequireInputs() {
					input, err = DecodeInput(d, o.Context)
					if err != nil {
						return errors.Newf("invalid spec %d[%d] in %q: %s", i, j+1, origin, err)
					}
					if err = Validate(input, ictx, origin); err != nil {
						return errors.Wrapf(err, "invalid spec %d[%d] in %q", i, j+1, origin)
					}
				}

				if err = r.Validate(o.Context, input); err != nil {
					return errors.Wrapf(err, "invalid spec %d[%d] in %q", i, j+1, origin)
				}

				resources = append(resources, NewResource(r, input, origin, i, j+1))
			}
		}
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
