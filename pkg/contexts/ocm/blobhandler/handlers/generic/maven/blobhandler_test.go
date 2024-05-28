package maven_test

import (
	"encoding/json"
	"github.com/open-component-model/ocm/pkg/maven/maventest"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	mavenblob "github.com/open-component-model/ocm/pkg/blobaccess/maven"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/generic/maven"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/maven"
)

const MAVEN_PATH = "/testdata/.m2/repository"

var _ = Describe("blobhandler generic maven tests", func() {
	var env *Builder
	var repo *maven.Repository

	BeforeEach(func() {
		env = NewBuilder(maventest.TestData())
		repo = maven.NewFileRepository(MAVEN_PATH, env.FileSystem())
	})

	It("Unmarshal upload response Body", func() {
		resp := `{ "repo" : "ocm-mvn-test",
		  			"path" : "/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar",
					"created" : "2024-04-11T15:09:28.920Z",
		  			"createdBy" : "john.doe",
		  			"downloadUri" : "https://ocm.sofware/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar",
		  			"mimeType" : "application/java-archive",
		  			"size" : "1792",
		  			"checksums" : {
		    			"sha1" : "99d9acac1ff93ac3d52229edec910091af1bc40a",
		    			"md5" : "6cb7520b65d820b3b35773a8daa8368e",
		    			"sha256" : "b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30" },
		  			"originalChecksums" : {
		    			"sha256" : "b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30" },
		  			"uri" : "https://ocm.sofware/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar" }`
		var body maven.Body
		err := json.Unmarshal([]byte(resp), &body)
		Expect(err).To(BeNil())
		Expect(body.Repo).To(Equal("ocm-mvn-test"))
		Expect(body.DownloadUri).To(Equal("https://ocm.sofware/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
		Expect(body.Uri).To(Equal("https://ocm.sofware/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
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
		bacc := Must(mavenblob.BlobAccessForMavenCoords(repo, coords, mavenblob.WithCachingFileSystem(env.FileSystem())))
		defer Close(bacc)
		ocmrepo := composition.NewRepository(env)
		defer Close(ocmrepo)
		cv := composition.NewComponentVersion(env, "acme.org/test", "1.0.0")
		MustBeSuccessful(cv.SetResourceBlob(Must(elements.ResourceMeta("test", resourcetypes.MAVEN_ARTIFACT)), bacc, coords.GAV(), nil))
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
