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
	"bytes"
	"fmt"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/config/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/spiff"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
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

type ExecutorContext struct {
	Spec  ExecutorSpecification
	Image *Image
	CV    ocm.ComponentVersionAccess
}

func GetResource(res ocm.ResourceAccess, target interface{}) error {
	m, err := res.AccessMethod()
	if err != nil {
		return errors.Wrapf(err, "failed to instantiate access")
	}
	data, err := m.Get()
	m.Close()
	if err != nil {
		return errors.Wrapf(err, "cannot get resource content")
	}
	return runtime.DefaultYAMLEncoding.Unmarshal(data, target)
}

func DetermineExecutor(executor *Executor, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*ExecutorContext, error) {
	espec := ExecutorContext{Image: executor.Image}

	if espec.Image == nil {
		if cv == nil {
			return nil, errors.Newf("resource access not possible without component version")
		}
		if executor.ResourceRef == nil {
			return nil, errors.Newf("executor resource reference required for toi package executor")
		}
		res, eff, err := utils.ResolveResourceReference(cv, *executor.ResourceRef, resolver)
		if err != nil {
			return nil, errors.ErrNotFoundWrap(err, "executor resource", executor.ResourceRef.String())
		}
		defer func() {
			if eff != nil {
				eff.Close()
			}
		}()
		switch res.Meta().Type {
		case consts.OCIImage:
		case TypeTOIExecutor:
			err := GetResource(res, &espec.Spec)
			if err != nil {
				return nil, errors.ErrInvalidWrap(err, "toi executor")
			}
			espec.Image = espec.Spec.Image
			if espec.Image == nil {
				if cv == nil {
					return nil, errors.Newf("resource access not possible without component version")
				}
				if espec.Spec.ImageRef == nil {
					return nil, errors.Newf("executor image reference required for toi executor")
				}
				var eff2 ocm.ComponentVersionAccess
				res, eff2, err = utils.ResolveResourceReference(eff, *espec.Spec.ImageRef, resolver)
				if err != nil {
					return nil, errors.ErrNotFoundWrap(err, "executor resource", executor.ResourceRef.String())
				}
				defer eff2.Close()
			}
		default:
			return nil, errors.ErrInvalid("executor resource type", res.Meta().Type, executor.ResourceRef.String())
		}

		if res.Meta().Type != consts.OCIImage {
			return nil, errors.ErrInvalid("executor resource type", res.Meta().Type)
		}
		ref, err := utils.GetOCIArtefactRef(octx, res)
		if err != nil {
			return nil, errors.Wrapf(err, "image ref for executor resource %s not found", executor.ResourceRef.String())
		}
		espec.Image = &Image{
			Ref: ref,
		}
		espec.CV, eff = eff, nil
	}
	return &espec, nil
}

func mappingKeyFor(value string, m map[string]string) string {
	if m == nil {
		return value
	}
	for k, v := range m {
		if v == value {
			return k
		}
	}
	return ""
}

// CheckCredentialRequests determine required credentials for executor.
func CheckCredentialRequests(executor *Executor, spec *PackageSpecification, espec *ExecutorSpecification) (map[string]CredentialsRequestSpec, map[string]string, error) {
	credentials := spec.Credentials
	credmapping := map[string]string{}

	if len(espec.Credentials) > 0 {
		if len(spec.CredentialsRequest.Credentials) > 0 {
			// first, determine mapping and subset of defined spec required for executor
			for k := range spec.Credentials {
				ke := k
				if executor.CredentialMapping != nil {
					if m, ok := executor.CredentialMapping[k]; ok {
						ke = m
					}
					if _, ok := espec.Credentials[ke]; ok {
						credmapping[k] = ke
					} else {
						delete(credentials, k)
					}
				}
			}

			// second, check for spec errors and complete package spec
			for ke, e := range espec.Credentials {
				ko := mappingKeyFor(ke, executor.CredentialMapping)
				if ko == "" {
					return nil, nil, errors.Newf("credential mapping missing for executor credential key %q", ke)
				}
				if o, ok := credentials[ko]; !ok {
					if !e.Optional {
						// implicit inheritance of executor spec setting
						credentials[ko] = e
						credmapping[ko] = ke
					}
				} else {
					if err := o.Match(&e); err != nil {
						return nil, nil, errors.Wrapf(err, "credential %q does not match executor setting %q", ko, ke)
					}
				}
			}
		} else {
			// no credential requests specified for package, use the one from the executor
			credentials = espec.Credentials
		}
	} else {
		if len(executor.CredentialMapping) > 0 {
			// determine subset of credentials required for executor
			credmapping = executor.CredentialMapping
			for k := range credentials {
				if _, ok := credmapping[k]; !ok {
					delete(credentials, k)
				} else {
					credmapping[k] = k
				}
			}
		} else {
			// assume to require all as defined
			for k := range credentials {
				credmapping[k] = k
			}
		}
	}
	return credentials, credmapping, nil
}

