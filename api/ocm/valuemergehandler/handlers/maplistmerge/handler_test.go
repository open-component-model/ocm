package maplistmerge_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm"
	me "ocm.software/ocm/api/ocm/valuemergehandler/handlers/maplistmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/simplemapmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
)

type (
	Value  = me.Value
	VEntry = simplemapmerge.Value
	Config = me.Config
)

const (
	ALGORITHM    = me.ALGORITHM
	MODE_NONE    = me.MODE_NONE
	MODE_LOCAL   = me.MODE_LOCAL
	MODE_INBOUND = me.MODE_INBOUND
)

var (
	NewConfig = me.NewConfig
	New       = me.New
)

var _ = Describe("list merge", func() {
	handler := New()

	var e1, e2, e3, e4 map[string]interface{}
	var va, vn Value
	var a, b hpi.Value

	BeforeEach(func() {
		e1 = VEntry{
			"name": "name1",
			"data": "entry1",
		}
		e2 = VEntry{
			"name": "name2",
			"data": "entry2",
		}
		e3 = VEntry{
			"name": "name3",
			"data": "entry3",
		}
		e4 = VEntry{
			"name": "name4",
			"data": "entry4",
		}

		va = Value{e1, e2}
		vn = Value{e1, e2}

		MustBeSuccessful(a.SetValue(va))
		b = a.Copy()
	})

	It("merges no change", func() {
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry", func() {
		MustBeSuccessful(a.SetValue(append(va, e3)))
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry on both sides", func() {
		MustBeSuccessful(a.SetValue(append(vn, e4)))
		MustBeSuccessful(b.SetValue(append(vn, e3)))

		MustBeSuccessful(handler.Merge(nil, a, &b, nil))

		var r Value
		MustBeSuccessful(b.GetValue(&r))

		Expect(r).To(DeepEqual(append(vn, e3, e4)))
	})

	It("updates to inbound", func() {
		vn[0]["data"] = "X"
		MustBeSuccessful(b.SetValue(vn))
		r := b.Copy()
		MustBeSuccessful(handler.Merge(nil, a, &b, NewConfig("name", MODE_INBOUND)))

		Expect(b).To(DeepEqual(r))
	})

	It("keeps local", func() {
		vn[0]["data"] = "X"
		MustBeSuccessful(b.SetValue(vn))
		MustBeSuccessful(handler.Merge(nil, a, &b, NewConfig("name", MODE_LOCAL)))

		Expect(b).To(DeepEqual(a))
	})

	It("fails for none mode", func() {
		vn[0]["data"] = "X"
		MustBeSuccessful(b.SetValue(vn))
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, NewConfig("name", MODE_NONE))), "[mapListMerge]: target value for \"name1\" changed")
	})

	It("fails for wrong type", func() {
		MustBeSuccessful(b.SetValue(true))
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, nil)), "[mapListMerge] inbound value is not valid: json: cannot unmarshal bool into Go value of type []map[string]interface {}")
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, b, &a, nil)), "[mapListMerge] local value is not valid: json: cannot unmarshal bool into Go value of type []map[string]interface {}")
	})

	Context("cascading", func() {
		var d1, d2 Value
		var cfg *Config
		var keycfg *Config
		var m1, m2 simplemapmerge.Value

		BeforeEach(func() {
			m1 = simplemapmerge.Value{
				"key":   "name1",
				"value": "value1",
			}
			m2 = simplemapmerge.Value{
				"key":   "name2",
				"value": "value3",
			}
			d1 = Value{
				m1, m2,
			}

			MustBeSuccessful(a.SetValue(d1))
			b = a.Copy()
			MustBeSuccessful(b.GetValue(&d2))
			cfg = NewConfig("key", "", Must(hpi.NewSpecification(simplemapmerge.ALGORITHM, simplemapmerge.NewConfig(simplemapmerge.MODE_INBOUND))))
			keycfg = NewConfig("key", "")
		})

		It("handles equal", func() {
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			Expect(b).To(DeepEqual(a))
		})

		It("handles merge", func() {
			d1[0]["local"] = "local"
			d2[0]["inbound"] = "inbound"

			MustBeSuccessful(a.SetValue(d1))
			MustBeSuccessful(b.SetValue(d2))

			d2[0]["local"] = "local"

			MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, keycfg)), "[mapListMerge]: target value for \"name1\" changed")
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			var r Value
			MustBeSuccessful(b.GetValue(&r))
			Expect(r).To(DeepEqual(d2))
		})

		It("resolves to inbound", func() {
			d2[0]["data"] = "inbound"

			MustBeSuccessful(a.SetValue(d1))
			MustBeSuccessful(b.SetValue(d2))

			MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, keycfg)), "[mapListMerge]: target value for \"name1\" changed")
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			var r Value
			MustBeSuccessful(b.GetValue(&r))
			Expect(r).To(DeepEqual(d2))
		})

		It("resolves to local", func() {
			cfg = NewConfig("key", "", Must(hpi.NewSpecification(simplemapmerge.ALGORITHM, simplemapmerge.NewConfig(MODE_LOCAL))))

			d1[0]["data"] = "local"

			MustBeSuccessful(a.SetValue(d1))
			MustBeSuccessful(b.SetValue(d2))

			MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, keycfg)), "[mapListMerge]: target value for \"name1\" changed")
			MustBeSuccessful(handler.Merge(ocm.DefaultContext(), a, &b, cfg))

			var r Value
			MustBeSuccessful(b.GetValue(&r))
			Expect(r).To(DeepEqual(d1))
		})
	})
})
