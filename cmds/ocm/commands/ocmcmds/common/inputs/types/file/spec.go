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

package file

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/clisupport"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
)

func ConfigHandler() clisupport.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig)
}

type Spec struct {
	cpi.MediaFileSpec `json:",inline"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path, mediatype string, compress bool) *Spec {
	return &Spec{
		MediaFileSpec: cpi.NewMediaFileSpec(TYPE, path, mediatype, compress),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	return (&ProcessSpec{s.MediaFileSpec, nil}).Validate(fldPath, ctx, inputFilePath)
}

func (s *Spec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	return (&ProcessSpec{s.MediaFileSpec, nil}).GetBlob(ctx, nv, inputFilePath)
}

func AddConfig(opts clisupport.ConfigOptions, config clisupport.Config) error {
	return cpi.AddMediaFileSpecConfig(opts, config)
}
