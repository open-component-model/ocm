package simplelistmerge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	me "github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplelistmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

type Value = me.Value
type Config = me.Config

var NewConfig = me.NewConfig
var New = me.New

var _ = Describe("list merge", func() {
	handler := New()

	var e1, e2 Value
	var a, b hpi.Value

	BeforeEach(func() {
		e1 = []interface{}{
			"name1",
			"entry1",
		}
		e2 = []interface{}{
			"name1",
			"entry1",
		}

		MustBeSuccessful(a.SetValue(e1))
		b = a.Copy()
	})

	It("merges no change", func() {
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry", func() {
		e1 = append(e1, "local")
		MustBeSuccessful(a.SetValue(e1))
		MustBeSuccessful(handler.Merge(nil, a, &b, nil))
		Expect(b).To(Equal(a))
	})

	It("adds new entry on both sides", func() {
		e1 = append(e1, "local")
		e2 = append(e2, "inbound")
		MustBeSuccessful(a.SetValue(e1))
		MustBeSuccessful(b.SetValue(e2))

		MustBeSuccessful(handler.Merge(nil, a, &b, nil))

		var r Value
		MustBeSuccessful(b.GetValue(&r))

		e2 = append(e2, "local")
		Expect(r).To(DeepEqual(e2))
	})

	It("fails for wrong type", func() {
		MustBeSuccessful(b.SetValue(true))
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, a, &b, nil)), "[simpleListMerge] inbound value is not valid: json: cannot unmarshal bool into Go value of type []interface {}")
		MustFailWithMessage(ErrorFrom(handler.Merge(nil, b, &a, nil)), "[simpleListMerge] local value is not valid: json: cannot unmarshal bool into Go value of type []interface {}")
	})
})
