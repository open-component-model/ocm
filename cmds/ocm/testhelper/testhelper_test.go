// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ = Describe("Test Environment", func() {

	It("loads test environment", func() {
		h := NewTestEnv(TestData())
		defer h.Cleanup()
		data, err := vfs.ReadFile(h.FileSystem(), "/testdata/testfile.txt")
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("this is some test data"))
	})
})
