package genericaccess_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/env/builder"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactaccess/genericaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
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
