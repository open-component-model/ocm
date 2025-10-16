package transfer_test

import (
	"github.com/go-test/deep"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/tools/transfer/internal"
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/maplistmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
)

func TouchLabels(s, d, e *metav1.Labels) {
	// overwrite old one
	s.SetValue("volatile", "modified")

	// keep local one
	d.Set("local-volatile", "local")

	// add new one
	s.Set("new-volatile", "new")

	*e = s.Copy()
	e.SetDef("local-volatile", d.GetDef("local-volatile"))
}

var _ = Describe("basic merge operations for transport", func() {
	Context("merging labels", func() {
		var src metav1.Labels
		var dst metav1.Labels

		BeforeEach(func() {
			src.Set("add-signed", "new signed label", metav1.WithSigning())
			src.Set("add-unsigned", "new unsigned label")
			src.Set("old-signed", "old signed label", metav1.WithSigning())
			src.Set("old-unsigned", "old signed label")

			dst.Set("old-signed", "signed label", metav1.WithSigning())
			dst.Set("old-unsigned", "signed label")
			dst.Set("new-signed", "signed label", metav1.WithSigning())
			dst.Set("new-unsigned", "unsigned label")
		})

		It("add signed additional ones", func() {
			res := dst.Copy()
			MustBeSuccessful(internal.MergeLabels(hpi.Log, ocm.DefaultContext(), src, &res))
			Expect(res).To(ConsistOf(append(dst, *src.GetDef("add-unsigned"))))
		})

		It("merges with list merger", func() {
			src.Clear()
			src.Set("test", maplistmerge.Value{
				map[string]interface{}{
					"name": "e1",
					"data": "data1",
				},
				map[string]interface{}{
					"name": "e2",
					"data": "data2",
				},
			})

			dst.Set("test", maplistmerge.Value{
				map[string]interface{}{
					"name": "e1",
					"data": "old",
				},
				map[string]interface{}{
					"name": "e3",
					"data": "new",
				},
			},
				metav1.WithMerging(maplistmerge.ALGORITHM, maplistmerge.NewConfig("", maplistmerge.MODE_LOCAL)))
			res := dst.Copy()
			MustBeSuccessful(internal.MergeLabels(hpi.Log, ocm.DefaultContext(), src, &res))

			var v maplistmerge.Value
			Expect(res.GetValue("test", &v)).To(BeTrue())
			Expect(v).To(DeepEqual(maplistmerge.Value{
				map[string]interface{}{
					"name": "e1",
					"data": "data1",
				},
				map[string]interface{}{
					"name": "e3",
					"data": "new",
				},
				map[string]interface{}{
					"name": "e2",
					"data": "data2",
				},
			}))
		})
	})

	////////////////////////////////////////////////////////////////////////////
	Context("merging signatures", func() {
		var src metav1.Signatures
		var dst metav1.Signatures

		BeforeEach(func() {
			src.Set(metav1.Signature{
				Name: "add",
				Digest: metav1.DigestSpec{
					Value: "add",
				},
				Signature: metav1.SignatureSpec{
					Value: "add-sig",
				},
			})
			src.Set(metav1.Signature{
				Name: "old",
				Digest: metav1.DigestSpec{
					Value: "old",
				},
				Signature: metav1.SignatureSpec{
					Value: "old-sig",
				},
			})

			dst.Set(metav1.Signature{
				Name: "new",
				Digest: metav1.DigestSpec{
					Value: "new",
				},
				Signature: metav1.SignatureSpec{
					Value: "new-sig",
				},
			})
			dst.Set(metav1.Signature{
				Name: "old",
				Digest: metav1.DigestSpec{
					Value: "old",
				},
				Signature: metav1.SignatureSpec{
					Value: "mod-sig",
				},
			})
		})

		It("add signed additional ones", func() {
			res := dst.Copy()
			MustBeSuccessful(internal.MergeSignatures(src, &res))
			Expect(res).To(ConsistOf(append(dst, *src.GetByName("add"))))
		})
	})

	////////////////////////////////////////////////////////////////////////////
	Context("merging component descriptors", func() {
		var src *compdesc.ComponentDescriptor
		var dst *compdesc.ComponentDescriptor

		BeforeEach(func() {
			labels := metav1.Labels{}
			labels.Set("volatile", "value1", metav1.WithVersion("v1"))
			labels.Set("non-volatile", "value2", metav1.WithVersion("v1"), metav1.WithSigning())

			src = &compdesc.ComponentDescriptor{
				Metadata: compdesc.Metadata{
					ConfiguredVersion: v2.SchemaVersion,
				},
				ComponentSpec: compdesc.ComponentSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:    "acme.org/test",
						Version: "v1.0.0",
						Labels:  labels.Copy(),
						Provider: metav1.Provider{
							Name:   "acme.org",
							Labels: labels.Copy(),
						},
						CreationTime: metav1.NewTimestampP(),
					},
					Sources: compdesc.Sources{
						compdesc.Source{
							SourceMeta: compdesc.SourceMeta{
								ElementMeta: compdesc.ElementMeta{
									Name:    "src1",
									Version: "v1.0.0",
									Labels:  labels.Copy(),
								},
								Type: "",
							},
							Access: nil,
						},
						compdesc.Source{
							SourceMeta: compdesc.SourceMeta{
								ElementMeta: compdesc.ElementMeta{
									Name:    "src2",
									Version: "v1.0.0",
									Labels:  labels.Copy(),
								},
								Type: "",
							},
							Access: nil,
						},
					},
					References: compdesc.References{
						compdesc.Reference{
							ElementMeta: compdesc.ElementMeta{
								Name:    "ref1",
								Version: "v1.0.0",
								Labels:  labels.Copy(),
							},
							ComponentName: "acme.org/ref",
						},
					},
					Resources: compdesc.Resources{
						compdesc.Resource{
							ResourceMeta: compdesc.ResourceMeta{
								ElementMeta: compdesc.ElementMeta{
									Name:    "rsc1",
									Version: "v1.0.0",
									Labels:  labels.Copy(),
								},
								Type: "",
								Digest: &metav1.DigestSpec{
									HashAlgorithm:          "alg",
									NormalisationAlgorithm: "norm",
									Value:                  "digest1",
								},
							},
							Access: ociartifact.New("ghcr.io/acme/test1"),
						},
						compdesc.Resource{
							ResourceMeta: compdesc.ResourceMeta{
								ElementMeta: compdesc.ElementMeta{
									Name:    "rsc2",
									Version: "v1.0.0",
									Labels:  labels.Copy(),
								},
								Type: "",
								Digest: &metav1.DigestSpec{
									HashAlgorithm:          "alg",
									NormalisationAlgorithm: "norm",
									Value:                  "digest2",
								},
							},
							Access: ociartifact.New("ghcr.io/acme/test2"),
						},
					},
				},
				Signatures: metav1.Signatures{
					metav1.Signature{
						Name: "acme.org",
						Digest: metav1.DigestSpec{
							HashAlgorithm:          "alg",
							NormalisationAlgorithm: "norm",
							Value:                  "digest",
						},
						Signature: metav1.SignatureSpec{
							Algorithm: "alg",
							Value:     "signature",
							MediaType: "bla",
							Issuer:    "acme.org",
						},
					},
				},
			}
			dst = src.Copy()
		})

		It("merges equal", func() {
			n := Must(internal.PrepareDescriptor(hpi.Log, ocm.DefaultContext(), src, dst))
			diff := deep.Equal(n, dst)
			Expect(diff).To(BeEmpty())
		})

		It("merges signatures", func() {
			s := src.Copy()
			d := dst.Copy()

			// overwrite old one
			s.Signatures[0] = metav1.Signature{
				Name: "acme.org",
				Digest: metav1.DigestSpec{
					HashAlgorithm:          "alg",
					NormalisationAlgorithm: "norm",
					Value:                  "digest",
				},
				Signature: metav1.SignatureSpec{
					Algorithm: "algmod",
					Value:     "signaturemod",
					MediaType: "bla",
					Issuer:    "acme.org",
				},
			}

			// keep local one
			d.Signatures = append(d.Signatures,
				metav1.Signature{
					Name: "local",
					Digest: metav1.DigestSpec{
						HashAlgorithm:          "alglocal",
						NormalisationAlgorithm: "normlocal",
						Value:                  "digestlocal",
					},
					Signature: metav1.SignatureSpec{
						Algorithm: "alglocal",
						Value:     "signaturelocal",
						MediaType: "bla",
						Issuer:    "local",
					},
				},
			)

			// add new one
			s.Signatures = append(s.Signatures,
				metav1.Signature{
					Name: "new",
					Digest: metav1.DigestSpec{
						HashAlgorithm:          "algnew",
						NormalisationAlgorithm: "normnew",
						Value:                  "digestnew",
					},
					Signature: metav1.SignatureSpec{
						Algorithm: "algnew",
						Value:     "signaturenew",
						MediaType: "bla",
						Issuer:    "new",
					},
				},
			)

			e := d.Copy()
			e.Signatures[0] = s.Signatures[0]
			e.Signatures = append(s.Signatures.Copy(), d.Signatures[1])

			n := Must(internal.PrepareDescriptor(hpi.Log, nil, s, d))
			Expect(n).To(DeepEqual(e))
		})

		It("merges provider", func() {
			s := src.Copy()
			d := dst.Copy()
			e := d.Copy()
			TouchLabels(&s.Provider.Labels, &d.Provider.Labels, &e.Provider.Labels)

			n := Must(internal.PrepareDescriptor(hpi.Log, nil, s, d))
			Expect(n).To(DeepEqual(e))
		})

		It("merges component labels", func() {
			s := src.Copy()
			d := dst.Copy()
			e := d.Copy()
			TouchLabels(&s.Labels, &d.Labels, &e.Labels)

			n := Must(internal.PrepareDescriptor(hpi.Log, nil, s, d))
			Expect(n).To(DeepEqual(e))
		})

		It("merges resources", func() {
			s := src.Copy()
			d := dst.Copy()
			d.Resources[1].Access = localblob.New("local1", "", artdesc.MediaTypeImageManifest, nil)
			e := d.Copy()
			TouchLabels(&s.Resources[0].Labels, &d.Resources[0].Labels, &e.Resources[0].Labels)

			n := Must(internal.PrepareDescriptor(hpi.Log, nil, s, d))
			Expect(n).To(DeepEqual(e))
		})

		It("merges sources", func() {
			s := src.Copy()
			d := dst.Copy()
			e := d.Copy()
			TouchLabels(&s.Sources[0].Labels, &d.Sources[0].Labels, &e.Sources[0].Labels)

			n := Must(internal.PrepareDescriptor(hpi.Log, nil, s, d))
			Expect(n).To(DeepEqual(e))
		})

		It("merges references", func() {
			s := src.Copy()
			d := dst.Copy()
			e := d.Copy()
			TouchLabels(&s.References[0].Labels, &d.References[0].Labels, &e.References[0].Labels)

			n := Must(internal.PrepareDescriptor(hpi.Log, nil, s, d))
			Expect(n).To(DeepEqual(e))
		})
	})
})
