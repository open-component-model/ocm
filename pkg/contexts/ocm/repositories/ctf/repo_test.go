// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf_test

import (
	"bytes"

	"github.com/mandelsoft/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/finalizer"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"github.com/tonglil/buflogr"

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
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

const COMPONENT = "github.com/mandelsoft/ocm"
const VERSION = "1.0.0"

var _ = Describe("access method", func() {
	var fs vfs.FileSystem
	ctx := ocm.DefaultContext()

	BeforeEach(func() {
		fs = memoryfs.New()
	})

	It("adds naked component version and later lookup", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a, "repository")
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c, "component")

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv, "version")

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		refmgmt.AllocLog.Trace("opening ctf")
		a = Must(ctf.Open(ctx, accessobj.ACC_READONLY, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)

		refmgmt.AllocLog.Trace("lookup component")
		c = Must(a.LookupComponent(COMPONENT))
		final.Close(c)

		refmgmt.AllocLog.Trace("lookup version")
		cv = Must(c.LookupVersion(VERSION))
		final.Close(cv)

		refmgmt.AllocLog.Trace("closing")
		MustBeSuccessful(final.Finalize())
	})

	It("adds naked component version and later shortcut lookup", func() {
		final := Finalizer{}
		defer Defer(final.Finalize)

		a := Must(ctf.Create(ctx, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a, "repository")
		c := Must(a.LookupComponent(COMPONENT))
		final.Close(c, "component")

		cv := Must(c.NewVersion(VERSION))
		final.Close(cv, "version")

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(final.Finalize())

		refmgmt.AllocLog.Trace("opening ctf")
		a = Must(ctf.Open(ctx, accessobj.ACC_READONLY, "ctf", 0o700, accessio.PathFileSystem(fs)))
		final.Close(a)

		refmgmt.AllocLog.Trace("lookup component version")
		cv = Must(a.LookupComponentVersion(COMPONENT, VERSION))
		final.Close(cv)

		refmgmt.AllocLog.Trace("closing")
		MustBeSuccessful(final.Finalize())
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

	It("provides error for invalid bloc access", func() {
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

	It("logs diff", func() {
		r := Must(ctf.Open(ctx, ctf.ACC_CREATE, "test.ctf", 0o700, accessio.FormatDirectory, accessio.PathFileSystem(fs)))
		defer Close(r, "repo")

		c := Must(r.LookupComponent("acme.org/test"))
		defer Close(c, "comp")

		cv := Must(c.NewVersion("v1"))

		ocmlog.PushContext(nil)
		ocmlog.Context().AddRule(logging.NewConditionRule(logging.DebugLevel, genericocireg.TAG_CDDIFF))
		var buf bytes.Buffer
		def := buflogr.NewWithBuffer(&buf)
		ocmlog.Context().SetBaseLogger(def)
		defer ocmlog.Context().ResetRules()
		defer ocmlog.PopContext()

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(cv.Close())

		cv = Must(c.LookupVersion("v1"))
		cv.GetDescriptor().Provider.Name = "acme.org"
		MustBeSuccessful(cv.Close())
		Expect("\n" + buf.String()).To(Equal(`
V[4] component descriptor has been changed realm ocm realm ocm/oci/mapping diff [ComponentSpec.ObjectMeta.Provider.Name: acme != acme.org]
V[4] component descriptor has been changed realm ocm realm ocm/oci/mapping diff [ComponentSpec.ObjectMeta.Provider.Name: acme != acme.org]
`))
	})

	It("handles readonly mode", func() {
		r := Must(ctf.Open(ctx, ctf.ACC_CREATE, "test.ctf", 0o700, accessio.FormatDirectory, accessio.PathFileSystem(fs)))
		defer Close(r, "repo")

		c := Must(r.LookupComponent("acme.org/test"))
		defer Close(c, "comp")

		cv := Must(c.NewVersion("v1"))

		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(cv.Close())

		cv = Must(c.LookupVersion("v1"))
		cv.SetReadOnly()
		cv.GetDescriptor().Provider.Name = "acme.org"
		ExpectError(cv.Close()).To(MatchError(accessio.ErrReadOnly))
	})
})
