package genericaccess_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm/compdesc"
	me "ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/accessio"
)

const (
	OCIPATH = "/tmp/oci"
	OCIHOST = "alias"
)

var _ = Describe("dir tree resource access", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates resource", func() {
		spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION))

		acc := Must(me.ResourceAccess(env.OCMContext(), compdesc.NewResourceMeta("test", resourcetypes.OCI_IMAGE, compdesc.LocalRelation), spec))

		Expect(acc.ReferenceHint()).To(Equal(OCINAMESPACE + ":" + OCIVERSION))
		Expect(acc.GlobalAccess()).To(BeNil())
		Expect(acc.Meta().Type).To(Equal(resourcetypes.OCI_IMAGE))

		blob := Must(acc.BlobAccess())
		defer Defer(blob.Close, "blob")
		Expect(blob.MimeType()).To(Equal(artifactset.MediaType(artdesc.MediaTypeImageManifest)))
	})
})
