package addhdlrs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/validation/field"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/errkind"
	common2 "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/template"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	cliutils "ocm.software/ocm/cmds/ocm/common/utils"
)

func ProcessDescriptions(ctx clictx.Context, printer common2.Printer, templ template.Options, h ElementSpecHandler, sources []ElementSource) ([]Element, inputs.Context, error) {
	ictx := inputs.NewContext(ctx, printer, templ.Vars)

	elems := []Element{}
	for _, source := range sources {
		tmp, err := DetermineElementsForSource(ctx, ictx, templ, h, source)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "%s", source.Origin())
		}
		elems = append(elems, tmp...)
	}
	err := ValidateElementIdentities(h.Key(), elems)
	if err != nil {
		return nil, nil, err
	}
	ictx.Printf("found %d %s\n", len(elems), cliutils.Plural(h.Key(), len(elems)))
	return elems, ictx, nil
}

func DetermineElementsForSource(ctx clictx.Context, ictx inputs.Context, templ template.Options, h ElementSpecHandler, source ElementSource) ([]Element, error) {
	resources := []Element{}
	origin := source.Origin()

	ictx = ictx.Section("processing %s...", origin)
	r, err := source.Get()
	if err != nil {
		return nil, err
	}
	parsed, err := templ.Execute(r)
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
		ictx := ictx.Section("processing document %d...", i)

		var list []json.RawMessage
		listkey := cliutils.Plural(h.Key(), 0)
		if reslist, ok := tmp[listkey]; ok {
			if len(tmp) != 1 {
				return nil, errors.Newf("invalid %s spec %d: either a list or a single spec possible for %s (found keys %s)", h.Key(), i, listkey, utils.StringMapKeys(tmp))
			}
			l, ok := reslist.([]interface{})
			if !ok {
				return nil, errors.Newf("invalid spec %d: invalid %s list", i, h.Key())
			}
			for j, e := range l {
				// cannot use json here, because yaml generates a map[interface{}]interface{}
				data, err := yaml.Marshal(e)
				if err != nil {
					return nil, errors.Newf("invalid %s spec %d[%d]: %s", h.Key(), i, j+1, err.Error())
				}
				list = append(list, data)
			}
		} else {
			if entry, ok := tmp[h.Key()]; ok {
				if m, ok := entry.(map[string]interface{}); ok {
					if len(tmp) != 1 {
						return nil, errors.Newf("invalid %s spec %d: either a list or a single spec possible for %s (found keys %s)", h.Key(), i, listkey, utils.StringMapKeys(tmp))
					}
					tmp = m
				}
			}
			if len(tmp) == 0 {
				return nil, errors.Newf("invalid %s spec %d: empty", h.Key(), i)
			}
			data, err := yaml.Marshal(tmp)
			if err != nil {
				return nil, err
			}
			list = append(list, data)
		}

		for j, d := range list {
			r, err := DetermineElementForData(ctx, ictx.Section("processing index %d", j+1), origin.Sub(i, j+1), d, h)
			if err != nil {
				return nil, errors.Newf("invalid %s spec %d[%d]: %s", h.Key(), i, j+1, err)
			}
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// DetermineElements maps a list of raw element specifications into an evaluated element list.
func DetermineElements(ctx clictx.Context, ictx inputs.Context, origin SourceInfo, d interface{}, h ElementSpecHandler) ([]Element, error) {
	list, ok := d.([]interface{})
	if !ok {
		return nil, fmt.Errorf("element list expected")
	}
	var elements []Element
	for i, e := range list {
		m, ok := e.(map[string]interface{})
		if !ok {
			return nil, errors.Newf("invalid %s spec %d: map expected", h.Key(), i+1)
		}
		r, err := DetermineElement(ctx, ictx, origin.Sub(i+1), m, h)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid %s spec %d", h.Key(), i+1)
		}
		elements = append(elements, r)
	}
	return elements, nil
}

func DetermineElement(ctx clictx.Context, ictx inputs.Context, si SourceInfo, d map[string]interface{}, h ElementSpecHandler) (Element, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return DetermineElementForData(ctx, ictx, si, data, h)
}

func DetermineElementForData(ctx clictx.Context, ictx inputs.Context, si SourceInfo, d []byte, h ElementSpecHandler) (Element, error) {
	var input *ResourceInput
	r, err := DecodeElement(d, h)
	if err != nil {
		return nil, err
	}

	if h.RequireInputs() {
		input, err = DecodeInput(d, ctx)
		if err != nil {
			return nil, err
		}
		if err = Validate(input, ictx, general.OptionalDefaulted(si.Origin(), input.SourceFile)); err != nil {
			return nil, err
		}
	}

	if err = r.Validate(ctx, input); err != nil {
		return nil, err
	}

	return NewElement(r, input, si, d), nil
}

func DecodeElement(data []byte, h ElementSpecHandler) (ElementSpec, error) {
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
	// delete(plainOrig, "input")
	err = cliutils.CheckForUnknown(nil, plainOrig, plainAccepted).ToAggregate()
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

	var plainOrig map[string]interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &plainOrig)
	if err != nil {
		return nil, err
	}
	var fldPath *field.Path
	err = CheckForUnknown(fldPath.Child("input"), plainOrig["input"], input.Input)
	return &input, err
}

