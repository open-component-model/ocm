// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	ocictf "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/oci/ocirepo"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const BASEURL = "baseurl.io"

func FakeOCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return BASEURL
}

var _ = Describe("disable upload", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()

		FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env.Builder)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})

		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers ctf with upload", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--copy-resources", ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 artifact[ociImage](ocm/value:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())

		ctf := Must(ctfocm.Open(env, accessobj.ACC_READONLY, OUT, 0, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "version")

		res := Must(cv.GetResource(metav1.NewIdentity("artifact")))

		acc := Must(res.Access())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		Expect(acc.Describe(env.OCMContext())).To(Equal("OCI artifact " + BASEURL + "/" + OCINAMESPACE + ":" + OCIVERSION))
	})

	It("transfers ctf without upload", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "components", "--disable-uploads", "--copy-resources", ARCH, OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
standard blob upload handlers are disabled.
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 artifact[ociImage](ocm/value:v2.0)...
...adding component version...
1 versions transferred
`))

		Expect(env.DirExists(OUT)).To(BeTrue())

		ctf := Must(ctfocm.Open(env, accessobj.ACC_READONLY, OUT, 0, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "version")

		res := Must(cv.GetResource(metav1.NewIdentity("artifact")))

		acc := Must(res.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))
	})
})
