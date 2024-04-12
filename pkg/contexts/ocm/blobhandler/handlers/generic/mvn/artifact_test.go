package mvn

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Maven Test Environment", func() {

	It("GAV, GroupPath, Path", func() {
		artifact := &Artifact{
			GroupId:    "ocm.software",
			ArtifactId: "hello-ocm",
			Version:    "0.0.1",
			Packaging:  "jar",
		}
		Expect(artifact.GAV()).To(Equal("ocm.software:hello-ocm:0.0.1"))
		Expect(artifact.GroupPath()).To(Equal("ocm/software"))
		Expect(artifact.Path()).To(Equal("ocm/software/hello-ocm/0.0.1/hello-ocm-0.0.1.jar"))
	})

	It("Body", func() {
		resp := `{ "repo" : "ocm-mvn-test",
  			"path" : "/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar",
			"created" : "2024-04-11T15:09:28.920Z",
  			"createdBy" : "d057539",
  			"downloadUri" : "https://int.repositories.cloud.sap/artifactory/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar",
  			"mimeType" : "application/java-archive",
  			"size" : "1792",
  			"checksums" : {
    			"sha1" : "99d9acac1ff93ac3d52229edec910091af1bc40a",
    			"md5" : "6cb7520b65d820b3b35773a8daa8368e",
    			"sha256" : "b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30" },
  			"originalChecksums" : {
    			"sha256" : "b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30" },
  			"uri" : "https://int.repositories.cloud.sap/artifactory/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar" }`
		var body Body
		err := json.Unmarshal([]byte(resp), &body)
		Expect(err).To(BeNil())
		Expect(body.Repo).To(Equal("ocm-mvn-test"))
		Expect(body.Path).To(Equal("/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
		Expect(body.DownloadUri).To(Equal("https://int.repositories.cloud.sap/artifactory/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
		Expect(body.Uri).To(Equal("https://int.repositories.cloud.sap/artifactory/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
		Expect(body.MimeType).To(Equal("application/java-archive"))
		Expect(body.Size).To(Equal("1792"))
		Expect(body.Checksums.Md5).To(Equal("6cb7520b65d820b3b35773a8daa8368e"))
		Expect(body.Checksums.Sha1).To(Equal("99d9acac1ff93ac3d52229edec910091af1bc40a"))
		Expect(body.Checksums.Sha256).To(Equal("b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30"))
		Expect(body.Checksums.Sha512).To(Equal(""))
	})

})
