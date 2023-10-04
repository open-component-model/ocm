// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessio_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("blob access ref counting", func() {
	It("handles ref counts", func() {

		blob := accessio.BlobAccessForString(mime.MIME_TEXT, "test")
		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		Expect(blob.Close()).To(MatchError("closed"))
		MustBeSuccessful(dup.Close())
	})

	It("releases temp file", func() {
		temp := Must(os.CreateTemp("", "testfile*"))
		path := temp.Name()
		temp.Close()
		blob := accessio.BlobAccessForTemporaryFilePath(mime.MIME_TEXT, path, osfs.OsFs)

		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeTrue())

		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeTrue())
		MustBeSuccessful(dup.Close())
		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeFalse())
	})
})
