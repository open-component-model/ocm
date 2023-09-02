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
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
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

type optionsChecker struct {
	standard.TransferOptionsCreator
}

var _ transferhandler.TransferOption = (*optionsChecker)(nil)

func (o *optionsChecker) ApplyTransferOption(options transferhandler.TransferOptions) error {
	if _, ok := options.(*standard.Options); !ok {
		return fmt.Errorf("unexpected options type %T", options)
	}
	return nil
}

var _ = Describe("Transfer handler", func() {
	var env *Builder
	var ldesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder()

		env.RSAKeyPair(SIGNATURE)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			ldesc = OCIManifest1(env)
		})

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					TestDataResource(env)
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

	})

	AfterEach(func() {
		env.Cleanup()
	})

	DescribeTable("it should copy a resource by value to a ctf file", func(acc string, topts ...transferhandler.TransferOption) {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(src, "source")
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "source cv")
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
		defer Close(tgt, "target")

		// handler, err := standard.New(standard.ResourcesByValue())
		p, buf := common.NewBufferedPrinter()
		opts := append(topts, standard.ResourcesByValue(), transfer.WithPrinter(p), &optionsChecker{})
		MustBeSuccessful(transfer.Transfer(cv, tgt, opts...))
		Expect(env.DirExists(OUT)).To(BeTrue())

		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[PlainText]...
...resource 1 artifact[ociImage](ocm/value:v2.0)...
...adding component version...
`))

		var nested finalizer.Finalizer
		defer Defer(nested.Finalize)

		list := Must(tgt.ComponentLister().GetComponents("", true))
		Expect(list).To(Equal([]string{COMPONENT}))
		tcv := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		nested.Close(tcv, "target cv")
		Expect(len(tcv.GetDescriptor().Resources)).To(Equal(2))
		data := Must(json.Marshal(tcv.GetDescriptor().Resources[1].Access))

		fmt.Printf("%s\n", string(data))
		hash := HashManifest1(artifactset.DefaultArtifactSetDescriptorFileName)
		Expect(string(data)).To(StringEqualWithContext(fmt.Sprintf(acc, hash)))

		tcd := tcv.GetDescriptor().Copy()
		r := Must(tcv.GetResourceByIndex(1))
		meth := Must(r.AccessMethod())
		nested.Close(meth, "method")
		reader := Must(meth.Reader())
		nested.Close(reader, "reader")
		set := Must(artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(reader)))
		nested.Close(set, "set")

		_, blob := Must2(set.GetBlobData(ldesc.Digest))
		nested.Close(blob)
		data = Must(blob.Get())
		Expect(string(data)).To(Equal("manifestlayer"))

		MustBeSuccessful(nested.Finalize())

		// retransfer
		buf.Reset()
		MustBeSuccessful(transfer.Transfer(cv, tgt, opts...))
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
  version "github.com/mandelsoft/test:v1" already present -> skip transport`))
		tcv = Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		nested.Close(tcv, "target cv")
		Expect(tcd).To(DeepEqual(tcv.GetDescriptor()))
		MustBeSuccessful(nested.Finalize())

		// modify volatile
		cv.GetDescriptor().Labels.Set("new", "newvalue")
		tcd.Labels.Set("new", "newvalue")
		buf.Reset()
		MustBeSuccessful(transfer.Transfer(cv, tgt, opts...))
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
  updating volatile properties of "github.com/mandelsoft/test:v1"
...adding component version...
`))
		tcv = Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		nested.Close(tcv, "target cv")
		Expect(tcd).To(DeepEqual(tcv.GetDescriptor()))
		MustBeSuccessful(nested.Finalize())

		// modify one artifact and overwrite
		MustBeSuccessful(cv.SetResourceBlob(Must(cv.GetResourceByIndex(0)).Meta().Fresh(), accessio.BlobAccessForString(mime.MIME_TEXT, "otherdata"), "", nil))
		tcd.Resources[0].Digest = DS_OTHERDATA
		tcd.Resources[0].Access = Must(runtime.ToUnstructuredVersionedTypedObject(localblob.New("sha256:"+D_OTHERDATA, "", mime.MIME_TEXT, nil)))
		buf.Reset()
		MustBeSuccessful(transfer.Transfer(cv, tgt, append(opts, standard.Overwrite())...))
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
warning:   version "github.com/mandelsoft/test:v1" already present, but differs, because some artifact digests are changed (transport enforced by overwrite option)
...resource 0 testdata[PlainText] (overwrite)
...resource 1 artifact[ociImage](ocm/value:v2.0) (already present)
...adding component version...
`))
		tcv = Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		nested.Close(tcv, "target cv")
		Expect(tcd).To(DeepEqual(tcv.GetDescriptor()))
		MustBeSuccessful(nested.Finalize())

	},
		Entry("without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}"),
		Entry("with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			standard.KeepGlobalAccess()),
	)

	It("disable value transport of oci access", func() {
		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()

		opts := &standard.Options{}
		Expect(opts.Apply(standard.ResourcesByValue(), standard.OmitAccessTypes(ociartifact.Type))).To(Succeed())
		Expect(opts.IsResourcesByValue()).To(BeTrue())
		Expect(opts.IsAccessTypeOmitted(ociartifact.Type)).To(BeTrue())
		Expect(opts.IsAccessTypeOmitted(ociartifact.LegacyType)).To(BeFalse())

		handler := standard.NewDefaultHandler(opts)
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

		r, err := comp.GetResourceByIndex(1)
		Expect(err).To(Succeed())

		a, err := r.Access()
		Expect(err).To(Succeed())
		Expect(a.GetType()).To(Equal(ociartifact.Type))
	})

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
			ocmsign.Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
			ocmsign.Resolver(resolver),
			ocmsign.Update(), ocmsign.VerifyDigests(),
		)
		Expect(opts.Complete(signingattr.Get(env.OCMContext()))).To(Succeed())
		digest := "45aefd9317bde6c66d5edca868cf7b9a5313a6f965609af4e58bbfd44ae6e92c"
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
