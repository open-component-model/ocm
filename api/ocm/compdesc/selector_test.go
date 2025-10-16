package compdesc_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/selectors/labelsel"
	"ocm.software/ocm/api/ocm/selectors/refsel"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
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
				res := Must(cd.SelectResources(rscsel.Name("r1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.SelectResources(rscsel.Version("v1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r2v1}))
			})
			It("selects by version", func() {
				res := Must(cd.SelectResources(rscsel.PartialIdentityByKeyPairs("version", "v1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1})) // no r2v1: version not part of identity
			})
		})

		Context("attr selection", func() {
			It("selects by name", func() {
				res := Must(cd.SelectResources(rscsel.Name("r1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.SelectResources(rscsel.Version("v1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r2v1}))
			})
			It("selects by type", func() {
				res := Must(cd.SelectResources(rscsel.ResourceType("t2")))
				Expect(res).To(Equal(compdesc.Resources{r2v1, r3v2}))
			})

			It("selects by label name", func() {
				res := Must(cd.SelectResources(rscsel.LabelName("l1")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r2v1, r3v2}))
			})
			It("selects by label version", func() {
				res := Must(cd.SelectResources(rscsel.LabelVersion("v2")))
				Expect(res).To(Equal(compdesc.Resources{r3v2}))
			})
			It("selects by label value", func() {
				res := Must(cd.SelectResources(rscsel.LabelValue("labelvalue")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r3v2}))
			})
			It("selects unrelated by label name and value", func() {
				res := Must(cd.SelectResources(rscsel.LabelName("l1"), rscsel.LabelValue("labelvalue")))
				Expect(res).To(Equal(compdesc.Resources{r1v1, r3v2})) // unrelated checks at resource level
			})
			It("selects related by label name and value", func() {
				res := Must(cd.SelectResources(rscsel.Label(labelsel.Name("l1"), labelsel.Value("labelvalue"))))
				Expect(res).To(Equal(compdesc.Resources{r1v1})) // related checks at label level
			})
			It("selects by signed label", func() {
				res := Must(cd.SelectResources(labelsel.Signed()))
				Expect(res).To(Equal(compdesc.Resources{r3v2}))
			})

			It("selects with extra identity", func() {
				res := Must(cd.SelectResources(rscsel.ExtraIdentityByKeyPairs("extra", "value", "other", "othervalue")))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})

			It("selects none with wrong extra identity value", func() {
				res := Must(cd.SelectResources(rscsel.ExtraIdentityByKeyPairs("extra", "other")))
				Expect(res).To(BeEmpty())
			})

			It("selects none with wrong extra identity key", func() {
				res := Must(cd.SelectResources(rscsel.ExtraIdentityByKeyPairs("extra2", "value")))
				Expect(res).To(BeEmpty())
			})

			It("selects by or", func() {
				res := Must(cd.SelectResources(rscsel.Or(rscsel.Name("r3"), rscsel.Version("v2"))))
				Expect(res).To(Equal(compdesc.Resources{r1v2, r3v2}))
			})

			It("selects by and", func() {
				res := Must(cd.SelectResources(rscsel.And(rscsel.Name("r1"), rscsel.Version("v1"))))
				Expect(res).To(Equal(compdesc.Resources{r1v1}))
			})

			It("selects by negated selector", func() {
				res := Must(cd.SelectResources(rscsel.Not(rscsel.Name("r1"))))
				Expect(res).To(Equal(compdesc.Resources{r2v1, r3v2, r4v3}))
			})

			It("selects by identity selector", func() {
				res := Must(cd.SelectResources(rscsel.IdentityByKeyPairs("r4", "extra", "value", "other", "othervalue")))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})

			It("selects none by identity selector with missing attribute", func() {
				res := Must(cd.SelectResources(rscsel.IdentityByKeyPairs("r4", "extra", "value")))
				Expect(res).To(BeEmpty())
			})

			It("selects by partial identity selector", func() {
				res := Must(cd.SelectResources(rscsel.PartialIdentityByKeyPairs("name", "r4", "extra", "value", "other", "othervalue")))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})

			It("selects by partial identity selector with partial attributes", func() {
				res := Must(cd.SelectResources(rscsel.PartialIdentityByKeyPairs("extra", "value")))
				Expect(res).To(Equal(compdesc.Resources{r4v3}))
			})

			It("selects none by partial identity selector with missing attribute", func() {
				res := Must(cd.SelectResources(rscsel.IdentityByKeyPairs("r4", "extra", "value", "dummy", "dummy")))
				Expect(res).To(BeEmpty())
			})
		})

		Context("select labels", func() {
			It("selects label by name", func() {
				res := Must(labelsel.Select(r3v2.Labels, labelsel.Name("l1")))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[0]}))
			})

			It("selects signed labels", func() {
				res := Must(labelsel.Select(r3v2.Labels, labelsel.Signed()))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[1]}))
			})

			It("selects no signed labels", func() {
				res := Must(labelsel.Select(r1v2.Labels, labelsel.Signed()))
				Expect(res).To(Equal(v1.Labels{}))
			})

			It("selects labels by or", func() {
				res := Must(labelsel.Select(r3v2.Labels, labelsel.Or(labelsel.Name("l1"), labelsel.Version("v3"))))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[0], r3v2.Labels[2]}))
			})

			It("selects labels by and", func() {
				res := Must(labelsel.Select(r3v2.Labels, labelsel.And(labelsel.Value("labelvalue"), labelsel.Version("v2"))))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[1]}))
			})

			It("selects labels by negated selector", func() {
				res := Must(labelsel.Select(r3v2.Labels, labelsel.Not(labelsel.Value("labelvalue"))))
				Expect(res).To(Equal(v1.Labels{r3v2.Labels[0]}))
			})
		})
	})

	Context("reference selection", func() {
		cd := &compdesc.ComponentDescriptor{}

		r1v1 := compdesc.Reference{
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
		r1v2 := compdesc.Reference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "r1",
				Version: "v2",
			},
			ComponentName: "c1",
		}
		r2v1 := compdesc.Reference{
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
		r3v2 := compdesc.Reference{
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

		r4v3 := compdesc.Reference{
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
				res := Must(cd.SelectReferences(refsel.Name("r1")))
				Expect(res).To(Equal(compdesc.References{r1v1, r1v2}))
			})

			It("selects by version", func() {
				res := Must(cd.SelectReferences(refsel.Version("v1")))
				Expect(res).To(Equal(compdesc.References{r1v1, r2v1}))
			})

			It("selects by version", func() {
				res := Must(cd.SelectReferences(refsel.ExtraIdentityByKeyPairs("version", "v1")))
				Expect(res).To(Equal(compdesc.References{r1v1})) // no r2v1: version not part of identity
			})
		})

		Context("attr selection", func() {
			It("selects by name", func() {
				res := Must(cd.SelectReferences(refsel.Name("r1")))
				Expect(res).To(Equal(compdesc.References{r1v1, r1v2}))
			})
			It("selects by version", func() {
				res := Must(cd.SelectReferences(refsel.Version("v1")))
				Expect(res).To(Equal(compdesc.References{r1v1, r2v1}))
			})
			It("selects by component", func() {
				res := Must(cd.SelectReferences(refsel.Component("c2")))
				Expect(res).To(Equal(compdesc.References{r2v1}))
			})

			It("selects by label name", func() {
				res := Must(cd.SelectReferences(refsel.LabelName("l1")))
				Expect(res).To(Equal(compdesc.References{r1v1, r2v1, r3v2}))
			})
			It("selects by label version", func() {
				res := Must(cd.SelectReferences(refsel.LabelVersion("v2")))
				Expect(res).To(Equal(compdesc.References{r3v2}))
			})
			It("selects by label value", func() {
				res := Must(cd.SelectReferences(refsel.LabelValue("labelvalue")))
				Expect(res).To(Equal(compdesc.References{r1v1, r3v2}))
			})
			It("selects unrelated by label name and value", func() {
				res := Must(cd.SelectReferences(refsel.LabelName("l1"), refsel.LabelValue("labelvalue")))
				Expect(res).To(Equal(compdesc.References{r1v1, r3v2})) // unrelated checks at resource level
			})
			It("selects related by label name and value", func() {
				res := Must(cd.SelectReferences(refsel.Label(labelsel.Name("l1"), labelsel.Value("labelvalue"))))
				Expect(res).To(Equal(compdesc.References{r1v1})) // related checks at label level
			})
			It("selects by signed label", func() {
				res := Must(cd.SelectReferences(labelsel.Signed()))
				Expect(res).To(Equal(compdesc.References{r3v2}))
			})

			It("selects with extra identity", func() {
				res := Must(cd.SelectReferences(refsel.ExtraIdentityByKeyPairs("extra", "value", "other", "othervalue")))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})

			It("selects none with wrong extra identity value", func() {
				res := Must(cd.SelectReferences(refsel.ExtraIdentityByKeyPairs("extra", "other")))
				Expect(res).To(BeEmpty())
			})

			It("selects none with wrong extra identity key", func() {
				res := Must(cd.SelectReferences(refsel.ExtraIdentityByKeyPairs("extra2", "value")))
				Expect(res).To(BeEmpty())
			})

			It("selects by or", func() {
				res := Must(cd.SelectReferences(refsel.Or(refsel.Name("r3"), refsel.Version("v2"))))
				Expect(res).To(Equal(compdesc.References{r1v2, r3v2}))
			})

			It("selects by and", func() {
				res := Must(cd.SelectReferences(refsel.And(refsel.Name("r1"), refsel.Version("v1"))))
				Expect(res).To(Equal(compdesc.References{r1v1}))
			})

			It("selects by negated selector", func() {
				res := Must(cd.SelectReferences(refsel.Not(refsel.Name("r1"))))
				Expect(res).To(Equal(compdesc.References{r2v1, r3v2, r4v3}))
			})

			It("selects by identity selector", func() {
				res := Must(cd.SelectReferences(refsel.IdentityByKeyPairs("r4", "extra", "value", "other", "othervalue")))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})
			It("selects none by identity selector with missing attribute", func() {
				res := Must(cd.SelectReferences(refsel.IdentityByKeyPairs("r4", "extra", "value")))
				Expect(res).To(BeEmpty())
			})

			It("selects by partial identity selector", func() {
				res := Must(cd.SelectReferences(refsel.PartialIdentityByKeyPairs("name", "r4", "extra", "value", "other", "othervalue")))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})
			It("selects by partial identity selector with partial attributes", func() {
				res := Must(cd.SelectReferences(refsel.PartialIdentityByKeyPairs("name", "r4", "extra", "value")))
				Expect(res).To(Equal(compdesc.References{r4v3}))
			})
			It("selects none by partial identity selector with missing attribute", func() {
				res := Must(cd.SelectReferences(refsel.IdentityByKeyPairs("r4", "extra", "value", "dummy", "dummy")))
				Expect(res).To(BeEmpty())
			})
		})
	})
})
