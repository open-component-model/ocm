// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package create_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
)

const ARCH = "/tmp/ctf"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates common transport archive", func() {

		Expect(env.Execute("create", "ctf", "-ft", "directory", ARCH)).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		Expect(env.ReadTextFile(env.Join(ARCH, ctf.ArtifactIndexFileName))).To(Equal("{\"schemaVersion\":1,\"artifacts\":null}"))
	})
})