func ProcessConfig(name string, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, template []byte, config []byte, libraries []metav1.ResourceReference, schemedata []byte) ([]byte, error) {
	var err error

	if len(config) == 0 {
		if len(schemedata) > 0 {
			err = ValidateByScheme([]byte("{}"), schemedata)
			if err != nil {
				return nil, errors.Wrapf(err, name+" validation failed")
			}
		}
		if len(template) == 0 {
			return nil, nil
		}
	}

	stubs := spiff.Options{}
	for i, lib := range libraries {
		res, eff, err := utils.ResolveResourceReference(cv, lib, resolver)
		if err != nil {
			return nil, errors.ErrNotFound("library resource %s not found", lib.String())
		}
		defer eff.Close()
		m, err := res.AccessMethod()
		if err != nil {
			return nil, errors.ErrNotFound("cannot access library resource", lib.String())
		}
		data, err := m.Get()
		m.Close()
		if err != nil {
			return nil, errors.ErrNotFound("cannot access library resource", lib.String())
		}
		stubs.Add(spiff.StubData(fmt.Sprintf("spiff lib%d", i), data))
	}

	if len(schemedata) > 0 || len(template) == 0 {
		// process input without template first to have final version without bfore using template
		// to be verified by json scheme
		if config != nil {
			config, err = spiff.CascadeWith(spiff.Context(octx), spiff.TemplateData(name, config), stubs)
			if err != nil {
				return nil, errors.Wrapf(err, "error processing "+name)
			}
		}
	}
	if len(schemedata) > 0 {
		logrus.Infof("validating %s by scheme...", name)
		err = ValidateByScheme(config, schemedata)
		if err != nil {
			return nil, errors.Wrapf(err, name+" validation failed")
		}
	}
	if len(template) > 0 {
		config, err = spiff.CascadeWith(spiff.Context(octx), spiff.TemplateData(name+" template", template), spiff.StubData(name, config), stubs)
		if err != nil {
			return nil, errors.Wrapf(err, "error processing "+name+" template")
		}
	}
	return config, err
}

func ExecuteAction(d Driver, name string, spec *PackageSpecification, creds *Credentials, params []byte, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*OperationResult, error) {
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

	// validate executor config
	espec, err := DetermineExecutor(executor, octx, cv, resolver)
	if err != nil {
		return nil, err
	}

	if espec.Spec.Actions != nil {
		found := false
		for _, a := range espec.Spec.Actions {
			if a == name {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.ErrNotSupported("action", name, "toi executor "+executor.ResourceRef.String())
		}
	}

	if espec.Spec.Outputs != nil {
		list := errors.ErrListf("invalid outputs")
		for o := range executor.Outputs {
			if _, ok := espec.Spec.Outputs[o]; !ok {
				list.Add(fmt.Errorf("output %s not available from executor", o))
			}
		}
		if list.Len() > 0 {
			return nil, list.Result()
		}
	}
	// prepare executor config
	econfig, err := ProcessConfig("executor config", octx, espec.CV, resolver, espec.Spec.Template, executor.Config, espec.Spec.Libraries, espec.Spec.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "error executor config")
	}

	if econfig == nil {
		logrus.Infof("no executor config found")
	} else {
		logrus.Infof("using executor config:\n%s", utils2.IndentLines(string(econfig), "  "))
	}
	// handle credentials
	credentials, credmapping, err := CheckCredentialRequests(executor, spec, &espec.Spec)
	if err != nil {
		return nil, err
	}

	// prepare ocm config with credential settings
	ccfg := config.New()
	if len(credentials) > 0 {
		if creds == nil {
			return nil, errors.Newf("credential settings required")
		}
		ccfg, err = GetCredentials(octx.CredentialsContext(), creds, credentials, credmapping)
		if err != nil {
			return nil, errors.Wrapf(err, "credential evaluation failed")
		}
	}

	// prepare user config
	params, err = ProcessConfig("parameter data", octx, cv, resolver, spec.Template, params, spec.Libraries, spec.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "error processing parameters")
	}
	if params == nil {
		logrus.Infof("no parameter config found")
	} else {
		logrus.Infof("using package parameters:\n%s", utils2.IndentLines(string(params), "  "))
	}

	if executor.ParameterMapping != nil {
		orig := params
		params, err = spiff.CascadeWith(
			spiff.TemplateData("executor parameter mapping", executor.ParameterMapping),
			spiff.StubData("package config", params))

		if err == nil {
			if !bytes.Equal(orig, params) {
				logrus.Infof("using executor parameters:\n%s", utils2.IndentLines(string(params), "  "))
			}
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error mapping parameters to executor")
	}

	names := []string{}
	for n := range credentials {
		m := credmapping[n]
		names = append(names, n+"->"+m)
	}
	sort.Strings(names)
	logrus.Infof("using executor image %s (%s)", espec.Image.Ref, executor.ResourceRef.String())
	logrus.Infof("with credentials: %v", names)
	op := &Operation{
		Action:      name,
		Image:       *espec.Image,
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
	if econfig != nil {
		op.Files[InputConfig] = accessio.BlobAccessForData(mime.MIME_OCTET, econfig)
	}
	if cv != nil {
		fs, err := osfs.NewTempFileSystem()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create temp file system")
		}
		defer vfs.Cleanup(fs)
		repo, err := ctf.Create(octx, accessobj.ACC_CREATE, "arch", 0o600, accessio.FormatTGZ, accessio.PathFileSystem(fs))
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
	op.ComponentVersion = common.VersionedElementKey(cv).String()
	return d.Exec(op)
}
