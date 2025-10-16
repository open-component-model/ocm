package v1_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/ocm/compdesc/equivalent/testhelper"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

var _ = Describe("labels", func() {
	Context("access", func() {
		It("modifies label set", func() {
			labels := v1.Labels{}

			// Extend

			MustBeSuccessful(labels.Set("l1", "v1"))
			MustBeSuccessful(labels.Set("l2", "v2", v1.WithSigning()))
			MustBeSuccessful(labels.Set("l3", "v3", v1.WithSigning(), v1.WithVersion("v10")))
			MustBeSuccessful(labels.Set("l4", "v4", v1.WithVersion("v11")))

			Expect(labels).To(ConsistOf(
				v1.Label{
					Name:  "l1",
					Value: []byte(`"v1"`),
				},
				v1.Label{
					Name:    "l2",
					Value:   []byte(`"v2"`),
					Signing: true,
				},
				v1.Label{
					Name:    "l3",
					Value:   []byte(`"v3"`),
					Signing: true,
					Version: "v10",
				},
				v1.Label{
					Name:    "l4",
					Value:   []byte(`"v4"`),
					Version: "v11",
				},
			))

			Expect(labels.GetDef("l3")).To(Equal(&v1.Label{
				Name:    "l3",
				Value:   []byte(`"v3"`),
				Signing: true,
				Version: "v10",
			}))

			Expect(labels.GetIndex("l3")).To(Equal(2))
			Expect(labels.GetIndex("lx")).To(Equal(-1))
			data, ok := labels.Get("l3")
			Expect(ok).To(BeTrue())
			Expect(data).To(Equal([]byte(`"v3"`)))
			var str string
			Expect(labels.GetValue("l3", &str)).To(BeTrue())
			Expect(str).To(Equal("v3"))

			// Modify

			labels.Set("l4", "modl4", v1.WithVersion("v100"))
			Expect(labels).To(ConsistOf(
				v1.Label{
					Name:  "l1",
					Value: []byte(`"v1"`),
				},
				v1.Label{
					Name:    "l2",
					Value:   []byte(`"v2"`),
					Signing: true,
				},
				v1.Label{
					Name:    "l3",
					Value:   []byte(`"v3"`),
					Signing: true,
					Version: "v10",
				},
				v1.Label{
					Name:    "l4",
					Value:   []byte(`"modl4"`),
					Version: "v100",
				},
			))
			Expect(labels.GetDef("l4")).To(Equal(&v1.Label{
				Name:    "l4",
				Value:   []byte(`"modl4"`),
				Version: "v100",
			}))
		})

		It("handles complex values", func() {
			var labels v1.Labels

			var nested v1.Labels
			nested.Set("label", "value", v1.WithSigning())
			meta := &v1.ObjectMeta{
				Name:    "value",
				Version: "v1.0.0",
				Labels:  nested,
				Provider: v1.Provider{
					Name:   "acme.org",
					Labels: nested,
				},
				CreationTime: nil,
			}

			labels.Set("l1", meta)
			Expect(labels).To(ConsistOf(
				v1.Label{
					Name:  "l1",
					Value: []byte(`{"name":"value","version":"v1.0.0","labels":[{"name":"label","value":"value","signing":true}],"provider":{"name":"acme.org","labels":[{"name":"label","value":"value","signing":true}]}}`),
				},
			))
			var get v1.ObjectMeta
			Expect(labels.GetValue("l1", &get)).To(BeTrue())
			Expect(&get).To(Equal(meta))
		})
	})

	Context("equivalence", func() {
		var labels v1.Labels

		BeforeEach(func() {
			labels.Clear()
			labels.Set("label1", "value1", v1.WithSigning())
			labels.Set("label2", "value2", v1.WithSigning(), v1.WithVersion("v1"))
			labels.Set("label3", "value3")
			labels.Set("label4", "value4", v1.WithVersion("v1"))
		})

		It("detects equal", func() {
			eq := labels.Equivalent(labels.Copy())
			CheckEquivalent(eq)
		})

		It("detects volatile value modification", func() {
			mod := labels.Copy()
			mod.Set("label3", "mod")
			CheckNotEquivalent(labels.Equivalent(mod))
			CheckNotEquivalent(mod.Equivalent(labels))
		})

		It("detects volatile version modification", func() {
			mod := labels.Copy()
			mod.Set("label4", "value4", v1.WithVersion("v2"))
			CheckNotEquivalent(labels.Equivalent(mod))
			CheckNotEquivalent(mod.Equivalent(labels))
		})

		It("detects new volatile label", func() {
			mod := labels.Copy()
			mod.Set("label5", "mod")
			CheckNotEquivalent(labels.Equivalent(mod))
			CheckNotEquivalent(mod.Equivalent(labels))
		})

		It("detects non-volatile value modification", func() {
			mod := labels.Copy()
			mod.Set("label2", "mod", v1.WithSigning(), v1.WithVersion("v1"))
			CheckNotLocalHashEqual(labels.Equivalent(mod))
			CheckNotLocalHashEqual(mod.Equivalent(labels))
		})

		It("detects non-volatile version modification", func() {
			mod := labels.Copy()
			mod.Set("label2", "value2", v1.WithSigning(), v1.WithVersion("v2"))
			CheckNotLocalHashEqual(labels.Equivalent(mod))
			CheckNotLocalHashEqual(mod.Equivalent(labels))
		})

		It("detects new non-volatile label", func() {
			mod := labels.Copy()
			mod.Set("label5", "mod", v1.WithSigning())
			CheckNotLocalHashEqual(labels.Equivalent(mod))
			CheckNotLocalHashEqual(mod.Equivalent(labels))
		})

		It("detects change og signing property", func() {
			mod := labels.Copy()
			mod.Set("label1", "value1")
			CheckNotLocalHashEqual(labels.Equivalent(mod))
			CheckNotLocalHashEqual(mod.Equivalent(labels))
		})
	})
})
