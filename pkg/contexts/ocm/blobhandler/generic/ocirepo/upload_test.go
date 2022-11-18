// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocirepo_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	tenv "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/mime"
)

const COMP = "github.com/compa"
const VERS = "1.0.0"
const CA = "ca"
const CTF = "ctf"
const COPY = "ctf.copy"
const TARGET = "/tmp/target"

const OCIHOST = "alias"
const OCIPATH = "/tmp/source"
const OCINAMESPACE = "ocm/value"
const OCIVERSION = "v2.0"

var _ = Describe("upload", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(tenv.NewEnvironment())

		// fake OCI registry
		spec := Must(ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCIPATH, accessio.PathFileSystem(env.FileSystem())))
		env.OCIContext().SetAlias(OCIHOST, spec)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			env.Namespace(OCINAMESPACE, func() {
				env.Manifest(OCIVERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "manifestlayer")
					})
				})
			})
		})

		env.OCICommonTransport(TARGET, accessio.FormatDirectory)

		env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERS, func() {
			env.Provider("mandelsoft")
			env.Resource("value", "", resourcetypes.OCI_IMAGE, v1.LocalRelation, func() {
				env.Access(
					ociartefact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
				)
			})
		})

		ca := Must(comparch.Open(env.OCMContext(), accessobj.ACC_READONLY, CA, 0, env))
		oca := accessio.OnceCloser(ca)
		defer Close(oca)

		ctf := Must(ctfocm.Create(env.OCMContext(), accessobj.ACC_CREATE, CTF, 0700, env))
		octf := accessio.OnceCloser(ctf)
		defer Close(octf)

		handler := Must(standard.New(standard.ResourcesByValue()))

		MustBeSuccessful(transfer.TransferVersion(nil, nil, ca, ctf, handler))

		// now we have a transport archive with local blob for the image
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers oci artefact", func() {
		ctx := env.OCMContext()

		ctf := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, CTF, 0700, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMP, VERS))
		ocv := accessio.OnceCloser(cv)
		defer Close(ocv)
		ra := Must(cv.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		// transfer component
		copy := Must(ctfocm.Create(ctx, accessobj.ACC_CREATE, COPY, 0700, env))
		ocopy := accessio.OnceCloser(copy)
		defer Close(ocopy)

		// prepare upload to target OCI repo
		attr := ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")
		ociuploadattr.Set(ctx, attr)

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, copy, nil))

		// check type
		cv2 := Must(copy.LookupComponentVersion(COMP, VERS))
		ocv2 := accessio.OnceCloser(cv2)
		defer Close(ocv2)
		ra = Must(cv2.GetResourceByIndex(0))
		acc = Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(ociartefact.Type))
		val := Must(ctx.AccessSpecForSpec(acc))
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartefact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0"))

		attr.Close()
		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtefact("copy/ocm/value", "v2.0")).To(BeTrue())
	})

	It("transfers oci artefact with named handler", func() {
		ctx := env.OCMContext()

		ctf := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, CTF, 0700, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMP, VERS))
		ocv := accessio.OnceCloser(cv)
		defer Close(ocv)
		ra := Must(cv.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		// transfer component
		copy := Must(ctfocm.Create(ctx, accessobj.ACC_CREATE, COPY, 0700, env))
		ocopy := accessio.OnceCloser(copy)
		defer Close(ocopy)

		// prepare upload to target OCI repo
		attr := ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")
		MustBeSuccessful(registration.RegisterBlobHandlerByName(ctx, "ocm/ociArtifacts", attr))

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, copy, nil))

		// check type
		cv2 := Must(copy.LookupComponentVersion(COMP, VERS))
		ocv2 := accessio.OnceCloser(cv2)
		defer Close(ocv2)
		ra = Must(cv2.GetResourceByIndex(0))
		acc = Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(ociartefact.Type))
		val := Must(ctx.AccessSpecForSpec(acc))
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartefact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0"))

		attr.Close()
		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtefact("copy/ocm/value", "v2.0")).To(BeTrue())
	})
})
