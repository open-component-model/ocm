package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/compdesc/equivalent"
	. "ocm.software/ocm/api/ocm/compdesc/equivalent/testhelper"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
)

var _ = Describe("equivalence", func() {
	var labels v1.Labels
	var modtime *v1.Timestamp

	_ = modtime

	BeforeEach(func() {
		labels.Clear()
		labels.Set("label1", "value1", v1.WithSigning())
		labels.Set("label3", "value3")
	})

	Context("element meta", func() {
		var a, b *compdesc.ElementMeta

		BeforeEach(func() {
			a = &compdesc.ElementMeta{
				Name:          "r1",
				Version:       "v1",
				ExtraIdentity: v1.NewExtraIdentity("extra", "extra"),
				Labels:        labels.Copy(),
			}
			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
		})

		It("handles name change", func() {
			b.Name = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles version change", func() {
			b.Version = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles extra id change", func() {
			b.ExtraIdentity["X"] = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles non-volatile label change", func() {
			b.Labels[0].Value = []byte("X")
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles volatile label change", func() {
			b.Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
		})
	})

	Context("resource", func() {
		var a, b *compdesc.Resource

		BeforeEach(func() {
			a = &compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:    "r1",
						Version: "v1",
						Labels:  labels.Copy(),
					},
					Type:     "test",
					Relation: v1.LocalRelation,
					Digest: &v1.DigestSpec{
						HashAlgorithm:          "hash",
						NormalisationAlgorithm: "norm",
						Value:                  "x",
					},
				},
				Access: localblob.New("test", "test", "test", nil),
			}
			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles volatile meta change", func() {
			b.Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles non-volatile meta change", func() {
			b.Labels[0].Value = []byte("X")
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles type change", func() {
			b.Type = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles version change", func() {
			b.Version = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles relation change", func() {
			b.Relation = compdesc.ExternalRelation
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles access change", func() {
			b.Access = ociartifact.New("test")
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles undetectable digest", func() {
			b.Digest = nil
			CheckNotDetectable(a.Equivalent(b))
			CheckNotDetectable(b.Equivalent(a))
			a.Digest = nil
			CheckNotDetectable(a.Equivalent(b))
			CheckNotDetectable(b.Equivalent(a))
		})

		It("handles different digest", func() {
			b.Digest.Value = "X"
			CheckNotArtifactEqual(a.Equivalent(b))
			CheckNotArtifactEqual(b.Equivalent(a))
		})

		It("handles none access", func() {
			a.Digest = nil
			a.Access = none.New()
			b.Digest = nil
			b.Access = none.New()
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles no-digest", func() {
			a.Digest = compdesc.NewExcludeFromSignatureDigest()
			b.Digest = compdesc.NewExcludeFromSignatureDigest()
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles no-digest on one side", func() {
			b.Digest = compdesc.NewExcludeFromSignatureDigest()
			CheckNotArtifactEqual(a.Equivalent(b))
			CheckNotArtifactEqual(b.Equivalent(a))
		})
	})

	Context("resources", func() {
		var a, b compdesc.Resources

		BeforeEach(func() {
			a = compdesc.Resources{
				compdesc.Resource{
					ResourceMeta: compdesc.ResourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:   "r1",
							Labels: labels.Copy(),
						},
						Type:     "test",
						Relation: v1.LocalRelation,
						Digest: &v1.DigestSpec{
							HashAlgorithm:          "hash",
							NormalisationAlgorithm: "norm",
							Value:                  "x",
						},
					},
					Access: localblob.New("test1", "test1", "test", nil),
				},
				compdesc.Resource{
					ResourceMeta: compdesc.ResourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:          "r2",
							ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
							Labels:        labels.Copy(),
						},
						Type:     "test",
						Relation: v1.LocalRelation,
						Digest: &v1.DigestSpec{
							HashAlgorithm:          "hash",
							NormalisationAlgorithm: "norm",
							Value:                  "y",
						},
					},
					Access: localblob.New("test2", "test2", "test", nil),
				},
			}

			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		/* order is now relevant, because it is relevant for the normalization
		It("handles order change", func() {
			b[0], b[1] = b[1], b[0]
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})
		*/

		It("handles volatile change", func() {
			b[0].Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles non-volatile attr change", func() {
			b[0].Type = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles different digest", func() {
			b[0].Digest.Value = "X"
			CheckNotArtifactEqual(a.Equivalent(b))
			CheckNotArtifactEqual(b.Equivalent(a))
		})

		It("handles undetectable digest", func() {
			b[0].Digest = nil
			CheckNotDetectable(a.Equivalent(b))
			CheckNotDetectable(b.Equivalent(a))
			a[0].Digest = nil
			CheckNotDetectable(a.Equivalent(b))
			CheckNotDetectable(b.Equivalent(a))
		})

		It("handles additional entry", func() {
			b = append(b, compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "r3",
						ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
						Labels:        labels.Copy(),
					},
					Type:     "test",
					Relation: v1.LocalRelation,
					Digest: &v1.DigestSpec{
						HashAlgorithm:          "hash",
						NormalisationAlgorithm: "norm",
						Value:                  "z",
					},
				},
				Access: localblob.New("test3", "test3", "test", nil),
			})
			Expect(a.Equivalent(b)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(true))))
			Expect(b.Equivalent(a)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(true))))
		})

		It("handles additional entry without any other meta data", func() {
			b = append(b, compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "r3",
						ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
					},
					Type:     "test",
					Relation: v1.LocalRelation,
					Digest: &v1.DigestSpec{
						HashAlgorithm:          "hash",
						NormalisationAlgorithm: "norm",
						Value:                  "z",
					},
				},
				Access: localblob.New("test3", "test3", "test", nil),
			})
			Expect(a.Equivalent(b)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(true))))
			Expect(b.Equivalent(a)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(true))))
		})

		It("handles additional entry without any other meta data and digest", func() {
			b = append(b, compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "r3",
						ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
					},
					Type:     "test",
					Relation: v1.LocalRelation,
				},
				Access: localblob.New("test3", "test3", "test", nil),
			})
			Expect(a.Equivalent(b)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(false))))
			Expect(b.Equivalent(a)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(false))))
		})

		It("handles additional non-digest entry", func() {
			b = append(b, compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "r3",
						ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
						Labels:        labels.Copy(),
					},
					Type:     "test",
					Relation: v1.LocalRelation,
					Digest:   compdesc.NewExcludeFromSignatureDigest(),
				},
				Access: localblob.New("test3", "test3", "test", nil),
			})
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))

			a = append(a, *b[2].Copy())
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))

			b[2].Access = localblob.New("test4", "test4", "test", nil)
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))

			b[2].Type = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})
	})

	Context("source", func() {
		var a, b *compdesc.Source

		BeforeEach(func() {
			a = &compdesc.Source{
				SourceMeta: compdesc.SourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:   "s1",
						Labels: labels.Copy(),
					},
					Type: "test",
				},
				Access: localblob.New("test", "test", "test", nil),
			}
			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles volatile meta change", func() {
			b.Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles non-volatile meta change", func() {
			b.Labels[0].Value = []byte("X")
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles version change", func() {
			b.Version = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles type change", func() {
			b.Type = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		// actually content not relevant
		It("handles access change", func() {
			b.Access = ociartifact.New("test")
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})
	})

	Context("sources", func() {
		var a, b compdesc.Sources

		BeforeEach(func() {
			a = compdesc.Sources{
				compdesc.Source{
					SourceMeta: compdesc.SourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:   "s1",
							Labels: labels.Copy(),
						},
						Type: "test",
					},
					Access: localblob.New("test1", "test1", "test", nil),
				},
				compdesc.Source{
					SourceMeta: compdesc.SourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:          "s2",
							ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
							Labels:        labels.Copy(),
						},
						Type: "test",
					},
					Access: localblob.New("test2", "test2", "test", nil),
				},
			}

			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		/* order is now relevant, because it is relevant for the normalization
		It("handles order change", func() {
			b[0], b[1] = b[1], b[0]
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})
		*/

		It("handles volatile change", func() {
			b[0].Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles non-volatile attr change", func() {
			b[0].Type = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles additional entry", func() {
			b = append(b, compdesc.Source{
				SourceMeta: compdesc.SourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "s3",
						ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
						Labels:        labels.Copy(),
					},
					Type: "test",
				},
				Access: localblob.New("test3", "test3", "test", nil),
			})
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles additional entry without any other meta data", func() {
			b = append(b, compdesc.Source{
				SourceMeta: compdesc.SourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "s3",
						ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
					},
					Type: "test",
				},
				Access: localblob.New("test3", "test3", "test", nil),
			})
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})
	})

	Context("reference", func() {
		var a, b *compdesc.Reference

		BeforeEach(func() {
			a = &compdesc.Reference{
				ElementMeta: compdesc.ElementMeta{
					Name:    "r1",
					Version: "v1",
					Labels:  labels.Copy(),
				},
				ComponentName: "test",
			}
			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles volatile meta change", func() {
			b.Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles non-volatile meta change", func() {
			b.Labels[0].Value = []byte("X")
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles version change", func() {
			b.Version = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles component change", func() {
			b.ComponentName = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})
	})

	Context("references", func() {
		var a, b compdesc.References

		BeforeEach(func() {
			a = compdesc.References{
				compdesc.Reference{
					ElementMeta: compdesc.ElementMeta{
						Name:    "s1",
						Version: "v1",
						Labels:  labels.Copy(),
					},
					ComponentName: "c1",
				},
				compdesc.Reference{
					ElementMeta: compdesc.ElementMeta{
						Name:    "s2",
						Version: "v1",
						Labels:  labels.Copy(),
					},
					ComponentName: "c2",
				},
			}

			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		/* order is now relevant, because it is relevant for the normalization
		It("handles order change", func() {
			b[0], b[1] = b[1], b[0]
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})
		*/

		It("handles volatile change", func() {
			b[0].Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles component change", func() {
			b[0].ComponentName = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles additional entry", func() {
			b = append(b, compdesc.Reference{
				ElementMeta: compdesc.ElementMeta{
					Name:          "s3",
					ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
					Labels:        labels.Copy(),
				},
				ComponentName: "c3",
			})
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles additional entry without any other metadata", func() {
			b = append(b, compdesc.Reference{
				ElementMeta: compdesc.ElementMeta{
					Name:          "s3",
					ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
				},
				ComponentName: "c3",
			})
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles additional entry without any other metadata", func() {
			b = append(b, compdesc.Reference{
				ElementMeta: compdesc.ElementMeta{
					Name:          "s3",
					ExtraIdentity: compdesc.NewExtraIdentity("platform", "linux"),
				},
				ComponentName: "c3",
				Digest: &v1.DigestSpec{
					HashAlgorithm:          "hash",
					NormalisationAlgorithm: "norm",
					Value:                  "z",
				},
			})
			Expect(a.Equivalent(b)).To(Equal(equivalent.StateNotLocalHashEqual().Apply(equivalent.StateNotArtifactEqual(true))))
		})
	})

	Context("signatures", func() {
		var s1, s2 *v1.Signature

		BeforeEach(func() {
			s1 = &v1.Signature{
				Name: "sig",
				Digest: v1.DigestSpec{
					HashAlgorithm:          "hash",
					NormalisationAlgorithm: "norm",
					Value:                  "H",
				},
				Signature: v1.SignatureSpec{
					Algorithm: "sign",
					Value:     "S",
					MediaType: "M",
					Issuer:    "issuer",
				},
			}
			s2 = s1.Copy()
			s2.Name = "other"
		})

		It("handles equal", func() {
			a := compdesc.Signatures{*s1}
			b := compdesc.Signatures{*s1.Copy()}
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles diff", func() {
			a := compdesc.Signatures{*s1}
			b := compdesc.Signatures{}
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})
	})
})
