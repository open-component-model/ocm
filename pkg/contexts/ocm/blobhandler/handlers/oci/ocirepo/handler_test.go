// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocirepo_test

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
	ocictf "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/keepblobattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/mapocirepoattr"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/oci/ocirepo"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ctf"
const ARCH2 = "/tmp/ctf2"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const COMPONENT2 = "github.com/mandelsoft/test2"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

func FakeOCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return "baseurl.io"
}

var _ = Describe("oci artifact transfer", func() {
	var env *Builder
	var ldesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder()

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			ldesc = OCIManifest1(env)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
					env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})

		_ = ldesc
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("it should copy a resource by value and export the OCI image but keep the local blob", func() {
		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
		keepblobattr.Set(env.OCMContext(), true)

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer tgt.Close()

		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		handler := standard.NewDefaultHandler(opts)

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
		Expect(env.DirExists(OUT)).To(BeTrue())

		list := Must(tgt.ComponentLister().GetComponents("", true))
		Expect(list).To(Equal([]string{COMPONENT}))
		comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
		data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

		fmt.Printf("%s\n", string(data))
		Expect(string(data)).To(StringEqualWithContext(`{"globalAccess":{"imageReference":"baseurl.io/ocm/value:v2.0","type":"ociArtifact"},"localReference":"sha256:b0692bcec00e0a875b6b280f3209d6776f3eca128adcb7e81e82fd32127c0c62","mediaType":"application/vnd.oci.image.manifest.v1+tar+gzip","referenceName":"ocm/value:v2.0","type":"localBlob"}`))
		ocirepo := genericocireg.GetOCIRepository(tgt)
		Expect(ocirepo).NotTo(BeNil())

		art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
		defer Close(art, "artifact")

		man := MustBeNonNil(art.ManifestAccess())
		Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
		Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

		blob := Must(man.GetBlob(ldesc.Digest))
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER))
	})

	It("it should copy a resource by value and export the OCI image", func() {
		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer tgt.Close()

		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		handler := standard.NewDefaultHandler(opts)

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
		Expect(env.DirExists(OUT)).To(BeTrue())

		list := Must(tgt.ComponentLister().GetComponents("", true))
		Expect(list).To(Equal([]string{COMPONENT}))
		comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
		data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

		fmt.Printf("%s\n", string(data))
		Expect(string(data)).To(StringEqualWithContext("{\"imageReference\":\"baseurl.io/ocm/value:v2.0\",\"type\":\"ociArtifact\"}"))

		ocirepo := genericocireg.GetOCIRepository(tgt)
		art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
		defer Close(art, "artifact")

		man := MustBeNonNil(art.ManifestAccess())
		Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
		Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

		blob := Must(man.GetBlob(ldesc.Digest))
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER))
	})

	It("it should copy a resource by value and export the OCI image with hashed repo name", func() {
		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer tgt.Close()

		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		handler := standard.NewDefaultHandler(opts)
		mapocirepoattr.Set(env.OCMContext(), &mapocirepoattr.Attribute{Mode: mapocirepoattr.ShortHashMode, Always: true})
		rdigest := "e9b6af2174cb2fb78b2882a1f487b01295b8f6bfa7e4c1ceb350440104c9ce65"

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
		Expect(env.DirExists(OUT)).To(BeTrue())

		list := Must(tgt.ComponentLister().GetComponents("", true))
		Expect(list).To(Equal([]string{COMPONENT}))
		comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
		data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

		fmt.Printf("%s\n", string(data))
		Expect(string(data)).To(StringEqualWithContext("{\"imageReference\":\"baseurl.io/" + rdigest[:8] + "/value:v2.0\",\"type\":\"ociArtifact\"}"))

		namespace := rdigest[:8] + "/value"
		ocirepo := genericocireg.GetOCIRepository(tgt)
		art := Must(ocirepo.LookupArtifact(namespace, OCIVERSION))
		defer Close(art, "artifact")

		man := MustBeNonNil(art.ManifestAccess())
		Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
		Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

		blob := Must(man.GetBlob(ldesc.Digest))
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER))
	})

	It("it should copy a resource by value and export the OCI image with hashed repo name and prefix", func() {
		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer tgt.Close()

		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		handler := standard.NewDefaultHandler(opts)
		prefix := "ocm"
		mapocirepoattr.Set(env.OCMContext(), &mapocirepoattr.Attribute{Mode: mapocirepoattr.ShortHashMode, Always: true, Prefix: &prefix})
		rdigest := "e9b6af2174cb2fb78b2882a1f487b01295b8f6bfa7e4c1ceb350440104c9ce65"

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
		Expect(env.DirExists(OUT)).To(BeTrue())

		list := Must(tgt.ComponentLister().GetComponents("", true))
		Expect(list).To(Equal([]string{COMPONENT}))
		comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
		data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

		fmt.Printf("%s\n", string(data))
		Expect(string(data)).To(StringEqualWithContext("{\"imageReference\":\"baseurl.io/ocm/" + rdigest[:8] + "/value:v2.0\",\"type\":\"ociArtifact\"}"))

		namespace := "ocm/" + rdigest[:8] + "/value"
		ocirepo := genericocireg.GetOCIRepository(tgt)
		art := Must(ocirepo.LookupArtifact(namespace, OCIVERSION))
		defer Close(art, "artifact")

		man := MustBeNonNil(art.ManifestAccess())
		Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
		Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

		blob := Must(man.GetBlob(ldesc.Digest))
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER))
	})
})
