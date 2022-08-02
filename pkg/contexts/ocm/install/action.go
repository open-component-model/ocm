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

package install

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/xeipuuv/gojsonschema"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/config/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ValidateByScheme(src []byte, schemedata []byte) error {
	data, err := yaml.YAMLToJSON(src)
	if err != nil {
		return errors.Wrapf(err, "converting data to json")
	}
	schemedata, err = yaml.YAMLToJSON(schemedata)
	if err != nil {
		return errors.Wrapf(err, "converting scheme to json")
	}
	documentLoader := gojsonschema.NewBytesLoader(data)

	scheme, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemedata))
	if err != nil {
		return errors.Wrapf(err, "invalid scheme")
	}
	res, err := scheme.Validate(documentLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		errs := res.Errors()
		errMsg := errs[0].String()
		for i := 1; i < len(errs); i++ {
			errMsg = fmt.Sprintf("%s;%s", errMsg, errs[i].String())
		}
		return errors.New(errMsg)
	}

	return nil
}

func ExecuteAction(d Driver, name string, spec *Specification, creds *Credentials, params []byte, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*OperationResult, error) {
	var err error

	var executor *Executor
	for _, e := range spec.Executors {
		if e.Actions == nil {
			executor = &e
			break
		}
		for _, a := range e.Actions {
			if a == name {
				executor = &e
				break
			}
		}
	}
	if executor == nil {
		return nil, errors.Newf("no executor found for action %s", name)
	}

	ccfg := config.New()
	if len(spec.CredentialsRequest.Credentials) > 0 {
		if creds == nil {
			return nil, errors.Newf("credential settings required")
		}
		ccfg, err = GetCredentials(octx.CredentialsContext(), creds, &spec.CredentialsRequest)
		if err != nil {
			return nil, errors.Wrapf(err, "credential evaluation failed")
		}
	}
	stubs := []spiffing.Node{}
	spiff := spiffing.New().WithFeatures(features.CONTROL, features.INTERPOLATION)

	if len(spec.Template) > 0 {

		templ, err := spiff.Unmarshal("template", spec.Template)
		if err != nil {
			return nil, errors.Newf("invalid parameter template: %s", err)
		}
		stubs = append(stubs, templ)
	}

	if len(spec.Scheme) > 0 {
		if params == nil {
			err = ValidateByScheme([]byte("{}"), spec.Scheme)
		} else {
			err = ValidateByScheme(params, spec.Scheme)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "parameter file validation failed")
		}
	}
	if params != nil {
		stub, err := spiff.Unmarshal("parameters", params)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid settings")
		}
		stubs = append(stubs, stub)
	}

	if len(stubs) == 0 {
		params = []byte("{}")
	} else {
		for i, lib := range spec.Libraries {
			res, eff, err := utils.ResolveResourceReference(cv, lib, resolver)
			if err != nil {
				return nil, errors.ErrNotFound("library resource %s not found", executor.ImageResourceRef.String())
			}
			if eff != cv {
				defer eff.Close()
			}
			m, err := res.AccessMethod()
			if err != nil {
				return nil, errors.ErrNotFound("cannot access library resource", lib.String())
			}
			data, err := m.Get()
			m.Close()
			if err != nil {
				return nil, errors.ErrNotFound("cannot access library resource", lib.String())
			}
			lib, err := spiff.Unmarshal(fmt.Sprintf("lib%d", i), data)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid spiff lib")
			}
			stubs = append(stubs, lib)
		}
		cfg, err := spiff.Cascade(stubs[0], stubs[1:])
		if err != nil {
			return nil, errors.Wrapf(err, "error processing parameters")
		}
		params, err = spiff.Marshal(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "parameter marshalling")
		}
	}
	image := executor.Image
	if image == nil {
		if cv == nil {
			return nil, errors.Newf("resource access not possible without component version")
		}
		res, eff, err := utils.ResolveResourceReference(cv, executor.ImageResourceRef, resolver)
		if err != nil {
			return nil, errors.ErrNotFound("executor resource %s not found", executor.ImageResourceRef.String())
		}
		if res.Meta().Type != "ociImage" {
			return nil, errors.ErrNotFound("invalid executor resource type", res.Meta().Type, executor.ImageResourceRef.String())
		}
		ref, err := utils.GetOCIArtefactRef(octx, res)
		if err != nil {
			return nil, errors.Wrapf(err, "image for executor resource %s not found", executor.ImageResourceRef.String())
		}
		if eff != cv {
			eff.Close()
		}
		// TODO: get digest if provided
		image = &Image{
			Ref: ref,
		}
	}
	fmt.Printf("using exector image %s\n", image.Ref)
	op := &Operation{
		Action:      name,
		Image:       *image,
		Environment: nil,
		Files:       nil,
		Outputs:     nil,
		Out:         nil,
		Err:         nil,
	}

	op.Files = map[string]accessio.BlobAccess{}
	if ccfg != nil {
		data, err := runtime.DefaultYAMLEncoding.Marshal(ccfg)
		if err != nil {
			return nil, errors.Wrapf(err, "marshalling ocm config failed")
		}
		op.Files[InputOCMConfig] = accessio.BlobAccessForData(mime.MIME_OCTET, data)
	}
	if params != nil {
		op.Files[InputParameters] = accessio.BlobAccessForData(mime.MIME_OCTET, params)
	}
	if executor.Config != nil {
		op.Files[InputConfig] = accessio.BlobAccessForData(mime.MIME_OCTET, executor.Config)
	}
	if cv != nil {
		fs, err := osfs.NewTempFileSystem()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create temp file system")
		}
		defer vfs.Cleanup(fs)
		repo, err := ctf.Create(octx, accessobj.ACC_CREATE, "arch", 0600, accessio.FormatTGZ, accessio.PathFileSystem(fs))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create repo for component version")
		}
		err = transfer.TransferVersion(nil, nil, cv.Repository(), cv, repo, nil)
		repo.Close()
		if err != nil {
			return nil, errors.Wrapf(err, "component version transport failed")
		}
		op.Files[InputOCMRepo] = accessio.BlobAccessForFile(mime.MIME_OCTET, "arch", fs)
	}
	op.Outputs = executor.Outputs

	err = d.SetConfig(map[string]string{})
	if err != nil {
		return nil, err
	}
	return d.Exec(op)
}
