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

package data

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func AddOne(e interface{}) interface{} {
	return e.(int) + 1
}

var _ = Describe("simple data processing", func() {

	Context("sequential", func() {

		It("map", func() {

			data := IndexedSliceAccess([]interface{}{1, 2, 3})

			result := Chain().Map(AddOne).Process(data).AsSlice()

			Expect([]interface{}(result)).To(Equal([]interface{}{2, 3, 4}))
		})

		It("map", func() {

			data := IndexedSliceAccess([]interface{}{1, 2, 3})

			result := Chain().Parallel(1).Map(Identity).Process(data).AsSlice()

			Expect([]interface{}(result)).To(Equal([]interface{}{1, 2, 3}))
		})
	})
})
