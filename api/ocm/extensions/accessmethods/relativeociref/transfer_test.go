package relativeociref_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	ocictf "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/relativeociref"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/keepblobattr"
	storagecontext "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const OUT = "/tmp/res"

func FakeOCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return "baseurl.io"
}

var _ = Describe("Transfer handler", func() {
	var env *Builder
	var ldesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder()

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			ldesc = OCIManifest1(env)
		})

		env.OCMCommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, COMPVERS, func() {
				env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
					env.Access(
						relativeociref.New(OCINAMESPACE + ":" + OCIVERSION),
					)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("it should copy an image by value to a ctf file", func() {
		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, OCIPATH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMP, COMPVERS)
		Expect(err).To(Succeed())
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()
		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		handler := standard.NewDefaultHandler(opts)
		// handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		list, err := tgt.ComponentLister().GetComponents("", true)
		Expect(err).To(Succeed())
		Expect(list).To(Equal([]string{COMP}))
		comp, err := tgt.LookupComponentVersion(COMP, COMPVERS)
		Expect(err).To(Succeed())
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(1))
		Expect(comp.GetDescriptor().Resources[0].Access.GetType()).To(Equal(localblob.Type))
		data, err := json.Marshal(comp.GetDescriptor().Resources[0].Access)
		Expect(err).To(Succeed())

		fmt.Printf("%s\n", string(data))
		hash := HashManifest1(artifactset.DefaultArtifactSetDescriptorFileName)
		Expect(string(data)).To(StringEqualWithContext(fmt.Sprintf(`{"localReference":"%s","mediaType":"application/vnd.oci.image.manifest.v1+tar+gzip","referenceName":"ocm/value:v2.0","type":"localBlob"}`, hash)))

		r, err := comp.GetResourceByIndex(0)
		Expect(err).To(Succeed())
		meth, err := r.AccessMethod()
		Expect(err).To(Succeed())
		defer meth.Close()
		reader, err := meth.Reader()
		Expect(err).To(Succeed())
		defer reader.Close()
		set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(reader))
		Expect(err).To(Succeed())
		defer set.Close()

		_, blob, err := set.GetBlobData(ldesc.Digest)
		Expect(err).To(Succeed())
		data, err = blob.Get()
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("manifestlayer"))
	})

	It("it should copy an image by value to an oci repo with uploader", func() {
		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
		keepblobattr.Set(env.OCMContext(), true)

		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, OCIPATH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMP, COMPVERS)
		Expect(err).To(Succeed())

		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer tgt.Close()
		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		handler := standard.NewDefaultHandler(opts)
		// handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		list, err := tgt.ComponentLister().GetComponents("", true)
		Expect(err).To(Succeed())
		Expect(list).To(Equal([]string{COMP}))
		comp, err := tgt.LookupComponentVersion(COMP, COMPVERS)
		Expect(err).To(Succeed())
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(1))
		Expect(comp.GetDescriptor().Resources[0].Access.GetType()).To(Equal(localblob.Type))
		data, err := json.Marshal(comp.GetDescriptor().Resources[0].Access)
		Expect(err).To(Succeed())

		fmt.Printf("%s\n", string(data))
		hash := HashManifest1(artifactset.DefaultArtifactSetDescriptorFileName)
		Expect(string(data)).To(StringEqualWithContext(fmt.Sprintf(`{"globalAccess":{"imageReference":"baseurl.io/ocm/value:v2.0@sha256:`+D_OCIMANIFEST1+`","type":"ociArtifact"},"localReference":"%s","mediaType":"application/vnd.oci.image.manifest.v1+tar+gzip","referenceName":"ocm/value:v2.0","type":"localBlob"}`, hash)))
	})
})
