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

package builder

import (
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/repositories/artefactset"
)

const T_OCIARTEFACTSET = "artefact set"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ArtefactSet(path string, fmt accessio.FileFormat, f ...func()) {
	r, err := artefactset.Open(accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0777, fmt, accessio.PathFileSystem(b.FileSystem()))
	b.failOn(err)

	b.configure(&oci_namespace{NamespaceAccess: r, kind: T_OCIARTEFACTSET}, f)
}
