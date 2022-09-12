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

package helm

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm/loader"
)

type Spec struct {
	// PathSpec hold the path that points to the helm chart file
	cpi.PathSpec `json:",inline"`
	Version      string `json:"version,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path string) *Spec {
	return &Spec{
		PathSpec: cpi.NewPathSpec(TYPE, path),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx clictx.Context, inputFilePath string) field.ErrorList {
	allErrs := s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	if s.Path != "" {
		path := fldPath.Child("path")
		inputInfo, filePath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(path, filePath, err.Error()))
		}
		if !inputInfo.Mode().IsDir() && !inputInfo.Mode().IsRegular() {
			allErrs = append(allErrs, field.Invalid(path, filePath, "no regular file or directory"))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx clictx.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	_, inputPath, err := inputs.FileInfo(ctx, s.Path, inputFilePath)
	if err != nil {
		return nil, "", err
	}
	chart, err := loader.Load(inputPath, ctx.FileSystem())
	if err != nil {
		return nil, "", err
	}
	vers := chart.Metadata.Version
	if s.Version != "" {
		vers = s.Version
	}
	if vers == "" {
		vers = nv.GetVersion()
	}
	blob, err := helm.SynthesizeArtefactBlob(inputPath, ctx.FileSystem())
	if err != nil {
		return nil, "", err
	}
	name := chart.Name()
	hint := fmt.Sprintf("%s/%s:%s", nv.GetName(), name, vers)
	if name == "" {
		hint = fmt.Sprintf("%s:%s", nv.GetName(), vers)
	}
	return blob, hint, err
}
