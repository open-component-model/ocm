package misc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	common "ocm.software/ocm/api/utils/misc"
)

type Elem struct {
	history common.History
	key     common.NameVersion
}

var _ common.HistoryElement = (*Elem)(nil)

func (e *Elem) GetHistory() common.History {
	return e.history
}

func (e *Elem) GetKey() common.NameVersion {
	return e.key
}

func New(n string, hist ...string) *Elem {
	e := &Elem{
		key: common.NewNameVersion(n, ""),
	}
	for _, h := range hist {
		e.history.Add("test", common.NewNameVersion(h, ""))
	}
	return e
}

var _ = Describe("processing buffer", func() {
	Context("history", func() {
		It("compare", func() {
			Expect(common.CompareHistoryElement(New("a"), New("b")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("b"), New("a")) > 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("a"), New("a")) == 0).To(BeTrue())

			Expect(common.CompareHistoryElement(New("a", "a"), New("b", "a")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("b", "a"), New("a", "a")) > 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("a", "a"), New("a", "a")) == 0).To(BeTrue())

			Expect(common.CompareHistoryElement(New("a", "a"), New("a", "b")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("b", "a"), New("a", "b")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("a", "a"), New("b", "b")) < 0).To(BeTrue())

			Expect(common.CompareHistoryElement(New("a", "a"), New("a", "a", "b")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("b", "a"), New("a", "a", "b")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("a", "a"), New("b", "a", "b")) < 0).To(BeTrue())

			Expect(common.CompareHistoryElement(New("a"), New("a", "a")) < 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("b"), New("a", "a")) > 0).To(BeTrue())

			Expect(common.CompareHistoryElement(New("a", "a"), New("a")) > 0).To(BeTrue())
			Expect(common.CompareHistoryElement(New("a", "a"), New("b")) < 0).To(BeTrue())
		})
		It("sort", func() {
			s := []*Elem{
				New("a"),
				New("c"),
				New("b"),
				New("a", "a", "b"),
				New("a", "a"),
				New("a", "a", "c"),
				New("a", "a", "c", "d"),
				New("b", "a", "b"),
				New("b", "a"),
				New("b", "a", "c"),
				New("b", "a", "c", "d"),
			}
			common.SortHistoryElements(s)
			r := []*Elem{
				New("a"),
				New("a", "a"),
				New("b", "a"),
				New("a", "a", "b"),
				New("b", "a", "b"),
				New("a", "a", "c"),
				New("b", "a", "c"),
				New("a", "a", "c", "d"),
				New("b", "a", "c", "d"),
				New("b"),
				New("c"),
			}

			Expect(s).To(Equal(r))
		})
	})
})
