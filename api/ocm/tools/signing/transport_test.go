package signing_test

import (
	"fmt"

	"github.com/mandelsoft/goutils/finalizer"
	// . "ocm.software/ocm/api/ocm/tools/signing"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	. "ocm.software/ocm/api/ocm/testhelper"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	LABEL_SIG       = "non-volatile"
	LABEL_VOL       = "volatile"
	LABEL_VOL_NEW   = "new-volatile"
	LABEL_VOL_LOCAL = "local-volatile"

	TARGET = "/tmp/target"

	D_COMPA = "05dae7a597029ffbde9f3460b6ca9d059c3d78ca7c72a1d640c4555c6e91c921"
)

type Modifier func(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor)

var _ = Describe("transport and signing", func() {
	var env *Builder

	descData := Must(AsStructure(`
  component:
    componentReferences: []
    labels:
    - name: non-volatile
      signing: true
      value: signed
    - name: volatile
      value: orig volatile
    name: github.com/mandelsoft/test
    provider:
      labels:
      - name: non-volatile
        signing: true
        value: signed
      - name: volatile
        value: orig volatile
      name: acme.org
    repositoryContexts: []
    resources:
    - access:
        localReference: sha256:${DIGEST}
        mediaType: text/plain
        type: localBlob
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: genericBlobDigest/v1
        value: ${DIGEST}
      labels:
      - name: non-volatile
        signing: true
        value: signed
      - name: volatile
        value: orig volatile
      name: testdata
      relation: local
      type: PlainText
      version: v1
    - access:
        imageReference: alias.alias/ocm/value:v2.0
        type: ociArtifact
      digest:
        hashAlgorithm: SHA-256
        normalisationAlgorithm: ociArtifactDigest/v1
        value: ${IMAGE}
      name: image
      relation: local
      type: ociImage
      version: v1
    sources: []
    version: v1
  meta:
    configuredSchemaVersion: v2
`, Substitutions{
		"DIGEST": D_TESTDATA,
		"IMAGE":  D_OCIMANIFEST1,
	}))

	BeforeEach(func() {
		env = NewBuilder()

		env.RSAKeyPair(SIGNATURE, SIGNATURE2)

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENTA, VERSION, func() {
				env.Label(LABEL_SIG, "signed", metav1.WithSigning())
				env.Label(LABEL_VOL, "orig volatile")

				env.Provider("acme.org", func() {
					env.Label(LABEL_SIG, "signed", metav1.WithSigning())
					env.Label(LABEL_VOL, "orig volatile")
				})
				TestDataResource(env, func() {
					env.Label(LABEL_SIG, "signed", metav1.WithSigning())
					env.Label(LABEL_VOL, "orig volatile")
				})
				OCIArtifactResource1(env, "image", OCIHOST)
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("provides expected base component", func() {
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
		defer Close(cv, "cv")
		Expect(cv.GetDescriptor()).To(YAMLEqual(descData))
	})

	DescribeTable("retransports after local signing", func(modify Modifier) {
		target := Must(ctf.Create(env, accessobj.ACC_WRITABLE, TARGET, 0o700, env))
		defer Close(target, "target")

		repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
		defer Close(cv, "cv")

		desc := cv.GetDescriptor().Copy()

		printer, buf := common.NewBufferedPrinter()
		// transport
		handler := Must(standard.New(standard.ResourcesByValue()))
		MustBeSuccessful(transfer.TransferVersion(printer, nil, cv, target, handler))

		var targetfinal finalizer.Finalizer
		defer Defer(targetfinal.Finalize, "target objects")

		tcv := Must(target.LookupComponentVersion(COMPONENTA, VERSION))
		targetfinal.Close(tcv, "tcv")

		ra := desc.GetResourceIndexByIdentity(metav1.NewIdentity("image"))
		Expect(ra).To(BeNumerically(">=", 0))
		// indeed, the artifact set archive hash seems to be reproducible
		desc.Resources[ra].Access = localblob.New("sha256:"+H_OCIARCHMANIFEST1, "ocm/value:v2.0", "application/vnd.oci.image.manifest.v1+tar+gzip", nil)
		Expect(tcv.GetDescriptor()).To(YAMLEqual(desc))

		descSigned := tcv.GetDescriptor().Copy()

		// sign in target
		sopts := signing.NewOptions(
			signing.Sign(signingattr.Get(env.OCMContext()).GetSigner(SIGN_ALGO), SIGNATURE),
			signing.Update(), signing.VerifyDigests(),
		)
		spec := Must(signing.Apply(printer, nil, tcv, sopts))

		digSpec := &metav1.DigestSpec{
			HashAlgorithm:          "SHA-256",
			NormalisationAlgorithm: "jsonNormalisation/v1",
			Value:                  D_COMPA,
		}
		descSigned.Signatures = []compdesc.Signature{
			{
				Name:   SIGNATURE,
				Digest: *digSpec,
				Signature: metav1.SignatureSpec{
					Algorithm: rsa.Algorithm,
					MediaType: rsa.MediaType,
					Value:     tcv.GetDescriptor().Signatures[0].Signature.Value,
				},
			},
		}
		Expect(spec).To(Equal(&metav1.DigestSpec{
			HashAlgorithm:          "SHA-256",
			NormalisationAlgorithm: "jsonNormalisation/v1",
			Value:                  D_COMPA,
		}))
		Expect(tcv.GetDescriptor()).To(DeepEqual(descSigned))

		merged := tcv.GetDescriptor().Copy()

		// change volatile data in origin
		modify(cv, tcv, merged)
		MustBeSuccessful(tcv.Update())

		MustBeSuccessful(targetfinal.Finalize())

		// transfer changed volatile data
		buf.Reset()
		err := transfer.TransferVersion(printer, nil, cv, target, handler)
		fmt.Printf("%s\n", buf.String())
		MustBeSuccessful(err)

		tcv = Must(target.LookupComponentVersion(COMPONENTA, VERSION))
		targetfinal.Close(tcv, "tcv")
		Expect(tcv.GetDescriptor()).To(DeepEqual(merged))

		// verify signature after modification
		sopts = signing.NewOptions(
			signing.VerifySignature(SIGNATURE),
			signing.VerifyDigests(),
		)
		Must(signing.Apply(printer, nil, tcv, sopts))
		if tcv.GetDescriptor().GetSignatureIndex(SIGNATURE2) >= 0 {
			sopts = signing.NewOptions(
				signing.VerifySignature(SIGNATURE2),
				signing.VerifyDigests(),
			)
			Must(signing.Apply(printer, nil, tcv, sopts))
		}
	},
		Entry("modify component label", componentLabel),
		Entry("modify provider label", providerLabel),
		Entry("modify resource label", resourceLabel),

		Entry("merge component label", mergeComponentLabel),
		Entry("merge provider label", mergeProviderLabel),
		Entry("merge resource label", mergeResourceLabel),

		Entry("source signature", sourceSignature),
	)
})

func componentLabel(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	cv.GetDescriptor().Labels.Set(LABEL_VOL, "changed-comp-volatile")
	merged.Labels.Set(LABEL_VOL, "changed-comp-volatile")
}

func mergeComponentLabel(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	cv.GetDescriptor().Labels.Set(LABEL_VOL_NEW, "new-volatile")
	merged.Labels.Set(LABEL_VOL_NEW, "new-volatile")

	tcv.GetDescriptor().Labels.Set(LABEL_VOL_LOCAL, "local-volatile")
	merged.Labels.Set(LABEL_VOL_LOCAL, "local-volatile")
}

func providerLabel(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	cv.GetDescriptor().Provider.Labels.Set(LABEL_VOL, "changed-prov-volatile")
	merged.Provider.Labels.Set(LABEL_VOL, "changed-prov-volatile")
}

func mergeProviderLabel(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	cv.GetDescriptor().Provider.Labels.Set(LABEL_VOL_NEW, "new-volatile")
	merged.Provider.Labels.Set(LABEL_VOL_NEW, "new-volatile")

	tcv.GetDescriptor().Provider.Labels.Set(LABEL_VOL_LOCAL, "local-volatile")
	merged.Provider.Labels.Set(LABEL_VOL_LOCAL, "local-volatile")
}

func resourceLabel(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	rid := metav1.NewIdentity("testdata")
	ra := Must(cv.GetResource(rid))
	tr := NotNil(merged.GetResourceByIdentity(rid))

	ra.Meta().SetLabel(LABEL_VOL, "changed-resource-volatile")
	tr.SetLabel(LABEL_VOL, "changed-resource-volatile")
}

func mergeResourceLabel(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	rid := metav1.NewIdentity("testdata")
	tr := NotNil(merged.GetResourceByIdentity(rid))

	ra := NotNil(cv.GetDescriptor().GetResourceByIdentity(rid))
	ra.SetLabel(LABEL_VOL_NEW, "new-resource-volatile")
	tr.SetLabel(LABEL_VOL_NEW, "new-resource-volatile")

	ra = NotNil(tcv.GetDescriptor().GetResourceByIdentity(rid))
	ra.SetLabel(LABEL_VOL_LOCAL, "local-resource-volatile")
	tr.SetLabel(LABEL_VOL_LOCAL, "local-resource-volatile")
}

func sourceSignature(cv, tcv ocm.ComponentVersionAccess, merged *compdesc.ComponentDescriptor) {
	// sign in source
	sopts := signing.NewOptions(
		signing.Sign(signingattr.Get(cv.GetContext()).GetSigner(SIGN_ALGO), SIGNATURE2),
		signing.Update(), signing.VerifyDigests(),
	)
	spec, err := signing.Apply(common.NewPrinter(nil), nil, cv, sopts)
	ExpectWithOffset(1, err).To(Succeed())

	signatures := []compdesc.Signature{
		{
			Name:   SIGNATURE2,
			Digest: *spec,
			Signature: metav1.SignatureSpec{
				Algorithm: rsa.Algorithm,
				MediaType: rsa.MediaType,
				Value:     cv.GetDescriptor().Signatures[0].Signature.Value,
			},
		},
	}
	ExpectWithOffset(1, spec).To(Equal(&metav1.DigestSpec{
		HashAlgorithm:          "SHA-256",
		NormalisationAlgorithm: "jsonNormalisation/v1",
		Value:                  D_COMPA,
	}))
	ExpectWithOffset(1, cv.GetDescriptor().Signatures).To(ConsistOf(signatures))

	merged.Signatures = append(signatures, merged.Signatures...)
}