func CheckForUnknownForData(fldPath *field.Path, orig []byte, accepted interface{}) error {
	var plainOrig map[string]interface{}
	err := runtime.DefaultYAMLEncoding.Unmarshal(orig, &plainOrig)
	if err != nil {
		return err
	}
	return CheckForUnknown(fldPath, plainOrig, accepted)
}

func CheckForUnknown(fldPath *field.Path, plainOrig, accepted interface{}) error {
	adata, err := runtime.DefaultJSONEncoding.Marshal(accepted)
	if err != nil {
		return err
	}
	var plainAccepted interface{}
	err = runtime.DefaultJSONEncoding.Unmarshal(adata, &plainAccepted)
	if err != nil {
		return err
	}
	return cliutils.CheckForUnknown(fldPath, plainOrig, plainAccepted).ToAggregate()
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
				acc, err := r.Access.Evaluate(ctx.OCMContext())
				if err != nil {
					if errors.IsErrUnknown(err) {
						err.(errors.Kinded).SetKind(errkind.KIND_ACCESSMETHOD)
					}
					raw, _ := r.Access.GetRaw()
					allErrs = append(allErrs, field.Invalid(fldPath.Child("access"), string(raw), err.Error()))
				} else if acc.IsLocal(ctx.OCMContext()) {
					kind := runtime.GetKind(r.Access)
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

func ValidateElementIdentities(kind string, elems []Element) error {
	list := errors.ErrList()
	ids := map[string]SourceInfo{}
	for _, r := range elems {
		var i interface{}
		err := runtime.DefaultYAMLEncoding.Unmarshal(r.Data(), &i)
		if err != nil {
			return errors.Wrapf(err, "cannot eval data %q", string(r.Data()))
		}
		id := r.Spec().GetRawIdentity()
		dig := id.Digest()
		if s, ok := ids[string(dig)]; ok {
			list.Add(fmt.Errorf("duplicate %s identity %s (%s and %s)", kind, id, r.Source(), s))
		}
		ids[string(dig)] = r.Source()
	}
	return list.Result()
}

// ValidateElementSpecIdentities validate the element specifications
// taken from some source (for example a resources.yaml or component-constructor.yaml).
// The parameter src somehow identifies the element source, for example
// the path of the parsed file.
func ValidateElementSpecIdentities(kind string, src string, elems []ElementSpec) error {
	list := errors.ErrList()
	ids := map[string]int{}
	for i, r := range elems {
		id := r.GetRawIdentity()
		dig := id.Digest()
		if s, ok := ids[string(dig)]; ok {
			list.Add(fmt.Errorf("duplicate %s identity %s (%s index %d and %d)", kind, id, src, i+1, s+1))
		}
		ids[string(dig)] = i
	}
	return list.Result()
}

func PrintElements(p common2.Printer, elems []Element, outfile string, fss ...vfs.FileSystem) error {
	if outfile != "" && outfile != "-" {
		f, err := utils.FileSystem(fss...).OpenFile(outfile, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0o644)
		if err != nil {
			return errors.Wrapf(err, "cannot create output file %q", outfile)
		}
		p = common2.NewPrinter(f)
	}

	for _, r := range elems {
		var i interface{}
		err := runtime.DefaultYAMLEncoding.Unmarshal(r.Data(), &i)
		if err != nil {
			return errors.Wrapf(err, "cannot eval data %q", string(r.Data()))
		}
		data, err := runtime.DefaultYAMLEncoding.Marshal(i)
		if err != nil {
			return err
		}
		p.Printf("---\n%s\n", string(data))
	}
	return nil
}
