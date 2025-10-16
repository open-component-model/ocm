package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	C1  = "github.com/gardener/landscaper"
	C11 = "github.com/gardener/landscaper/container-deployer"
	C12 = "github.com/gardener/landscaper/helm-deployer"
	C13 = "github.com/gardener/landscaper/manifest-deployer"
	C14 = "github.com/gardener/landscaper/mock-deployer"
)

const (
	CS = "github.com/gardener/landscaper-service"
	CI = "github.com/gardener/landscaper-instance"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(C11, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("container-deployer-blueprint", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
			env.Component(C12, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("helm-deployer-blueprint", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
			env.Component(C13, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("manifest-deployer-blueprint", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
			env.Component(C14, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("mock-deployer-blueprint", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})

			env.Component(C1, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("c1-testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
					env.Reference("mock-deployer", C14, VERSION)
					env.Reference("manifest-deployer", C13, VERSION)
					env.Reference("helm-deployer", C11, VERSION)
					env.Reference("container-deployer", C12, VERSION)
				})
			})

			env.Component(CI, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("installation-blueprint", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
					env.Reference("landscaper", C1, VERSION)
				})
			})

			env.Component(CS, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("landscaper-service-blueprint", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
					env.Reference("landscaper-instance", CI, VERSION)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists all components in a deep structure", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "resources", "--recursive", "-o", "tree", "--repo", ARCH, CS+":"+VERSION)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT                                                     NAME                         VERSION IDENTITY TYPE      RELATION
└─ github.com/gardener/landscaper-service                                                  v1                         
   ├─                                                         landscaper-service-blueprint v1               PlainText local
   └─ github.com/gardener/landscaper-instance                 landscaper-instance          v1                         
      ├─                                                      installation-blueprint       v1               PlainText local
      └─ github.com/gardener/landscaper                       landscaper                   v1                         
         ├─                                                   c1-testdata                  v1               PlainText local
         ├─ github.com/gardener/landscaper/container-deployer helm-deployer                v1                         
         │  └─                                                container-deployer-blueprint v1               PlainText local
         ├─ github.com/gardener/landscaper/helm-deployer      container-deployer           v1                         
         │  └─                                                helm-deployer-blueprint      v1               PlainText local
         ├─ github.com/gardener/landscaper/manifest-deployer  manifest-deployer            v1                         
         │  └─                                                manifest-deployer-blueprint  v1               PlainText local
         └─ github.com/gardener/landscaper/mock-deployer      mock-deployer                v1                         
            └─                                                mock-deployer-blueprint      v1               PlainText local
`))
	})
})
