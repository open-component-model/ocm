// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package simplelistmerge

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var _ = Describe("list merge", func() {
	var e1, e2, e3, e4 map[string]interface{}
	var va, vn LabelValue
	var a, b *metav1.Label

	BeforeEach(func() {
		e1 = map[string]interface{}{
			"name": "name1",
			"data": "entry1",
		}
		e2 = map[string]interface{}{
			"name": "name2",
			"data": "entry2",
		}
		e3 = map[string]interface{}{
			"name": "name3",
			"data": "entry3",
		}
		e4 = map[string]interface{}{
			"name": "name4",
			"data": "entry4",
		}

		va = LabelValue{e1, e2}
		vn = LabelValue{e1, e2}

		a = Must(metav1.NewLabel("label", va))
		b = a.DeepCopy()
	})

	It("merges no change", func() {
		MustBeSuccessful(Handler{}.Merge(nil, a, b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry", func() {
		a = Must(metav1.NewLabel("label", append(va, e3)))
		MustBeSuccessful(Handler{}.Merge(nil, a, b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry on both sides", func() {
		a = Must(metav1.NewLabel("label", append(vn, e4)))
		b = Must(metav1.NewLabel("label", append(vn, e3)))

		MustBeSuccessful(Handler{}.Merge(nil, a, b, nil))

		var r LabelValue
		MustBeSuccessful(b.GetValue(&r))

		Expect(r).To(DeepEqual(append(vn, e3, e4)))
	})

	It("updates to inbound", func() {
		vn[0]["data"] = "X"
		b = Must(metav1.NewLabel("label", vn))
		r := b.DeepCopy()
		MustBeSuccessful(Handler{}.Merge(nil, a, b, NewConfig("name", MODE_INBOUND)))

		Expect(b).To(DeepEqual(r))
	})

	It("keeps local", func() {
		vn[0]["data"] = "X"
		b = Must(metav1.NewLabel("label", vn))
		MustBeSuccessful(Handler{}.Merge(nil, a, b, NewConfig("name", MODE_LOCAL)))

		Expect(b).To(DeepEqual(a))
	})

	It("fails for none mode", func() {
		vn[0]["data"] = "X"
		b = Must(metav1.NewLabel("label", vn))
		MustFailWithMessage(Handler{}.Merge(nil, a, b, NewConfig("name", MODE_NONE)), "target value for \"name1\" changed")
	})

	It("fails for wrong type", func() {
		b = Must(metav1.NewLabel("label", true))
		MustFailWithMessage(Handler{}.Merge(nil, a, b, nil), "inbound label value is no list of objects: json: cannot unmarshal bool into Go value of type simplelistmerge.LabelValue")
		MustFailWithMessage(Handler{}.Merge(nil, b, a, nil), "local label value is no list of objects: json: cannot unmarshal bool into Go value of type simplelistmerge.LabelValue")
	})
})
