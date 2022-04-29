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
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm/loader"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Handler struct{}

var _ inputs.InputHandler = (*Handler)(nil)

func (h Handler) RequireFilePath() bool {
	return true
}

func (h Handler) Validate(fldPath *field.Path, ctx clictx.Context, input *inputs.BlobInput) field.ErrorList {
	allErrs := field.ErrorList{}
	path := fldPath.Child("compress")
	if input.CompressWithGzip != nil {
		allErrs = append(allErrs, field.Required(path, "compress option not possble for type helm"))
	}
	return allErrs
}

func (h Handler) CreateBlob(fs vfs.FileSystem, inputPath string, input *inputs.BlobInput) (accessio.TemporaryBlobAccess, string, error) {
	chart, err := loader.Load(fs, inputPath)
	if err != nil {
		return nil, "", err
	}
	blob, err := helm.SynthesizeArtefactBlob(inputPath, fs)
	if err != nil {
		return nil, "", err
	}
	return blob, chart.Metadata.Name, err
}
