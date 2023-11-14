// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package env

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

var _ = Describe("Environment", func() {

	It("loads environment", func() {
		h := NewEnvironment(TestData())
		defer h.Cleanup()
		data, err := vfs.ReadFile(h.FileSystem(), "/testdata/testfile.txt")
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("this is some test data"))
	})

	It("reuses context", func() {
		ctx := ocm.New()
		h := NewEnvironment(OCMContext(ctx), FileSystem(osfs.OsFs))
		Expect(h.OCMContext()).To(BeIdenticalTo(ctx))
	})

})
