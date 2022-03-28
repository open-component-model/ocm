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
	"fmt"

	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func AddOne(e interface{}) interface{} {
	fmt.Printf("add 1 to %d\n", e.(int))
	return e.(int) + 1
}

func Mul(n, fac int) ExplodeFunction {
	return func(e interface{}) []interface{} {
		r := []interface{}{}
		v := e.(int)
		fmt.Printf("explode  %d\n", e.(int))
		for i := 1; i <= n; i++ {
			r = append(r, v)
			v = v * fac
		}
		return r
	}
}

var _ = Describe("simple data processing", func() {

	Context("sequential", func() {
		It("map", func() {
			fmt.Printf("*** sequential map\n")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Map(AddOne).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{2, 3, 4}))
		})

		It("explode", func() {
			fmt.Printf("*** sequential explode\n")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Map(AddOne).Explode(Mul(3, 2)).Map(Identity).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 4, 8,
				3, 6, 12,
				4, 8, 16,
			}))
		})
	})
	Context("parallel", func() {
		It("map", func() {
			fmt.Printf("*** parallel map\n")
			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Map(Identity).Parallel(3).Map(AddOne).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 3, 4,
			}))
		})
		It("explode", func() {
			fmt.Printf("*** parallel explode\n")

			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Parallel(3).Explode(Mul(3, 2)).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				1, 2, 4,
				2, 4, 8,
				3, 6, 12,
			}))
		})
		It("explode-map", func() {
			fmt.Printf("*** parallel explode\n")

			data := data.IndexedSliceAccess([]interface{}{1, 2, 3})
			result := Chain().Parallel(3).Explode(Mul(3, 2)).Map(AddOne).Process(data).AsSlice()
			Expect([]interface{}(result)).To(Equal([]interface{}{
				2, 3, 5,
				3, 5, 9,
				4, 7, 13,
			}))
		})
	})
})

/*



	})
})

*/
