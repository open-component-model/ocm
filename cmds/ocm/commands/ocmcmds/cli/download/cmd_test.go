// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download_test

import (
	"bytes"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	env2 "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/cli/download"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
)

const CTF = "/tmp/ctf"
const OCM = "/testdata/bin/ocm" + download.EXECUTABLE_SUFFIX

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var cur string

	BeforeEach(func() {
		env = NewTestEnv(env2.ModifiableTestData())
		cur = os.Getenv("PATH") // TODO: introduce environ abstraction
		os.Setenv("PATH", "/testdata/other"+string(os.PathListSeparator)+"/testdata/bin")

		env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
			env.Component(download.COMPONENT, func() {
				env.Version("1.0.0", func() {
					env.Provider("ocm.software")
					env.Resource(download.RESOURCE, "1.0.0", resourcetypes.EXECUTABLE, metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_OCTET, "ocm script")
						env.ExtraIdentity(consts.ExecutableOperatingSystem, runtime.GOOS)
						env.ExtraIdentity(consts.ExecutableArchitecture, runtime.GOARCH)
					})
				})
			})
		})
	})

	AfterEach(func() {
		os.Setenv("PATH", cur)
		env.Cleanup()
	})

	It("downloads ocm cli", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "cli", "--repo", CTF)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
updating OCM CLI command at /testdata/bin/ocm
/testdata/bin/ocm: 10 byte(s) written
`))
		Expect(env.FileExists(OCM)).To(BeTrue())
		Expect(env.ReadFile(OCM)).To(Equal([]byte("ocm script")))

		Expect(Must(env.Stat(OCM)).Mode() & os.ModePerm).To(BeNumerically("==", 0o755))
	})

})
