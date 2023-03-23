// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package standard_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
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
const SIGNATURE = "test"
const SIGN_ALGO = rsa.Algorithm

var _ = Describe("Transfer handler", func() {
	var env *Builder
	var ldesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())

		env.RSAKeyPair(SIGNATURE)

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

		env.OCMCommonTransport(ARCH2, accessio.FormatDirectory, func() {
			env.Component(COMPONENT2, func() {
				env.Version(VERSION, func() {
					env.Reference("ref", COMPONENT, VERSION)
					env.Provider(PROVIDER)
				})
			})
		})

		FakeOCIRepo(env, OCIPATH, OCIHOST)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	DescribeTable("it should copy a resource by value to a ctf file", func(acc string, topts ...transferhandler.TransferOption) {
		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()
		opts := &standard.Options{}
		opts.SetResourcesByValue(true)
		transferhandler.ApplyOptions(opts, topts...)
		handler := standard.NewDefaultHandler(opts)
		//handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		list, err := tgt.ComponentLister().GetComponents("", true)
		Expect(err).To(Succeed())
		Expect(list).To(Equal([]string{COMPONENT}))
		comp, err := tgt.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
		Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
		data, err := json.Marshal(comp.GetDescriptor().Resources[1].Access)
		Expect(err).To(Succeed())

		fmt.Printf("%s\n", string(data))
		hash := HashManifest1(artifactset.DefaultArtifactSetDescriptorFileName)
		Expect(string(data)).To(StringEqualWithContext(fmt.Sprintf(acc, hash)))

		r, err := comp.GetResourceByIndex(1)
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

		blob, err := set.GetBlob(ldesc.Digest)
		Expect(err).To(Succeed())
		data, err = blob.Get()
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("manifestlayer"))
	},
		Entry("without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}"),
		Entry("with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			standard.KeepGlobalAccess()),
	)

	It("it should use additional resolver to resolve component ref", func() {
		parentSrc, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH2, 0, env)
		Expect(err).To(Succeed())
		cv, err := parentSrc.LookupComponentVersion(COMPONENT2, VERSION)
		Expect(err).To(Succeed())
		childSrc, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()
		handler, err := standard.New(standard.Recursive(), standard.Resolver(childSrc))
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		list, err := tgt.ComponentLister().GetComponents("", true)
		Expect(err).To(Succeed())
		Expect(list).To(ContainElements([]string{COMPONENT2, COMPONENT}))
		_, err = tgt.LookupComponentVersion(COMPONENT2, VERSION)
		Expect(err).To(Succeed())
		_, err = tgt.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
	})

	It("it should copy signatures", func() {
		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())

		resolver := ocm.NewCompoundResolver(src)

		opts := ocmsign.NewOptions(
			ocmsign.Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
			ocmsign.Resolver(resolver),
			ocmsign.Update(), ocmsign.VerifyDigests(),
		)
		Expect(opts.Complete(signingattr.Get(env.OCMContext()))).To(Succeed())
		digest := "4540fdc0c542b1cf266b77ef7e378ac892aa62b0487ffd82b6118b623834d5cf"
		dig, err := ocmsign.Apply(nil, nil, cv, opts)
		Expect(err).To(Succeed())
		fmt.Print(dig.Value)
		Expect(dig.Value).To(Equal(digest))

		Expect(len(cv.GetDescriptor().Signatures)).To(Equal(1))

		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()
		handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		resolver = ocm.NewCompoundResolver(tgt)

		opts = ocmsign.NewOptions(
			ocmsign.Resolver(resolver),
			ocmsign.VerifySignature(SIGNATURE),
			ocmsign.Update(), ocmsign.VerifyDigests(),
		)
		Expect(opts.Complete(signingattr.Get(env.OCMContext()))).To(Succeed())
		dig, err = ocmsign.Apply(nil, nil, cv, opts)
		Expect(err).To(Succeed())
		Expect(dig.Value).To(Equal(digest))
	})
})
