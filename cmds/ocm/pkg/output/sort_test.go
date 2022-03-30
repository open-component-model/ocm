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

package output

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		slice.Sort(compare_column(2))
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
		slice.Sort(compare_column(3))
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
		sortFixed(1, slice, compare_column(2))
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
		//slice.SortIndexed(compare_fixed_column(1, 2))
		sortFixed(1, slice, compare_column(2))
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
