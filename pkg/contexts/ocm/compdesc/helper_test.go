package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("helper", func() {

	It("should inject a new repository context if none is defined", func() {
		cd := &compdesc.ComponentDescriptor{}
		compdesc.DefaultComponent(cd)

		repoCtx := ocireg.NewRepositorySpec("example.com", nil)
		Expect(cd.AddRepositoryContext(repoCtx)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(1))

		Expect(cd.AddRepositoryContext(repoCtx)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(1))

		repoCtx2 := ocireg.NewRepositorySpec("example.com/dev", nil)
		Expect(cd.AddRepositoryContext(repoCtx2)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(2))
	})

	Context("resource selection", func() {
		cd := &compdesc.ComponentDescriptor{}

		r1v1 := compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "r1",
					Version: "v1",
					Labels: v1.Labels{
						v1.Label{
							Name:    "l1",
							Value:   []byte("\"labelvalue\""),
							Version: "v1",
							Signing: false,
						},
					},
				},
				Type:     "t1",
				Relation: v1.LocalRelation,
			},
		}
		r1v2 := compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "r1",
					Version: "v2",
				},
				Type:     "t1",
				Relation: v1.LocalRelation,
			},
		}
		r2v1 := compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "r2",
					Version: "v1",
					Labels: v1.Labels{
						v1.Label{
							Name:    "l1",
							Value:   []byte("\"othervalue\""),
							Version: "v1",
							Signing: false,
						},
					},
				},
				Type:     "t2",
				Relation: v1.LocalRelation,
			},
		}
		r3v2 := compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "r3",
					Version: "v2",
					Labels: v1.Labels{
						v1.Label{
							Name:    "l1",
							Value:   []byte("\"dummy\""),
							Version: "v2",
							Signing: false,
						},
						v1.Label{
							Name:    "l2",
							Value:   []byte("\"labelvalue\""),
							Version: "v2",
							Signing: true,
						},
						v1.Label{
							Name:    "l3",
							Value:   []byte("\"labelvalue\""),
							Version: "v3",
						},
					},
				},
				Type:     "t2",
				Relation: v1.LocalRelation,
			},
		}

		r4v3 := compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "r4",
					Version: "v3",
					ExtraIdentity: v1.Identity{
						"extra": "value",
						"other": "othervalue",
					},
				},
				Type:     "t3",
				Relation: v1.LocalRelation,
			},
		}

		cd.Resources = compdesc.Resources{
			r1v1,
			r1v2,
			r2v1,
			r3v2,
			r4v3,
		}

		Context("id selection", func() {
			It("selects by name", func() {
				res := Must(cd.GetResourcesByIdentitySelectors(compdesc.ByName("r1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.GetResourcesByIdentitySelectors(compdesc.ByVersion("v1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1})) // no r2v1: version nor part of identity
			})
		})

		Context("attr selection", func() {
			It("selects by name", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByName("r1")}))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByVersion("v1")}))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r2v1}))
			})
			It("selects by type", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByResourceType("t2")}))
				Expect(res).To(Equal(compdesc.Resources{r2v1, r3v2}))
			})

			It("selects by label name", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByLabelName("l1")}))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r2v1, r3v2}))
			})
			It("selects by label version", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByLabelVersion("v2")}))
				Expect(res).To(Equal(compdesc.Resources{r3v2}))
			})
			It("selects by label value", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByLabelValue("labelvalue")}))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r3v2}))
			})
			It("selects unrelated by label name and value", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByLabelName("l1"), compdesc.ByLabelValue("labelvalue")}))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r3v2})) // unrelated checks at resource level
			})
			It("selects related by label name and value", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByLabel(compdesc.ByLabelName("l1"), compdesc.ByLabelValue("labelvalue"))}))
				Expect(res).To(Equal(compdesc.Resources{r1v1})) // related checks at label level
			})
			It("selects by signed label", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.BySignedLabel()}))
				Expect(res).To(Equal(compdesc.Resources{r3v2}))
			})

			It("selects with extra identity", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.WithExtraIdentity("extra", "value")}))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
				res = Must(cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.WithExtraIdentity("extra", "value")}, nil))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))

			})
			It("selects none with wrong extra identity value", func() {
				_, err := cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.WithExtraIdentity("extra", "other")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.WithExtraIdentity("extra", "other")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})

			It("selects none with wrong extra identity key", func() {
				_, err := cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.WithExtraIdentity("extra2", "value")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.WithExtraIdentity("extra2", "value")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})

			It("selects by or", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.OrR(compdesc.ByName("r3"), compdesc.ByVersion("v2"))}))
				Expect(res).To(Equal(compdesc.Resources{r1v2, r3v2}))
			})

			It("selects by and", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.AndR(compdesc.ByName("r1"), compdesc.ByVersion("v1"))}))
				Expect(res).To(Equal(compdesc.Resources{r1v1}))
			})

			It("selects by negated selector", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.NotR(compdesc.ByName("r1"))}))
				Expect(res).To(Equal(compdesc.Resources{r2v1, r3v2, r4v3}))
			})

			It("selects by identity selector", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByIdentity("r4", "extra", "value", "other", "othervalue")}))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
				res = Must(cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.ByIdentity("r4", "extra", "value", "other", "othervalue")}, nil))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})
			It("selects none by identity selector with missing attribute", func() {
				_, err := cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByIdentity("r4", "extra", "value")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.ByIdentity("r4", "extra", "value")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})

			It("selects by partial identity selector", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByPartialIdentity("r4", "extra", "value", "other", "othervalue")}))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
				res = Must(cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.ByPartialIdentity("r4", "extra", "value", "other", "othervalue")}, nil))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})
			It("selects by partial identity selector with partial attributes", func() {
				res := Must(cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByPartialIdentity("r4", "extra", "value")}))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
				res = Must(cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.ByPartialIdentity("r4", "extra", "value")}, nil))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})
			It("selects none by partial identity selector with missing attribute", func() {
				_, err := cd.GetResourcesBySelectors(nil, []compdesc.ResourceSelector{compdesc.ByIdentity("r4", "extra", "value", "dummy", "dummy")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetResourcesBySelectors([]compdesc.IdentitySelector{compdesc.ByIdentity("r4", "extra", "value", "dummy", "dummy")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})
		})

		Context("select labels", func() {
			It("selects label by name", func() {
				res := Must(compdesc.SelectLabels(r3v2.Labels, compdesc.ByLabelName("l1")))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[0]}))
			})
			It("selects signed labels", func() {
				res := Must(compdesc.SelectLabels(r3v2.Labels, compdesc.BySignedLabel()))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[1]}))
			})
			It("selects no signed labels", func() {
				res := Must(compdesc.SelectLabels(r1v2.Labels, compdesc.BySignedLabel()))
				Expect(res).To(Equal(v1.Labels{}))
			})

			It("selects labels by or", func() {
				res := Must(compdesc.SelectLabels(r3v2.Labels, compdesc.OrL(compdesc.ByLabelName("l1"), compdesc.ByLabelVersion("v3"))))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[0], r3v2.Labels[2]}))
			})

			It("selects labels by and", func() {
				res := Must(compdesc.SelectLabels(r3v2.Labels, compdesc.AndL(compdesc.ByLabelValue("labelvalue"), compdesc.ByLabelVersion("v2"))))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[1]}))
			})

			It("selects labels by negated selector", func() {
				res := Must(compdesc.SelectLabels(r3v2.Labels, compdesc.NotL(compdesc.ByLabelValue("labelvalue"))))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[0]}))
			})
		})
	})

	Context("reference selection", func() {
		cd := &compdesc.ComponentDescriptor{}

		r1v1 := compdesc.ComponentReference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "r1",
				Version: "v1",
				Labels: v1.Labels{
					v1.Label{
						Name:    "l1",
						Value:   []byte("\"labelvalue\""),
						Version: "v1",
						Signing: false,
					},
				},
			},
			ComponentName: "c1",
		}
		r1v2 := compdesc.ComponentReference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "r1",
				Version: "v2",
			},
			ComponentName: "c1",
		}
		r2v1 := compdesc.ComponentReference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "r2",
				Version: "v1",
				Labels: v1.Labels{
					v1.Label{
						Name:    "l1",
						Value:   []byte("\"othervalue\""),
						Version: "v1",
						Signing: false,
					},
				},
			},
			ComponentName: "c2",
		}
		r3v2 := compdesc.ComponentReference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "r3",
				Version: "v2",
				Labels: v1.Labels{
					v1.Label{
						Name:    "l1",
						Value:   []byte("\"dummy\""),
						Version: "v2",
						Signing: false,
					},
					v1.Label{
						Name:    "l2",
						Value:   []byte("\"labelvalue\""),
						Version: "v2",
						Signing: true,
					},
					v1.Label{
						Name:    "l3",
						Value:   []byte("\"labelvalue\""),
						Version: "v3",
					},
				},
			},
			ComponentName: "c3",
		}

		r4v3 := compdesc.ComponentReference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "r4",
				Version: "v3",
				ExtraIdentity: v1.Identity{
					"extra": "value",
					"other": "othervalue",
				},
			},
			ComponentName: "c4",
		}

		cd.References = compdesc.References{
			r1v1,
			r1v2,
			r2v1,
			r3v2,
			r4v3,
		}

		Context("id selection", func() {
			It("selects by name", func() {
				res := Must(cd.GetReferencesByIdentitySelectors(compdesc.ByName("r1")))
				Expect(res).To(Equal(compdesc.References{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.GetReferencesByIdentitySelectors(compdesc.ByVersion("v1")))
				Expect(res).To(Equal(compdesc.References{r1v1})) // no r2v1: version nor part of identity
			})
		})

		Context("attr selection", func() {
			It("selects by name", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByName("r1")}))
				Expect(res).To(Equal(compdesc.References{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByVersion("v1")}))
				Expect(res).To(Equal(compdesc.References{r1v1, r2v1}))
			})
			It("selects by component", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByComponent("c2")}))
				Expect(res).To(Equal(compdesc.References{r2v1}))
			})

			It("selects by label name", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByLabelName("l1")}))
				Expect(res).To(Equal(compdesc.References{r1v1, r2v1, r3v2}))
			})
			It("selects by label version", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByLabelVersion("v2")}))
				Expect(res).To(Equal(compdesc.References{r3v2}))
			})
			It("selects by label value", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByLabelValue("labelvalue")}))
				Expect(res).To(Equal(compdesc.References{r1v1, r3v2}))
			})
			It("selects unrelated by label name and value", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByLabelName("l1"), compdesc.ByLabelValue("labelvalue")}))
				Expect(res).To(Equal(compdesc.References{r1v1, r3v2})) // unrelated checks at resource level
			})
			It("selects related by label name and value", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByLabel(compdesc.ByLabelName("l1"), compdesc.ByLabelValue("labelvalue"))}))
				Expect(res).To(Equal(compdesc.References{r1v1})) // related checks at label level
			})
			It("selects by signed label", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.BySignedLabel()}))
				Expect(res).To(Equal(compdesc.References{r3v2}))
			})

			It("selects with extra identity", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.WithExtraIdentity("extra", "value")}))
				Expect(res).To(Equal(compdesc.References{r4v3}))
				res = Must(cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.WithExtraIdentity("extra", "value")}, nil))
				Expect(res).To(Equal(compdesc.References{r4v3}))

			})
			It("selects none with wrong extra identity value", func() {
				_, err := cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.WithExtraIdentity("extra", "other")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.WithExtraIdentity("extra", "other")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})

			It("selects none with wrong extra identity key", func() {
				_, err := cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.WithExtraIdentity("extra2", "value")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.WithExtraIdentity("extra2", "value")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})

			It("selects by or", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.Or[compdesc.ReferenceSelector](compdesc.ByName("r3"), compdesc.ByVersion("v2"))}))
				Expect(res).To(Equal(compdesc.References{r1v2, r3v2}))
			})

			It("selects by and", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.AndC(compdesc.ByName("r1"), compdesc.ByVersion("v1"))}))
				Expect(res).To(Equal(compdesc.References{r1v1}))
			})

			It("selects by negated selector", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.NotC(compdesc.ByName("r1"))}))
				Expect(res).To(Equal(compdesc.References{r2v1, r3v2, r4v3}))
			})

			It("selects by identity selector", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByIdentity("r4", "extra", "value", "other", "othervalue")}))
				Expect(res).To(Equal(compdesc.References{r4v3}))
				res = Must(cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.ByIdentity("r4", "extra", "value", "other", "othervalue")}, nil))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})
			It("selects none by identity selector with missing attribute", func() {
				_, err := cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByIdentity("r4", "extra", "value")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.ByIdentity("r4", "extra", "value")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})

			It("selects by partial identity selector", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByPartialIdentity("r4", "extra", "value", "other", "othervalue")}))
				Expect(res).To(Equal(compdesc.References{r4v3}))
				res = Must(cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.ByPartialIdentity("r4", "extra", "value", "other", "othervalue")}, nil))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})
			It("selects by partial identity selector with partial attributes", func() {
				res := Must(cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByPartialIdentity("r4", "extra", "value")}))
				Expect(res).To(Equal(compdesc.References{r4v3}))
				res = Must(cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.ByPartialIdentity("r4", "extra", "value")}, nil))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})
			It("selects none by partial identity selector with missing attribute", func() {
				_, err := cd.GetReferencesBySelectors(nil, []compdesc.ReferenceSelector{compdesc.ByIdentity("r4", "extra", "value", "dummy", "dummy")})
				Expect(err).To(MatchError(compdesc.NotFound))
				_, err = cd.GetReferencesBySelectors([]compdesc.IdentitySelector{compdesc.ByIdentity("r4", "extra", "value", "dummy", "dummy")}, nil)
				Expect(err).To(MatchError(compdesc.NotFound))
			})
		})
	})
})
