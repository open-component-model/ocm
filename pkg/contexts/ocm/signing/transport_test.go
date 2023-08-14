// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"

	// . "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	tenv "github.com/open-component-model/ocm/pkg/env"
)

const (
	LABEL_SIG     = "non-volatile"
	LABEL_VOL     = "volatile"
	LABEL_VOL_NEW = "new-volatile"

	TARGET = "/tmp/target"
)

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
    sources: []
    version: v1
  meta:
    configuredSchemaVersion: v2
`, Substitutions{
		"DIGEST": D_TESTDATA,
	}))

	BeforeEach(func() {
		env = NewBuilder(tenv.NewEnvironment())

		env.RSAKeyPair(SIGNATURE, SIGNATURE2)

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
			})
		})

	})

	It("provides expected base component", func() {
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
		defer Close(cv, "cv")
		Expect(cv.GetDescriptor()).To(YAMLEqual(descData))
	})

	It("retransports after local signing", func() {
		target := Must(ctf.Create(env, accessobj.ACC_WRITABLE, TARGET, 0o700, env))
		defer Close(target, "target")

		repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
		defer Close(cv, "cv")

		printer, buf := common.NewBufferedPrinter()
		// transport
		handler := Must(standard.New())
		MustBeSuccessful(transfer.TransferVersion(printer, nil, cv, target, handler))

		var targetfinal finalizer.Finalizer
		defer Defer(targetfinal.Finalize, "target objects")

		tcv := Must(target.LookupComponentVersion(COMPONENTA, VERSION))
		targetfinal.Close(tcv, "tcv")
		Expect(tcv.GetDescriptor()).To(YAMLEqual(descData))

		descSigned := tcv.GetDescriptor().Copy()

		// sign in target
		sopts := signing.NewOptions(
			signing.Sign(signing.DefaultHandlerRegistry().GetSigner(SIGN_ALGO), SIGNATURE),
			signing.Update(), signing.VerifyDigests(),
		)
		spec := Must(signing.Apply(printer, nil, tcv, sopts))

		D_COMPA := "4dd928ad1d9d7d47a822f5d84bd16097188bb03bb99a059f78271583ee92f8b9"
		digSpec := &metav1.DigestSpec{
			HashAlgorithm:          "SHA-256",
			NormalisationAlgorithm: "jsonNormalisation/v1",
			Value:                  D_COMPA,
		}
		descSigned.Signatures = []compdesc.Signature{
			compdesc.Signature{
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
		MustBeSuccessful(targetfinal.Finalize())

		// change volatile data in origin
		rid := metav1.NewIdentity("testdata")
		ra := Must(cv.GetResource(rid))
		tr := Must(merged.GetResourceByIdentity(rid))

		ra.Meta().SetLabel(LABEL_VOL, "changed-resource-volatile")
		tr.SetLabel(LABEL_VOL, "changed-resource-volatile")
		cv.GetDescriptor().Labels.Set(LABEL_VOL, "changed-comp-volatile")
		merged.Labels.Set(LABEL_VOL, "changed-comp-volatile")
		cv.GetDescriptor().Provider.Labels.Set(LABEL_VOL, "changed-prov-volatile")
		merged.Provider.Labels.Set(LABEL_VOL, "changed-prov-volatile")

		// transfer changed volatile data
		buf.Reset()
		err := transfer.TransferVersion(printer, nil, cv, target, handler)
		fmt.Printf("%s\n", buf.String())
		MustBeSuccessful(err)

		tcv = Must(target.LookupComponentVersion(COMPONENTA, VERSION))
		targetfinal.Close(tcv, "tcv")
		Expect(tcv.GetDescriptor()).To(DeepEqual(merged))
	})
})
