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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("simple data processing", func() {

	Context("add", func() {

		It("end", func() {
			root := DLL{}

			d1 := NewDLL(1)

			root.Append(d1)
			Expect(root.Next()).To(BeIdenticalTo(d1))
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeNil())
			Expect(d1.Prev()).To(BeIdenticalTo(&root))

			Expect(Slice(&DLLRoot{root: root})).To(Equal([]interface{}{d1}))
		})
		It("middle", func() {
			root := DLL{}

			d1 := NewDLL(1)
			d2 := NewDLL(2)

			root.Append(d2)
			root.Append(d1)

			Expect(root.Next()).To(BeIdenticalTo(d1))
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeIdenticalTo(d2))
			Expect(d1.Prev()).To(BeIdenticalTo(&root))

			Expect(d2.Next()).To(BeNil())
			Expect(d2.Prev()).To(BeIdenticalTo(d1))

			Expect(Slice(&DLLRoot{root: root})).To(Equal([]interface{}{d1, d2}))
		})
	})
	Context("remove", func() {

		It("end", func() {
			root := DLL{}

			d1 := NewDLL(1)

			root.Append(d1)

			d1.Remove()

			Expect(root.Next()).To(BeNil())
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeNil())
			Expect(d1.Prev()).To(BeNil())
		})
		It("middle", func() {
			root := DLL{}

			d1 := NewDLL(1)
			d2 := NewDLL(2)

			root.Append(d2)
			root.Append(d1)

			d1.Remove()

			Expect(root.Next()).To(BeIdenticalTo(d2))
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeNil())
			Expect(d1.Prev()).To(BeNil())

			Expect(d2.Next()).To(BeNil())
			Expect(d2.Prev()).To(BeIdenticalTo(&root))

		})
	})

})
