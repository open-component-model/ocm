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

package accessio

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const testData = "test data"

var _ = Describe("resettable reader", func() {
	It("pipes data", func() {
		w, err := NewFileBuffer()
		Expect(err).To(Succeed())
		defer w.Release()

		r, err := w.Reader()
		Expect(err).To(Succeed())
		Expect(r).NotTo(BeNil())
		n, err := w.Write([]byte(testData))
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(testData)))
		Expect(w.Close()).To(Succeed())

		data, err := io.ReadAll(r)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(testData))
		Expect(r.Close()).To(Succeed())

		Expect(w.Release()).To(Succeed())
		Expect(vfs.Exists(osfs.New(), w.path)).To(BeFalse())
	})

	It("rereads data", func() {
		w, err := NewFileBuffer()
		Expect(err).To(Succeed())
		defer w.Release()

		n, err := w.Write([]byte(testData))
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(testData)))
		Expect(w.Close()).To(Succeed())

		r, err := w.Reader()
		Expect(err).To(Succeed())
		data, err := io.ReadAll(r)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(testData))
		Expect(r.Close()).To(Succeed())

		r, err = w.Reader()
		Expect(err).To(Succeed())
		data, err = io.ReadAll(r)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(testData))
		Expect(r.Close()).To(Succeed())

		Expect(w.Release()).To(Succeed())
		Expect(vfs.Exists(osfs.New(), w.path)).To(BeFalse())
	})

	It("delays delete", func() {
		w, err := NewFileBuffer()
		Expect(err).To(Succeed())
		defer w.Release()

		n, err := w.Write([]byte(testData))
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(testData)))
		Expect(w.Close()).To(Succeed())

		r, err := w.Reader()
		Expect(err).To(Succeed())
		data, err := io.ReadAll(r)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(testData))
		Expect(r.Close()).To(Succeed())

		r, err = w.Reader()

		Expect(w.Release()).To(Succeed())
		Expect(vfs.Exists(osfs.New(), w.path)).To(BeTrue())

		Expect(err).To(Succeed())
		data, err = io.ReadAll(r)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(testData))
		Expect(r.Close()).To(Succeed())

		Expect(vfs.Exists(osfs.New(), w.path)).To(BeFalse())
	})

})
