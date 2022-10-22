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

package spiff

import (
	"encoding/json"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
	"github.com/open-component-model/ocm/pkg/clisupport"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func ConfigHandler() clisupport.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig,
		options.LibrariesOption, options.ValuesOption)
}

type Spec struct {
	cpi.MediaFileSpec `json:",inline"`
	// Values provide additional binding for the template processing
	Values json.RawMessage `json:"values,omitempty"`
	// Libraries specifies a list of spiff libraries to include in template processing
	Libraries []string `json:"libraries,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path, mediatype string, compress bool, values interface{}, libs ...string) (*Spec, error) {
	var v []byte
	var err error
	if x, ok := values.([]byte); ok {
		v = x
	} else {
		v, err = runtime.DefaultJSONEncoding.Marshal(values)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid values")
		}
	}
	return &Spec{
		MediaFileSpec: cpi.NewMediaFileSpec(TYPE, path, mediatype, compress),
		Values:        v,
		Libraries:     libs,
	}, nil
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	allErrs := (&file.ProcessSpec{s.MediaFileSpec, nil}).Validate(fldPath, ctx, inputFilePath)
	for i, v := range s.Libraries {
		pathField := fldPath.Index(i)
		fileInfo, filePath, err := inputs.FileInfo(ctx, v, inputFilePath)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(pathField, filePath, err.Error()))
		} else if !fileInfo.Mode().IsRegular() {
			allErrs = append(allErrs, field.Invalid(pathField, filePath, "no regular file"))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	return (&file.ProcessSpec{s.MediaFileSpec, s.process}).GetBlob(ctx, nv, inputFilePath)
}

func (s *Spec) process(ctx inputs.Context, inputFilePath string, data []byte) ([]byte, error) {
	fs, err := cwdfs.New(ctx.FileSystem(), inputFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create local directory view %q", inputFilePath)
	}
	env := spiffing.New().WithFeatures(features.INTERPOLATION, features.CONTROL).WithFileSystem(fs)

	templ, err := env.Unmarshal(s.Path, data)
	if err != nil {
		return nil, err
	}

	add := map[string]interface{}{}
	if ctx.Variables() != nil {
		add["values"] = ctx.Variables()
	}

	if s.Values != nil {
		var values interface{}
		err := json.Unmarshal(s.Values, &values)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse values")
		}
		add["inputvalues"] = values
	}
	if len(add) > 0 {
		env, err = env.WithValues(add)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "invalid values")
	}
	var stubs []spiffing.Node
	for i, l := range s.Libraries {
		stub, err := env.UnmarshalSource(spiffing.NewSourceFile(l, ctx.FileSystem()))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid spiff library %d(%s)", i+1, l)
		}
		stubs = append(stubs, stub)
	}
	out, err := env.Cascade(templ, stubs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to process template")
	}
	return env.Marshal(out)
}

func AddConfig(opts clisupport.ConfigOptions, config clisupport.Config) error {
	if err := cpi.AddMediaFileSpecConfig(opts, config); err != nil {
		return err
	}
	if v, ok := opts.GetValue(options.LibrariesOption.Name()); ok {
		config["libraries"] = v
	}
	if v, ok := opts.GetValue(options.LibrariesOption.Name()); ok {
		config["values"] = v
	}
	return nil
}
