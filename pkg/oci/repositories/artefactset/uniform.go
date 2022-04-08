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

package artefactset

import (
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/datacontext/vfsattr"
	"github.com/gardener/ocm/pkg/oci/cpi"
)

func init() {
	h := &repospechandler{}
	cpi.RegisterRepositorySpecHandler(h, "")
	cpi.RegisterRepositorySpecHandler(h, ArtefactSetType)
	for _, f := range SupportedFormats() {
		cpi.RegisterRepositorySpecHandler(h, string(f))
	}
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	path := u.Info
	if u.Info == "" {
		if u.Host == "" || u.Type == "" {
			return nil, nil
		}
		path = u.Host
	}
	fs := vfsattr.Get(ctx)

	if ok, err := accessobj.CheckFile(ArtefactSetType, u.Type == ArtefactSetType, path, fs, ArtefactSetDescriptorFileName); !ok || err != nil {
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
	}
	return NewRepositorySpec(accessobj.ACC_WRITABLE, path, accessio.FileFormat(u.Type), accessio.PathFileSystem(fs)), nil
}
