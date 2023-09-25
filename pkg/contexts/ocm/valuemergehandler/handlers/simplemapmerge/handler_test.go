// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package simplemapmerge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplemapmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
)

type Value = me.Value
type Config = me.Config

const ALGORITHM = me.ALGORITHM
const MODE_NONE = me.MODE_NONE
const MODE_LOCAL = me.MODE_LOCAL
const MODE_INBOUND = me.MODE_INBOUND

var NewConfig = me.NewConfig
var New = me.New

var _ = Describe("list merge", func() {
	handler := New()

	var e1, e2 Value
	var a, b hpi.Value

	BeforeEach(func() {
		e1 = map[string]interface{}{
			"name": "name1",
			"data": "entry1",
		}
		e2 = map[string]interface{}{
			"name": "name1",
			"data": "entry1",
		}

		MustBeSuccessful(a.SetValue(e1))
		b = a.Copy()
	})

	It("merges no change", func() {
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry", func() {
		e1["local"] = "local"
		MustBeSuccessful(a.SetValue(e1))
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry on both sides", func() {
		e1["local"] = "local"
		e2["inbound"] = "inbound"
		MustBeSuccessful(a.SetValue(e1))
		MustBeSuccessful(b.SetValue(e2))

		MustBeSuccessful(handler.Merge(nil, a, &b, nil))

		var r Value
		MustBeSuccessful(b.GetValue(&r))

		e2["local"] = "local"
		Expect(r).To(DeepEqual(e2))
	})

	It("updates to inbound", func() {
		e2["name"] = "inbound"
		MustBeSuccessful(b.SetValue(e2))
		r := b.Copy()
		MustBeSuccessful(handler.Merge(nil, a, &b, NewConfig(MODE_INBOUND, nil)))

		Expect(b).To(DeepEqual(r))
	})

	It("keeps local", func() {
		e2["name"] = "inbound"
		MustBeSuccessful(b.SetValue(e1))
		MustBeSuccessful(handler.Merge(nil, a, &b, NewConfig(MODE_LOCAL, nil)))

		Expect(b).To(DeepEqual(a))
	})

	It("fails for none mode", func() {
		e2["data"] = "X"
		MustBeSuccessful(b.SetValue(e2))
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, NewConfig(MODE_NONE, nil))), "[simpleMapMerge]: target value for \"data\" changed")
	})

	It("fails for wrong type", func() {
		MustBeSuccessful(b.SetValue(true))
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, nil)), "[simpleMapMerge] inbound value is not valid: json: cannot unmarshal bool into Go value of type map[string]interface {}")
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, b, &a, nil)), "[simpleMapMerge] local value is not valid: json: cannot unmarshal bool into Go value of type map[string]interface {}")
	})

	Context("cascading", func() {
		var d1, d2 Value
		var cfg *Config

		BeforeEach(func() {
			d1 = Value{
				"k1": e1,
				"k2": e2,
			}

			MustBeSuccessful(a.SetValue(d1))
			b = a.Copy()
			MustBeSuccessful(b.GetValue(&d2))
			cfg = NewConfig("", Must(hpi.NewSpecification(ALGORITHM, NewConfig(MODE_INBOUND))))
		})

		It("handles equal", func() {
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			Expect(b).To(DeepEqual(a))
		})

		It("handles merge", func() {
			d1["k1"].(Value)["local"] = "local"
			d2["k1"].(Value)["inbound"] = "inbound"

			MustBeSuccessful(a.SetValue(d1))
			MustBeSuccessful(b.SetValue(d2))

			d2["k1"].(Value)["local"] = "local"

			MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, nil)), "[simpleMapMerge]: target value for \"k1\" changed")
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			var r Value
			MustBeSuccessful(b.GetValue(&r))
			Expect(r).To(DeepEqual(d2))
		})

		It("resolves to inbound", func() {
			d2["k1"].(Value)["data"] = "inbound"

			MustBeSuccessful(a.SetValue(d1))
			MustBeSuccessful(b.SetValue(d2))

			MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, nil)), "[simpleMapMerge]: target value for \"k1\" changed")
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			var r Value
			MustBeSuccessful(b.GetValue(&r))
			Expect(r).To(DeepEqual(d2))
		})

		It("resolves to local", func() {
			cfg = NewConfig("", Must(hpi.NewSpecification(ALGORITHM, NewConfig(MODE_LOCAL))))

			d1["k1"].(Value)["data"] = "local"

			MustBeSuccessful(a.SetValue(d1))
			MustBeSuccessful(b.SetValue(d2))

			MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, nil)), "[simpleMapMerge]: target value for \"k1\" changed")
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			var r Value
			MustBeSuccessful(b.GetValue(&r))
			Expect(r).To(DeepEqual(d1))
		})

	})
})
