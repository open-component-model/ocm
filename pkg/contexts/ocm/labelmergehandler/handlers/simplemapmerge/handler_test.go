// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package simplemapmerge

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var _ = Describe("list merge", func() {
	var e1, e2 LabelValue
	var a, b *metav1.Label

	BeforeEach(func() {
		e1 = map[string]interface{}{
			"name": "name1",
			"data": "entry1",
		}
		e2 = map[string]interface{}{
			"name": "name1",
			"data": "entry1",
		}

		a = Must(metav1.NewLabel("label", e1))
		b = a.DeepCopy()
	})

	It("merges no change", func() {
		MustBeSuccessful(Handler{}.Merge(nil, a, b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry", func() {
		e1["local"] = "local"
		a = Must(metav1.NewLabel("label", e1))
		MustBeSuccessful(Handler{}.Merge(nil, a, b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry on both sides", func() {
		e1["local"] = "local"
		e2["inbound"] = "inbound"
		a = Must(metav1.NewLabel("label", e1))
		b = Must(metav1.NewLabel("label", e2))

		MustBeSuccessful(Handler{}.Merge(nil, a, b, nil))

		var r LabelValue
		MustBeSuccessful(b.GetValue(&r))

		e2["local"] = "local"
		Expect(r).To(DeepEqual(e2))
	})

	It("updates to inbound", func() {
		e2["name"] = "inbound"
		b = Must(metav1.NewLabel("label", e2))
		r := b.DeepCopy()
		MustBeSuccessful(Handler{}.Merge(nil, a, b, NewConfig(MODE_INBOUND)))

		Expect(b).To(DeepEqual(r))
	})

	It("keeps local", func() {
		e2["name"] = "inbound"
		b = Must(metav1.NewLabel("label", e1))
		MustBeSuccessful(Handler{}.Merge(nil, a, b, NewConfig(MODE_LOCAL)))

		Expect(b).To(DeepEqual(a))
	})

	It("fails for none mode", func() {
		e2["data"] = "X"
		b = Must(metav1.NewLabel("label", e2))
		MustFailWithMessage(Handler{}.Merge(nil, a, b, NewConfig(MODE_NONE)), "target value for \"data\" changed")
	})

	It("fails for wrong type", func() {
		b = Must(metav1.NewLabel("label", true))
		MustFailWithMessage(Handler{}.Merge(nil, a, b, nil), "inbound label value is no map object: json: cannot unmarshal bool into Go value of type simplemapmerge.LabelValue")
		MustFailWithMessage(Handler{}.Merge(nil, b, a, nil), "local label value is no map object: json: cannot unmarshal bool into Go value of type simplemapmerge.LabelValue")
	})
})
