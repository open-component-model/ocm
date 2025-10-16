package standard_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/compositionmodeattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/resolvers"
	. "ocm.software/ocm/api/ocm/testhelper"
	ocmsign "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ARCH       = "/tmp/ctf"
	ARCH2      = "/tmp/ctf2"
	PROVIDER   = "mandelsoft"
	VERSION    = "v1"
	COMPONENT  = "github.com/mandelsoft/test"
	COMPONENT2 = "github.com/mandelsoft/test2"
	OUT        = "/tmp/res"
	OCIPATH    = "/tmp/oci"
	OCIHOST    = "alias"
	SIGNATURE  = "test"
	SIGN_ALGO  = rsa.Algorithm
)

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

	It("cancelled", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(src, "source")
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "source cv")
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer Close(tgt, "target")

		// handler, err := standard.New(standard.ResourcesByValue())
		p, buf := common.NewBufferedPrinter()
		opts := []transferhandler.TransferOption{standard.ResourcesByValue(), &optionsChecker{}}

		ctx, cancel := context.WithCancel(context.Background())
		ctx = common.WithPrinter(ctx, p)
		cancel()
		ExpectError(transfer.TransferWithContext(ctx, cv, tgt, opts...)).To(MatchError(context.Canceled))

		Expect(buf.String()).To(Equal("transfer cancelled by caller\n"))
	})

	It("test", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(src, "source")
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "source cv")
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer Close(tgt, "target")

		tcv := Must(tgt.NewComponentVersion(cv.GetName(), cv.GetVersion()))
		defer Close(tcv, "target version")

		res := Must(cv.GetResource(metav1.NewIdentity("artifact")))
		acc := Must(res.Access())

		m := Must(acc.AccessMethod(cv))
		defer Close(m, "method")

		blob := Must(accspeccpi.BlobAccessForAccessMethod(m))
		defer Close(blob, "blob")
		MustBeSuccessful(tcv.SetResourceBlob(res.Meta(), blob, "", nil, ocm.SkipVerify()))

		MustBeSuccessful(tgt.AddComponentVersion(tcv))
	})

	DescribeTable("it should copy a resource by value to a ctf file", func(acc string, compose bool, topts ...transferhandler.TransferOption) {
		compositionmodeattr.Set(env.OCMContext(), compose)
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(src, "source")
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "source cv")
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
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
	},
		Entry("without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}",
			false),
		Entry("with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			false,
			standard.KeepGlobalAccess()),

		Entry("with composition and without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}",
			true),
		Entry("with composition and with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			true,
			standard.KeepGlobalAccess()),
	)

	DescribeTable("it should copy a resource by value to a ctf file", func(acc string, compose bool, topts ...transferhandler.TransferOption) {
		compositionmodeattr.Set(env.OCMContext(), compose)
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(src, "source")
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "source cv")
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
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
		MustBeSuccessful(cv.SetResourceBlob(Must(cv.GetResourceByIndex(0)).Meta().Fresh(), blobaccess.ForString(mime.MIME_TEXT, "otherdata"), "", nil))
		tcd.Resources[0].Digest = DS_OTHERDATA
		tcd.Resources[0].Access = Must(runtime.ToUnstructuredVersionedTypedObject(localblob.New("sha256:"+D_OTHERDATA, "", mime.MIME_TEXT, nil)))
		buf.Reset()
		MustBeSuccessful(transfer.Transfer(cv, tgt, append(opts, standard.Overwrite())...))
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
warning:   version "github.com/mandelsoft/test:v1" already present, but differs because some artifact digests are changed (transport enforced by overwrite option)
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
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}",
			false),
		Entry("with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			false,
			standard.KeepGlobalAccess()),

		Entry("with composition and without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}",
			true),
		Entry("with composition and with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			true,
			standard.KeepGlobalAccess()),
	)

	DescribeTable("it should copy a resource by value to a ctf file for re-transport", func(acc string, mode bool, topts ...transferhandler.TransferOption) {
		compositionmodeattr.Set(env.OCMContext(), mode)
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(src, "source")
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "source cv")
		tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer Close(tgt, "target")

		// transfer by reference, first
		p, buf := common.NewBufferedPrinter()
		opts := append(topts, transfer.WithPrinter(p), &optionsChecker{})
		MustBeSuccessful(transfer.Transfer(cv, tgt, opts...))
		Expect(env.DirExists(OUT)).To(BeTrue())
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
...resource 0 testdata[PlainText]...
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

		tcd := tcv.GetDescriptor().Copy()
		r := Must(tcv.GetResourceByIndex(1))
		Expect(Must(r.Access()).GetType()).To(Equal(ociartifact.Type))
		MustBeSuccessful(nested.Finalize())

		// retransfer with value transport
		buf.Reset()
		opts = append(topts, standard.ResourcesByValue(), transfer.WithPrinter(p), &optionsChecker{})
		MustBeSuccessful(transfer.Transfer(cv, tgt, opts...))
		Expect(env.DirExists(OUT)).To(BeTrue())
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
  version "github.com/mandelsoft/test:v1" already present -> but requires resource transport
...resource 0 testdata[PlainText] (already present)
...resource 1 artifact[ociImage](ocm/value:v2.0) (copy)
...adding component version...
`))

		list = Must(tgt.ComponentLister().GetComponents("", true))
		Expect(list).To(Equal([]string{COMPONENT}))
		tcv = Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		nested.Close(tcv, "target cv")
		Expect(len(tcv.GetDescriptor().Resources)).To(Equal(2))
		data = Must(json.Marshal(tcv.GetDescriptor().Resources[1].Access))

		fmt.Printf("%s\n", string(data))
		hash := HashManifest1(artifactset.DefaultArtifactSetDescriptorFileName)
		Expect(string(data)).To(StringEqualWithContext(fmt.Sprintf(acc, hash)))

		tcd = tcv.GetDescriptor().Copy()
		r = Must(tcv.GetResourceByIndex(1))
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
		MustBeSuccessful(cv.SetResourceBlob(Must(cv.GetResourceByIndex(0)).Meta().Fresh(), blobaccess.ForString(mime.MIME_TEXT, "otherdata"), "", nil))
		tcd.Resources[0].Digest = DS_OTHERDATA
		tcd.Resources[0].Access = Must(runtime.ToUnstructuredVersionedTypedObject(localblob.New("sha256:"+D_OTHERDATA, "", mime.MIME_TEXT, nil)))
		buf.Reset()
		transfer.Breakpoints = true
		MustBeSuccessful(transfer.Transfer(cv, tgt, append(opts, standard.ResourcesByValue(), standard.Overwrite())...))
		transfer.Breakpoints = false
		Expect(string(buf.Bytes())).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test:v1"...
warning:   version "github.com/mandelsoft/test:v1" already present, but differs because some artifact digests are changed (transport enforced by overwrite option)
...resource 0 testdata[PlainText] (overwrite)
...resource 1 artifact[ociImage](ocm/value:v2.0) (already present)
...adding component version...
`))
		tcv = Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		nested.Close(tcv, "target cv")
		ntcd := tcv.GetDescriptor()
		Expect(tcd).To(DeepEqual(ntcd))
		MustBeSuccessful(nested.Finalize())
	},
		Entry("without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}",
			false),
		Entry("with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			false,
			standard.KeepGlobalAccess()),
		Entry("with composition and without preserve global",
			"{\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\""+OCINAMESPACE+":"+OCIVERSION+"\",\"type\":\"localBlob\"}",
			true),
		Entry("with composition and with preserve global",
			"{\"globalAccess\":{\"imageReference\":\"alias.alias/ocm/value:v2.0\",\"type\":\"ociArtifact\"},\"localReference\":\"%s\",\"mediaType\":\"application/vnd.oci.image.manifest.v1+tar+gzip\",\"referenceName\":\"ocm/value:v2.0\",\"type\":\"localBlob\"}",
			true,
			standard.KeepGlobalAccess()),
	)

	It("disable value transport of oci access", func() {
		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env)
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
		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env)
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
		src, err := ctf.Open(env.OCMContext(), accessobj.ACC_WRITABLE, ARCH, 0, env)
		Expect(err).To(Succeed())
		cv, err := src.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())

		resolver := resolvers.NewCompoundResolver(src)

		opts := ocmsign.NewOptions(
			ocmsign.Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
			ocmsign.Resolver(resolver),
			ocmsign.Update(), ocmsign.VerifyDigests(),
		)
		Expect(opts.Complete(env.OCMContext())).To(Succeed())
		digest := "45aefd9317bde6c66d5edca868cf7b9a5313a6f965609af4e58bbfd44ae6e92c"
		dig, err := ocmsign.Apply(nil, nil, cv, opts)
		Expect(err).To(Succeed())
		fmt.Print(dig.Value)
		Expect(dig.Value).To(Equal(digest))

		Expect(len(cv.GetDescriptor().Signatures)).To(Equal(1))

		tgt, err := ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env)
		Expect(err).To(Succeed())
		defer tgt.Close()
		handler, err := standard.New(standard.ResourcesByValue())
		Expect(err).To(Succeed())
		err = transfer.TransferVersion(nil, nil, cv, tgt, handler)
		Expect(err).To(Succeed())
		Expect(env.DirExists(OUT)).To(BeTrue())

		resolver = resolvers.NewCompoundResolver(tgt)

		opts = ocmsign.NewOptions(
			ocmsign.Resolver(resolver),
			ocmsign.VerifySignature(SIGNATURE),
			ocmsign.Update(), ocmsign.VerifyDigests(),
		)
		Expect(opts.Complete(env.OCMContext())).To(Succeed())
		dig, err = ocmsign.Apply(nil, nil, cv, opts)
		Expect(err).To(Succeed())
		Expect(dig.Value).To(Equal(digest))
	})
})
