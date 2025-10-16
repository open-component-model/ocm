package add_test

import (
	"strings"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	mavenacc "ocm.software/ocm/api/ocm/extensions/accessmethods/maven"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/tech/maven/maventest"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	MAVEN_CENTRAL_ADDRESS = "repo.maven.apache.org:443"
	MAVEN_CENTRAL         = "https://repo.maven.apache.org/maven2/"
	MAVEN_GROUP_ID        = "maven"
	MAVEN_ARTIFACT_ID     = "maven"
	MAVEN_VERSION         = "1.1"
)

const (
	ARCH      = "/tmp/ctf"
	DEST_ARCH = "/tmp/ctf-dest"
	VERSION   = "1.0.0"
	COMPONENT = "ocm.software/demo/test"
	OUT       = "/tmp/res"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData(), maventest.TestData("/maven/testdata"))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("upload maven package from localblob during transfer", func() {
		coords := maven.NewCoordinates(maventest.GROUP_ID, maventest.ARTIFACT_ID, maventest.VERSION)
		Expect(env.Execute("add", "cv", "-fc", "--file", ARCH, "testdata/components.yaml")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		repo := Must(ctf.Open(env, ctf.ACC_READONLY, ARCH, 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)
		Expect(len(cv.GetDescriptor().Resources)).To(Equal(1))
		acc := Must(env.OCMContext().AccessSpecForSpec(cv.GetDescriptor().Resources[0].Access))
		Expect(acc.IsLocal(env.OCMContext())).To(BeTrue())
		Expect(acc.(*localblob.AccessSpec).ReferenceName).To(Equal(strings.Join([]string{maventest.GROUP_ID, maventest.ARTIFACT_ID, maventest.VERSION}, ":")))

		Expect(env.Execute("transfer", "ctf", ARCH, DEST_ARCH, "--uploader", "ocm/mavenPackage=file://localhost/mavenrepo")).To(Succeed())
		Expect(env.DirExists(DEST_ARCH)).To(BeTrue())
		Expect(env.DirExists("/mavenrepo/" + coords.GavPath())).To(BeTrue())
		mavenrepo := maven.NewFileRepository("/mavenrepo", env.FileSystem())
		Expect(mavenrepo.GavFiles(coords, nil)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-random-content.json: 5
sdk-modules-bom-5.7.0-random-content.txt: 5
sdk-modules-bom-5.7.0-sources.jar: 5
sdk-modules-bom-5.7.0.jar: 5
sdk-modules-bom-5.7.0.pom: 5`))
	})

	It("upload maven package from localblob during component composition", func() {
		coords := maven.NewCoordinates(maventest.GROUP_ID, maventest.ARTIFACT_ID, maventest.VERSION)
		Expect(env.Execute("add", "cv", "-fc", "--file", ARCH, "testdata/components.yaml", "--uploader", "ocm/mavenPackage=file://localhost/mavenrepo")).To(Succeed())
		Expect(env.DirExists(ARCH)).To(BeTrue())
		repo := Must(ctf.Open(env, ctf.ACC_READONLY, ARCH, 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)
		Expect(len(cv.GetDescriptor().Resources)).To(Equal(1))
		acc := Must(env.OCMContext().AccessSpecForSpec(cv.GetDescriptor().Resources[0].Access))
		Expect(acc.IsLocal(env.OCMContext())).To(BeFalse())
		Expect(acc.GetKind()).To(Equal(mavenacc.Type))
		Expect(acc.(*mavenacc.AccessSpec).GAV()).To(Equal(strings.Join([]string{maventest.GROUP_ID, maventest.ARTIFACT_ID, maventest.VERSION}, ":")))

		Expect(env.DirExists("/mavenrepo/" + coords.GavPath())).To(BeTrue())
		mavenrepo := maven.NewFileRepository("/mavenrepo", env.FileSystem())
		Expect(mavenrepo.GavFiles(coords, nil)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-random-content.json: 5
sdk-modules-bom-5.7.0-random-content.txt: 5
sdk-modules-bom-5.7.0-sources.jar: 5
sdk-modules-bom-5.7.0.jar: 5
sdk-modules-bom-5.7.0.pom: 5`))
	})

	Context("maven http repository", func() {
		if PingTCPServer(MAVEN_CENTRAL_ADDRESS, time.Second) == nil {
			var coords *maven.Coordinates
			BeforeEach(func() {
				coords = maven.NewCoordinates(MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION)
			})
			It("upload maven package from access method", func() {
				Expect(env.Execute("add", "cv", "-fc", "--file", ARCH, "testdata/components2.yaml")).To(Succeed())
				Expect(env.DirExists(ARCH)).To(BeTrue())
				repo := Must(ctf.Open(env, ctf.ACC_READONLY, ARCH, 0, env))
				defer Close(repo)
				cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
				defer Close(cv)
				Expect(len(cv.GetDescriptor().Resources)).To(Equal(1))
				acc := Must(env.OCMContext().AccessSpecForSpec(cv.GetDescriptor().Resources[0].Access))
				Expect(acc.IsLocal(env.OCMContext())).To(BeFalse())
				Expect(acc.(ocm.HintProvider).GetReferenceHint(cv)).To(Equal(coords.GAV()))

				Expect(env.Execute("transfer", "ctf", ARCH, DEST_ARCH, "--copy-resources", "--uploader", "ocm/mavenPackage=file://localhost/mavenrepo")).To(Succeed())
				Expect(env.DirExists(DEST_ARCH)).To(BeTrue())
				Expect(env.DirExists("/mavenrepo/" + coords.GavPath())).To(BeTrue())
				mavenrepo := maven.NewFileRepository("/mavenrepo", env.FileSystem())
				Expect(mavenrepo.GavFiles(coords, nil)).To(YAMLEqual(`
maven-1.1-RC1.javadoc.javadoc.jar: 5
maven-1.1-sources.jar: 5
maven-1.1.jar: 5
maven-1.1.pom: 5`))
			})
		}
	})
})
