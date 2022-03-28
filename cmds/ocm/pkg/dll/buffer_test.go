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

package dll

import (
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Result struct {
	result []interface{}
	wg     sync.WaitGroup
}

func Gather(i Iterable) *Result {
	r := &Result{}
	r.wg.Add(1)
	go func() {
		r.result = AsSlice(i)
		r.wg.Done()
	}()
	return r
}

func (r *Result) Get() []interface{} {
	r.wg.Wait()
	return r.result
}

var _ = Describe("processing buffer", func() {

	Context("simple", func() {
		It("add", func() {
			buf := NewSimpleBuffer()

			promise := Gather(buf)

			buf.Add(NewEntry(1, 1))
			buf.Add(NewEntry(2, 2))
			buf.Add(NewEntry(3, 3))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{1, 2, 3}))
			Expect(AsSlice(ValueIterable(buf))).To(Equal([]interface{}{1, 2, 3}))
		})

		It("add filtered", func() {
			buf := NewSimpleBuffer()

			promise := Gather(buf)

			buf.Add(NewEntry(1, 1))
			buf.Add(NewEntry(2, 2, false))
			buf.Add(NewEntry(3, 3))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{1, 3}))
			Expect(AsSlice(ValueIterable(buf))).To(Equal([]interface{}{1, 2, 3}))
		})
	})

	Context("add ordered", func() {
		It("add", func() {
			buf := NewOrderedBuffer()

			promise := Gather(buf)

			buf.Add(NewEntry(1, 1))
			buf.Add(NewEntry(2, 2))
			buf.Add(NewEntry(3, 3))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{1, 2, 3}))
			Expect(AsSlice(ValueIterable(buf))).To(Equal([]interface{}{1, 2, 3}))

		})
		It("add filtered", func() {
			buf := NewOrderedBuffer()

			promise := Gather(buf)

			buf.Add(NewEntry(1, 1))
			buf.Add(NewEntry(2, 2, false))
			buf.Add(NewEntry(3, 3))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{1, 3}))
			Expect(AsSlice(ValueIterable(buf))).To(Equal([]interface{}{1, 2, 3}))

		})
		It("add mixed order", func() {
			buf := NewOrderedBuffer()

			promise := Gather(buf)

			buf.Add(NewEntry(3, 3))
			buf.Add(NewEntry(1, 1))
			buf.Add(NewEntry(2, 2))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{1, 2, 3}))
			Expect(AsSlice(ValueIterable(buf))).To(Equal([]interface{}{3, 1, 2}))
		})
		It("add mixed order filtered", func() {
			buf := NewOrderedBuffer()

			promise := Gather(buf)

			buf.Add(NewEntry(3, 3))
			buf.Add(NewEntry(1, 1, false))
			buf.Add(NewEntry(2, 2))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{2, 3}))
			Expect(AsSlice(ValueIterable(buf))).To(Equal([]interface{}{3, 1, 2}))
		})
	})
})
