// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch/comparch"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/validation/field"

	_ "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types"
)

type ResourceInput struct {
	Access *runtime.UnstructuredTypedObject `json:"access"`
	//Input  *inputs.BlobInput                `json:"input,omitempty"`
	Input *inputs.GenericInputSpec `json:"input,omitempty"`
}

type ResourceSpecHandler interface {
	RequireInputs() bool
	Decode(data []byte) (ResourceSpec, error)
	Set(v ocm.ComponentVersionAccess, r Resource, acc compdesc.AccessSpec) error
}

type ResourceSpec interface {
	GetName() string
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

type ResourceAdderCommand struct {
	utils.BaseCommand

	Archive    string
	Paths      []string
	Envs       []string
	Templating template.Options
}

func (o *ResourceAdderCommand) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.Envs, "settings", "s", nil, "settings file with variable settings (yaml)")
	o.Templating.AddFlags(fs)
}

func (o *ResourceAdderCommand) Complete(args []string) error {
	o.Archive = args[0]
	o.Templating.Complete(o.Context.FileSystem())

	err := o.Templating.ParseSettings(o.Context.FileSystem(), o.Envs...)
	if err != nil {
		return err
	}

	o.Paths = o.Templating.FilterSettings(args[1:]...)

	return nil
}

func (o *ResourceAdderCommand) ProcessResourceDescriptions(listkey string, h ResourceSpecHandler) error {
	fs := o.Context.FileSystem()
	resources := []*resource{}
	for _, filePath := range o.Paths {
		data, err := vfs.ReadFile(fs, filePath)
		if err != nil {
			return errors.Wrapf(err, "cannot read resource file %q", filePath)
		}

		parsed, err := o.Templating.Execute(string(data))
		if err != nil {
			return errors.Wrapf(err, "error during variable substitution for %q", filePath)
		}
		// sigs parser has no multi document stream parsing
		// but yaml.v3 does not recognize json tagged fields.
		// Therefore we first use the v3 parser to parse the multi doc,
		// marshal it again and finally unmarshal it with the sigs parser.
		decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(parsed)))
		i := 0
		for {
			var tmp map[string]interface{}

			i++
			err := decoder.Decode(&tmp)
			if err != nil {
				if err != io.EOF {
					return err
				}
				break
			}
			if (tmp["input"] != nil || tmp["access"] != nil) && !h.RequireInputs() {
				return errors.Newf("invalid spec %d in %q: no input or access possible for %s", i+1, filePath, listkey)
			}

			var list []json.RawMessage
			if reslist, ok := tmp[listkey]; ok {
				if len(tmp) != 1 {
					return errors.Newf("invalid spec %d in %q: either a list or a single spec possible", i+1, filePath)
				}
				l, ok := reslist.([]interface{})
				if !ok {
					return errors.Newf("invalid spec %d in %q: invalid resource list", i+1, filePath)
				}
				for j, e := range l {
					// cannot use json here, because yaml generates a map[interface{}]interface{}
					data, err = yaml.Marshal(e)
					if err != nil {
						return errors.Newf("invalid spec %d[%d] in %q: %s", i+1, j+1, filePath, err.Error())
					}
					list = append(list, data)
				}
			} else {
				if len(tmp) == 0 {
					return errors.Newf("invalid spec %d in %q: empty", i+1, filePath)
				}
				data, err := yaml.Marshal(tmp)
				if err != nil {
					return err
				}
				list = append(list, data)
			}

			for j, d := range list {
				var input *ResourceInput
				r, err := DecodeResource(d, h)
				if err != nil {
					return errors.Newf("invalid spec %d[%d] in %q: %s", i+1, j+1, filePath, err)
				}

				if h.RequireInputs() {
					input, err = DecodeInput(d, o.Context)
					if err != nil {
						return errors.Newf("invalid spec %d[%d] in %q: %s", i+1, j+1, filePath, err)
					}
					if err = Validate(input, o.Context, filePath); err != nil {
						return errors.Wrapf(err, "invalid spec %d[%d] in %q", i+1, j+1, filePath)
					}
				} else {

				}

				if err = r.Validate(o.Context, input); err != nil {
					return errors.Wrapf(err, "invalid spec %d[%d] in %q", i+1, j+1, filePath)
				}

				resources = append(resources, NewResource(r, input, filePath, i, j))
			}
		}
	}

	obj, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_WRITABLE, o.Archive, 0, accessio.PathFileSystem(fs))
	if err != nil {
		return err
	}
	defer obj.Close()

	for _, r := range resources {
		if h.RequireInputs() {
			if r.input.Input != nil {
				var acc ocm.AccessSpec
				// Local Blob
				blob, hint, berr := r.input.Input.GetBlob(o.Context, r.path)
				if berr != nil {
					return errors.Wrapf(err, "cannot get resource blob for %q(%s)", r.spec.GetName(), r.source)
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

func Validate(r *ResourceInput, ctx clictx.Context, inputFilePath string) error {
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
					raw, _ := r.Access.GetRaw()
					allErrs = append(allErrs, field.Invalid(fldPath.Child("access"), string(raw), err.Error()))
				} else {
					if acc.(ocm.AccessSpec).IsLocal(ctx.OCMContext()) {
						kind := runtime.ObjectVersionedType(r.Access.ObjectType).GetKind()
						allErrs = append(allErrs, field.Invalid(fldPath.Child("access", "type"), kind, "local access no possible"))
					}
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
