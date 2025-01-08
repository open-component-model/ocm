package maven_test

import (
	"encoding/json"
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/elements"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	me "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/maven"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/tech/maven/maventest"
	mavenblob "ocm.software/ocm/api/utils/blobaccess/maven"
)

const MAVEN_PATH = "/testdata/.m2/repository"

var _ = Describe("blobhandler generic maven tests", func() {
	var env *Builder
	var repo *maven.Repository

	BeforeEach(func() {
		env = NewBuilder(maventest.TestData())
		repo = maven.NewFileRepository(MAVEN_PATH, env.FileSystem())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("Unmarshal upload response Body", func() {
		resp := `{ "repo" : "ocm-mvn-test",
		  			"path" : "/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar",
					"created" : "2024-04-11T15:09:28.920Z",
		  			"createdBy" : "john.doe",
		  			"downloadUri" : "https://ocm.software/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar",
		  			"mimeType" : "application/java-archive",
		  			"size" : "1792",
		  			"checksums" : {
		    			"sha1" : "99d9acac1ff93ac3d52229edec910091af1bc40a",
		    			"md5" : "6cb7520b65d820b3b35773a8daa8368e",
		    			"sha256" : "b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30" },
		  			"originalChecksums" : {
		    			"sha256" : "b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30" },
		  			"uri" : "https://ocm.software/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar" }`
		var body maven.Body
		err := json.Unmarshal([]byte(resp), &body)
		Expect(err).To(BeNil())
		Expect(body.Repo).To(Equal("ocm-mvn-test"))
		Expect(body.DownloadUri).To(Equal("https://ocm.software/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
		Expect(body.Uri).To(Equal("https://ocm.software/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
		Expect(body.MimeType).To(Equal("application/java-archive"))
		Expect(body.Size).To(Equal("1792"))
		Expect(body.Checksums["md5"]).To(Equal("6cb7520b65d820b3b35773a8daa8368e"))
		Expect(body.Checksums["sha1"]).To(Equal("99d9acac1ff93ac3d52229edec910091af1bc40a"))
		Expect(body.Checksums["sha256"]).To(Equal("b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30"))
		Expect(body.Checksums["sha512"]).To(Equal(""))
	})

	It("Upload artifact to file system", func() {
		env.OCMContext().BlobHandlers().Register(me.NewArtifactHandler(me.NewFileConfig("target", env.FileSystem())))
		coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		bacc := Must(mavenblob.BlobAccessForCoords(repo, coords, mavenblob.WithCachingFileSystem(env.FileSystem())))
		defer Close(bacc)
		ocmrepo := composition.NewRepository(env)
		defer Close(ocmrepo)
		cv := composition.NewComponentVersion(env, "acme.org/test", "1.0.0")
		MustBeSuccessful(cv.SetResourceBlob(Must(elements.ResourceMeta("test", resourcetypes.MAVEN_PACKAGE)), bacc, coords.GAV(), nil))
		MustBeSuccessful(ocmrepo.AddComponentVersion(cv))
		l := sliceutils.Transform(Must(vfs.ReadDir(env.FileSystem(), "target/com/sap/cloud/sdk/sdk-modules-bom/5.7.0")),
			func(info os.FileInfo) string {
				return info.Name()
			})
		Expect(l).To(ConsistOf(
			"sdk-modules-bom-5.7.0-random-content.json",
			"sdk-modules-bom-5.7.0-random-content.json.md5",
			"sdk-modules-bom-5.7.0-random-content.json.sha1",
			"sdk-modules-bom-5.7.0-random-content.json.sha256",
			"sdk-modules-bom-5.7.0-random-content.txt",
			"sdk-modules-bom-5.7.0-random-content.txt.md5",
			"sdk-modules-bom-5.7.0-random-content.txt.sha1",
			"sdk-modules-bom-5.7.0-random-content.txt.sha256",
			"sdk-modules-bom-5.7.0-sources.jar",
			"sdk-modules-bom-5.7.0-sources.jar.md5",
			"sdk-modules-bom-5.7.0-sources.jar.sha1",
			"sdk-modules-bom-5.7.0-sources.jar.sha256",
			"sdk-modules-bom-5.7.0.jar",
			"sdk-modules-bom-5.7.0.jar.md5",
			"sdk-modules-bom-5.7.0.jar.sha1",
			"sdk-modules-bom-5.7.0.jar.sha256",
			"sdk-modules-bom-5.7.0.pom",
			"sdk-modules-bom-5.7.0.pom.md5",
			"sdk-modules-bom-5.7.0.pom.sha1",
			"sdk-modules-bom-5.7.0.pom.sha256"))
	})
})
