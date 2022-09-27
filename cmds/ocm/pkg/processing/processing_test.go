// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package processing

import (
	"strings"

	"github.com/mandelsoft/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/pkg/utils/logger"
)

var AddOne = func(log logging.Logger) func(e interface{}) interface{} {
	return func(e interface{}) interface{} {
		log.Info("add one to number", "num", e.(int))
		return e.(int) + 1
	}
}

var Mul = func(log logging.Logger) func(n, fac int) ExplodeFunction {
	return func(n, fac int) ExplodeFunction {
		return func(e interface{}) []interface{} {
			r := []interface{}{}
			v := e.(int)
			log.Info("explode", "num", e.(int))
			for i := 1; i <= n; i++ {
				r = append(r, v)
				v = v * fac
			}
			return r
		}
	}
}

var _ = Describe("simple data processing", func() {
	var log logging.Logger

	BeforeEach(func() {
		log = logger.NewDefaultLoggerContext().Logger()
	})

	Context("sequential", func() {
		It("map", func() {
			By("*** sequential map")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Map(AddOne(log)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{2, 3, 4}))
		})

		It("explode", func() {
			By("*** sequential explode")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Map(AddOne(log)).Explode(Mul(log)(3, 2)).Map(Identity).Process(data).AsSlice()
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
			result := Chain().Map(Identity).Parallel(3).Map(AddOne(log)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 3, 4,
			}))
		})
		It("explode", func() {
			By("*** parallel explode")

			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Parallel(3).Explode(Mul(log)(3, 2)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				1, 2, 4,
				2, 4, 8,
				3, 6, 12,
			}))
		})
		It("explode-map", func() {
			By("*** parallel explode")

			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Parallel(3).Explode(Mul(log)(3, 2)).Map(AddOne(log)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 3, 5,
				3, 5, 9,
				4, 7, 13,
			}))
		})
	})
	Context("compose", func() {
		It("appends a chain", func() {
			chain := Chain().Map(AddOne(log))
			slice := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			sub := Chain().Explode(Mul(log)(2, 2))
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
			result := Chain().Explode(Split).Parallel(3).Filter(Filter).Sort(Compare).Process(data.IndexedSliceAccess(input)).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				"is", "multi-line", "some", "text", "this", "with", "words",
			}))
		})
	})
})
