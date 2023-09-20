// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package relativeociref_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	ocictf "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/relativeociref"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/keepblobattr"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/oci/ocirepo"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
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
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
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

		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
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
		Expect(string(data)).To(StringEqualWithContext(fmt.Sprintf(`{"globalAccess":{"imageReference":"baseurl.io/ocm/value:v2.0","type":"ociArtifact"},"localReference":"%s","mediaType":"application/vnd.oci.image.manifest.v1+tar+gzip","referenceName":"ocm/value:v2.0","type":"localBlob"}`, hash)))
	})
})
