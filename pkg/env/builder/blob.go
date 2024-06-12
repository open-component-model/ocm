package builder

import (
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/dirtree"
	"github.com/open-component-model/ocm/pkg/utils"
)

const T_BLOBACCESS = "blob access"

func (b *Builder) BlobStringData(mime string, data string) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	*(b.blob) = blobaccess.ForData(mime, []byte(data))
}

func (b *Builder) BlobData(mime string, data []byte) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	*(b.blob) = blobaccess.ForData(mime, data)
}

func (b *Builder) BlobFromFile(mime string, path string) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	*(b.blob) = blobaccess.ForFile(mime, path, b.FileSystem())
	b.failOn(utils.ValidateObject(*(b.blob)))
}

func (b *Builder) BlobFromDirTree(path string, opts ...dirtree.Option) {
	b.expect(b.blob, T_BLOBACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	var err error
	*(b.blob), err = dirtree.BlobAccess(path, append([]dirtree.Option{dirtree.WithFileSystem(b.FileSystem())}, opts...)...)
	b.failOn(err)
}
