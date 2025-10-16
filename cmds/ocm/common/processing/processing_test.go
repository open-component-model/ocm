package processing

import (
	"bytes"
	"strings"

	"github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/cmds/ocm/common/data"
)

var AddOne = func(logger logging.Logger) func(e interface{}) interface{} {
	return func(e interface{}) interface{} {
		logger.Info("add one to number", "num", e.(int))
		return e.(int) + 1
	}
}

var Mul = func(logger logging.Logger) func(n, fac int) ExplodeFunction {
	return func(n, fac int) ExplodeFunction {
		return func(e interface{}) []interface{} {
			r := []interface{}{}
			v := e.(int)
			logger.Info("explode", "num", e.(int))
			for i := 1; i <= n; i++ {
				r = append(r, v)
				v = v * fac
			}
			return r
		}
	}
}

var _ = Describe("simple data processing", func() {
	var (
		log    logging.Context
		logger logging.Logger
		buf    *bytes.Buffer
	)

	BeforeEach(func() {
		log, buf = ocmlog.NewBufferedContext()
		logger = log.Logger()
	})

	Context("sequential", func() {
		It("map", func() {
			By("*** sequential map")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain(log).Map(AddOne(logger)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{2, 3, 4}))

			Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[3] add one to number num 1
V[3] add one to number num 2
V[3] add one to number num 3
`))
		})

		It("explode", func() {
			By("*** sequential explode")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain(log).Map(AddOne(logger)).Explode(Mul(logger)(3, 2)).Map(Identity).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 4, 8,
				3, 6, 12,
				4, 8, 16,
			}))
		})
	})
	Context("parallel", func() {
		It("map", func() {
			By("*** parallel map")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain(log).Map(Identity).Parallel(3).Map(AddOne(logger)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 3, 4,
			}))
		})
		It("explode", func() {
			By("*** parallel explode")

			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain(log).Parallel(3).Explode(Mul(logger)(3, 2)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				1, 2, 4,
				2, 4, 8,
				3, 6, 12,
			}))
		})
		It("explode-map", func() {
			By("*** parallel explode")

			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain(log).Parallel(3).Explode(Mul(logger)(3, 2)).Map(AddOne(logger)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 3, 5,
				3, 5, 9,
				4, 7, 13,
			}))
		})
	})
	Context("compose", func() {
		It("appends a chain", func() {
			chain := Chain(log).Map(AddOne(logger))
			slice := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			sub := Chain(log).Explode(Mul(logger)(2, 2))
			r := chain.Append(sub).Process(slice).AsSlice()
			Expect(r).To(Equal(data.IndexedSliceAccess([]interface{}{
				2, 4, 3, 6, 4, 8,
			})))
		})
	})

	Split := func(text interface{}) []interface{} {
		var words []interface{}
		t := text.(string)
		for t != "" {
			i := strings.IndexAny(t, " \t\n\r.,:!?")
			w := t
			t = ""
			if i >= 0 {
				t = w[i+1:]
				w = w[:i]
			}
			if w != "" {
				words = append(words, w)
			}
		}
		return words
	}

	ignore := []string{"a", "an", "the"}

	Filter := func(e interface{}) bool {
		s := e.(string)
		for _, w := range ignore {
			if s == w {
				return false
			}
		}
		return true
	}

	Compare := func(a, b interface{}) int {
		return strings.Compare(a.(string), b.(string))
	}

	Context("example", func() {
		It("example 1", func() {
			input := []interface{}{
				"this is a multi-line",
				"text with some words.",
			}

			_ = Compare
			result := Chain(log).Explode(Split).Parallel(3).Filter(Filter).Sort(Compare).Process(data.IndexedSliceAccess(input)).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				"is", "multi-line", "some", "text", "this", "with", "words",
			}))
		})
	})
})
