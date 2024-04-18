package mvn

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Maven Test Environment", func() {

	It("GAV, GroupPath, Path", func() {
		artifact := &Artifact{
			GroupId:    "ocm.software",
			ArtifactId: "hello-ocm",
			Version:    "0.0.1",
			Extension:  "jar",
		}
		Expect(artifact.GAV()).To(Equal("ocm.software:hello-ocm:0.0.1"))
		Expect(artifact.GroupPath()).To(Equal("ocm/software"))
		Expect(artifact.Path()).To(Equal("ocm/software/hello-ocm/0.0.1/hello-ocm-0.0.1.jar"))
	})

	It("ClassifierExtensionFrom", func() {
		artifact := &Artifact{
			GroupId:    "ocm.software",
			ArtifactId: "hello-ocm",
			Version:    "0.0.1",
		}
		artifact.ClassifierExtensionFrom("hello-ocm-0.0.1.pom")
		Expect(artifact.Classifier).To(Equal(""))
		Expect(artifact.Extension).To(Equal("pom"))

		artifact.ClassifierExtensionFrom("hello-ocm-0.0.1-tests.jar")
		Expect(artifact.Classifier).To(Equal("tests"))
		Expect(artifact.Extension).To(Equal("jar"))

		artifact.ArtifactId = "apache-maven"
		artifact.Version = "3.9.6"
		artifact.ClassifierExtensionFrom("apache-maven-3.9.6-bin.tar.gz")
		Expect(artifact.Classifier).To(Equal("bin"))
		Expect(artifact.Extension).To(Equal("tar.gz"))
	})

	It("parse GAV", func() {
		gav := "org.apache.commons:commons-compress:1.26.1:cyclonedx:xml"
		artifact := ArtifactFromHint(gav)
		Expect(artifact.GroupId).To(Equal("org.apache.commons"))
		Expect(artifact.ArtifactId).To(Equal("commons-compress"))
		Expect(artifact.Version).To(Equal("1.26.1"))
		Expect(artifact.Classifier).To(Equal("cyclonedx"))
		Expect(artifact.Extension).To(Equal("xml"))
		Expect(artifact.Path()).To(Equal("org/apache/commons/commons-compress/1.26.1/commons-compress-1.26.1-cyclonedx.xml"))
	})

	/*
			It("Body", func() {
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
				var body Body
				err := json.Unmarshal([]byte(resp), &body)
				Expect(err).To(BeNil())
				Expect(body.Repo).To(Equal("ocm-mvn-test"))
				Expect(body.Path).To(Equal("/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
				Expect(body.DownloadUri).To(Equal("https://ocm.sofware/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
				Expect(body.Uri).To(Equal("https://ocm.sofware/repository/ocm-mvn-test/open-component-model/hello-ocm/0.0.2/hello-ocm-0.0.2.jar"))
				Expect(body.MimeType).To(Equal("application/java-archive"))
				Expect(body.Size).To(Equal("1792"))
				Expect(body.Checksums["md5"]).To(Equal("6cb7520b65d820b3b35773a8daa8368e"))
				Expect(body.Checksums["sha1"]).To(Equal("99d9acac1ff93ac3d52229edec910091af1bc40a"))
				Expect(body.Checksums["sha256"]).To(Equal("b19dcd275f72a0cbdead1e5abacb0ef25a0cb55ff36252ef44b1178eeedf9c30"))
				Expect(body.Checksums["sha512"]).To(Equal(""))
			})
	*/
})
