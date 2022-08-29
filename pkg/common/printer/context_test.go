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

package printer_test

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/printer"
)

var _ = Describe("Printer", func() {

	var buf *bytes.Buffer
	var pr printer.Printer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		pr = printer.NewPrinter(buf)
	})


	It("no printer", func() {
		ctx := context.Background()
		ctx2 := printer.WithGap(ctx, "  ")
		printer.Printf(ctx, "test")
		printer.Printf(ctx2, "test")
	})

	It("no context", func() {
		ctx := printer.WithPrinter(nil, pr)
		Expect(ctx).NotTo(BeNil())
		printer.Printf(ctx, "test")
		Expect(buf.String()).To(Equal("test"))
	})

	It("print", func() {
		ctx := printer.WithPrinter(nil, pr)
		ctx2 := printer.WithGap(ctx, "  ")
		printer.Printf(ctx2, "gapped\n")
		printer.Printf(ctx, "test")
		Expect(buf.String()).To(Equal("  gapped\ntest"))
	})
})
