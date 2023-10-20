// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/finalizer"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
)

const COMPONENT = "github.com/mandelsoft/ocm"
const VERSION = "1.0.0"

var _ = Describe("access method", func() {
	var fs vfs.FileSystem
	ctx := ocm.DefaultContext()

	BeforeEach(func() {
		fs = memoryfs.New()
	})

	It("adds component version", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		// add resource
		MustBeSuccessful(cv.SetResourceBlob(compdesc.NewResourceMeta("text1", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
		Expect(Must(cv.GetResource(compdesc.NewIdentity("text1"))).Meta().Digest).To(Equal(DS_TESTDATA))

		// add resource with digest
		meta := compdesc.NewResourceMeta("text2", resourcetypes.PLAIN_TEXT, metav1.LocalRelation)
		meta.SetDigest(DS_TESTDATA)
		MustBeSuccessful(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
		Expect(Must(cv.GetResource(compdesc.NewIdentity("text2"))).Meta().Digest).To(Equal(DS_TESTDATA))

		// reject resource with wrong digest
		meta = compdesc.NewResourceMeta("text3", resourcetypes.PLAIN_TEXT, metav1.LocalRelation)
		meta.SetDigest(TextResourceDigestSpec("fake"))
		Expect(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil)).To(MatchError("unable to set resource: digest mismatch: " + D_TESTDATA + " != fake"))

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		a = Must(ctf.Open(ctx, accessobj.ACC_READONLY, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)

		cv = Must(a.LookupComponentVersion(COMPONENT, VERSION))
		final.Close(cv)
	})

	It("adds omits unadded new component version", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		MustBeSuccessful(final.Finalize())

		a = Must(ctf.Open(ctx, accessobj.ACC_READONLY, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)

		_, err := a.LookupComponentVersion(COMPONENT, VERSION)

		Expect(err).To(MatchError(ContainSubstring("component version \"github.com/mandelsoft/ocm:1.0.0\" not found: oci artifact \"1.0.0\" not found in component-descriptors/github.com/mandelsoft/ocm")))
	})

	It("provided error for invalid bloc access", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv)

		// add resource
		Expect(ErrorFrom((cv.SetResourceBlob(compdesc.NewResourceMeta("text1", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blobaccess.ForFile(mime.MIME_TEXT, "non-existing-file"), "", nil)))).To(MatchError(`file "non-existing-file" not found`))

		MustBeSuccessful(final.Finalize())
	})
})
