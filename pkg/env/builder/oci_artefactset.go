// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
)

const T_OCIARTEFACTSET = "artefact set"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ArtefactSet(path string, fmt accessio.FileFormat, f ...func()) {
	r, err := artefactset.Open(accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0o777, fmt, accessio.PathFileSystem(b.FileSystem()))
	b.failOn(err)

	b.configure(&ociNamespace{NamespaceAccess: r, kind: T_OCIARTEFACTSET}, f)
}
