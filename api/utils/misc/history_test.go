package misc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/misc"
)

type Elem struct {
	history misc.History
	key     misc.NameVersion
}

var _ misc.HistoryElement = (*Elem)(nil)

func (e *Elem) GetHistory() misc.History {
	return e.history
}

func (e *Elem) GetKey() misc.NameVersion {
	return e.key
}

func New(n string, hist ...string) *Elem {
	e := &Elem{
		key: misc.NewNameVersion(n, ""),
	}
	for _, h := range hist {
		e.history.Add("test", misc.NewNameVersion(h, ""))
	}
	return e
}

var _ = Describe("processing buffer", func() {
	Context("history", func() {
		It("compare", func() {
			Expect(misc.CompareHistoryElement(New("a"), New("b")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("b"), New("a")) > 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("a"), New("a")) == 0).To(BeTrue())

			Expect(misc.CompareHistoryElement(New("a", "a"), New("b", "a")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("b", "a"), New("a", "a")) > 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("a", "a"), New("a", "a")) == 0).To(BeTrue())

			Expect(misc.CompareHistoryElement(New("a", "a"), New("a", "b")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("b", "a"), New("a", "b")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("a", "a"), New("b", "b")) < 0).To(BeTrue())

			Expect(misc.CompareHistoryElement(New("a", "a"), New("a", "a", "b")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("b", "a"), New("a", "a", "b")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("a", "a"), New("b", "a", "b")) < 0).To(BeTrue())

			Expect(misc.CompareHistoryElement(New("a"), New("a", "a")) < 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("b"), New("a", "a")) > 0).To(BeTrue())

			Expect(misc.CompareHistoryElement(New("a", "a"), New("a")) > 0).To(BeTrue())
			Expect(misc.CompareHistoryElement(New("a", "a"), New("b")) < 0).To(BeTrue())
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
			misc.SortHistoryElements(s)
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
