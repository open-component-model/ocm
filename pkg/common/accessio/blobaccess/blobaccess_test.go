// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blobaccess_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/spi"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("blob access ref counting", func() {
	It("handles ref count less access", func() {

		blob := blobaccess.ForString(mime.MIME_TEXT, "test")
		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		MustBeSuccessful(blob.Close())
		Expect(dup.Get()).To(Equal([]byte("test")))
		MustBeSuccessful(dup.Close())
	})

	It("handles ref count ", func() {
		blob := spi.NewBlobAccessForBase(spi.BaseAccessForDataAccess(mime.MIME_TEXT, blobaccess.DataAccessForString("test")))
		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		ExpectError(blob.Close()).To(Equal(blobaccess.ErrClosed))
		ExpectError(blob.Get()).To(Equal(blobaccess.ErrClosed))
		ExpectError(blob.Reader()).To(Equal(blobaccess.ErrClosed))
		Expect(dup.Get()).To(Equal([]byte("test")))
		ExpectError(dup.Digest().String()).To(Equal("sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"))
		ExpectError(dup.Size()).To(Equal(int64(4)))
		MustBeSuccessful(dup.Close())
	})

	It("releases temp file", func() {
		temp := Must(os.CreateTemp("", "testfile*"))
		path := temp.Name()
		temp.Close()
		blob := blobaccess.ForTemporaryFilePath(mime.MIME_TEXT, path, osfs.OsFs)

		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeTrue())

		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeTrue())
		MustBeSuccessful(dup.Close())
		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeFalse())
	})
})
