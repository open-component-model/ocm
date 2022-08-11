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

package ctf

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

func SupportedFormats() []accessio.FileFormat {
	return ctf.SupportedFormats()
}

func init() {
	h := &repospechandler{}
	cpi.RegisterRepositorySpecHandler(h, "")
	cpi.RegisterRepositorySpecHandler(h, ctf.Type)
	cpi.RegisterRepositorySpecHandler(h, "ctf")
	for _, f := range SupportedFormats() {
		cpi.RegisterRepositorySpecHandler(h, string(f))
		cpi.RegisterRepositorySpecHandler(h, "ctf+"+string(f))
		cpi.RegisterRepositorySpecHandler(h, ctf.Type+"+"+string(f))
	}
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	if u.Info == "" {
		if u.Host == "" || u.Type == "" {
			return nil, nil
		}
	}
	spec, err := ctf.MapReference(ctx.OCIContext(), &oci.UniformRepositorySpec{
		Type:            u.Type,
		Host:            u.Host,
		Info:            u.Info,
		CreateIfMissing: u.CreateIfMissing,
		TypeHint:        u.TypeHint,
	})
	if err != nil || spec == nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
}
