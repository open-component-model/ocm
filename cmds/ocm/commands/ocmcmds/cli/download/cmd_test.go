package download_test

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	env2 "ocm.software/ocm/api/helper/env"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extraid"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/cli/download"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	CTF = "/tmp/ctf"
	OCM = "/testdata/bin/ocm" + download.EXECUTABLE_SUFFIX
)

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
						env.ExtraIdentity(extraid.ExecutableOperatingSystem, runtime.GOOS)
						env.ExtraIdentity(extraid.ExecutableArchitecture, runtime.GOARCH)
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
		Expect(env.CatchOutput(buf).Execute("download", "cli", "--path", "--repo", CTF)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
updating OCM CLI command at /testdata/bin/ocm
/testdata/bin/ocm: 10 byte(s) written
`))
		Expect(env.FileExists(OCM)).To(BeTrue())
		Expect(env.ReadFile(OCM)).To(Equal([]byte("ocm script")))

		Expect(Must(env.Stat(OCM)).Mode() & os.ModePerm).To(BeNumerically("==", 0o755))
	})

	It("downloads ocm cli", func() {
		fmt.Printf("%s\n", Must(os.Executable()))
	})
})
