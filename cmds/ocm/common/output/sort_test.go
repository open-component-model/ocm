package output

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/cmds/ocm/common/data"
)

var _ = Describe("sort", func() {
	h1i2a1b3 := []string{"h1", "i2", "a1", "b3"}
	h2i2a1b3 := []string{"h2", "i2", "a1", "b3"}
	h1i2a3b2 := []string{"h1", "i2", "a3", "b2"}
	h2i2a3b2 := []string{"h2", "i2", "a3", "b2"}
	h1i2a2b1 := []string{"h1", "i2", "a2", "b1"}
	h2i2a2b1 := []string{"h2", "i2", "a2", "b1"}

	values := []interface{}{
		h1i2a1b3,
		h1i2a3b2,
		h1i2a2b1,
		h2i2a1b3,
		h2i2a3b2,
		h2i2a2b1,
	}

	It("sort a", func() {
		slice := data.IndexedSliceAccess(values).Copy()
		slice.Sort(compareColumn(2))
		Expect(slice).To(Equal(data.IndexedSliceAccess{
			h1i2a1b3,
			h2i2a1b3,
			h1i2a2b1,
			h2i2a2b1,
			h1i2a3b2,
			h2i2a3b2,
		}))
	})
	It("sort b", func() {
		slice := data.IndexedSliceAccess(values).Copy()
		slice.Sort(compareColumn(3))
		Expect(slice).To(Equal(data.IndexedSliceAccess{
			h1i2a2b1,
			h2i2a2b1,
			h1i2a3b2,
			h2i2a3b2,
			h1i2a1b3,
			h2i2a1b3,
		}))
	})
	It("sort fixed h a", func() {
		slice := data.IndexedSliceAccess(values).Copy()
		sortFixed(1, slice, compareColumn(2))
		Expect(slice).To(Equal(data.IndexedSliceAccess{
			h1i2a1b3,
			h1i2a2b1,
			h1i2a3b2,
			h2i2a1b3,
			h2i2a2b1,
			h2i2a3b2,
		}))

		values := []interface{}{
			h1i2a3b2,
			h2i2a1b3,
			h1i2a1b3,
			h2i2a3b2,
			h1i2a2b1,
			h2i2a2b1,
		}
		slice = data.IndexedSliceAccess(values)
		// slice.SortIndexed(compare_fixed_column(1, 2))
		sortFixed(1, slice, compareColumn(2))
		Expect(slice).To(Equal(data.IndexedSliceAccess{
			h1i2a1b3,
			h2i2a1b3,
			h1i2a2b1,
			h2i2a2b1,
			h1i2a3b2,
			h2i2a3b2,
		}))
	})
})
