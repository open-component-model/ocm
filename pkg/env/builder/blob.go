// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/dirtree"
)

const T_BLOBACCESS = "blob access"

func (b *Builder) BlobStringData(mime string, data string) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	*(b.blob) = accessio.BlobAccessForData(mime, []byte(data))
}

func (b *Builder) BlobData(mime string, data []byte) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	*(b.blob) = accessio.BlobAccessForData(mime, data)
}

func (b *Builder) BlobFromFile(mime string, path string) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	*(b.blob) = accessio.BlobAccessForFile(mime, path, b.FileSystem())
	b.failOn(accessio.ValidateObject(*(b.blob)))
}

func (b *Builder) BlobFromDirTree(path string, opts ...dirtree.Option) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	var err error
	*(b.blob), err = dirtree.BlobAccessForDirTree(path, append([]dirtree.Option{dirtree.WithFileSystem(b.FileSystem())}, opts...)...)
	b.failOn(err)
}
