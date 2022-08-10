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

package common

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Printer", func() {

	var buf *bytes.Buffer
	var printer Printer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		printer = NewPrinter(buf)
	})

	It("print", func() {
		printer.Printf("test")
		Expect(buf.String()).To(Equal("test"))
	})
	It("prints line", func() {
		printer.Printf("test\n")
		Expect(buf.String()).To(Equal("test\n"))
	})
	It("prints lines", func() {
		printer.Printf("line\ntest\n")
		Expect(buf.String()).To(Equal("line\ntest\n"))
	})

	It("prints gap", func() {
		printer.Printf("line\n")
		p := printer.AddGap("  ")
		p.Printf("test\n")
		Expect(buf.String()).To(Equal("line\n  test\n"))
		p.Printf("next\n")
		Expect(buf.String()).To(Equal("line\n  test\n  next\n"))
		printer.Printf("back\n")
		Expect(buf.String()).To(Equal("line\n  test\n  next\nback\n"))
	})
})
