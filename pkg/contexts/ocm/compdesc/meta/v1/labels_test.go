// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package v1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var _ = Describe("labels", func() {
	Context("access", func() {
		It("modififies label set", func() {
			labels := v1.Labels{}

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
})
