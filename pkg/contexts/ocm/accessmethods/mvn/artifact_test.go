package mvn

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Maven Test Environment", func() {

	It("GAV, GroupPath, FileName", func() {
		artifact := &Artifact{
			GroupId:    "ocm.software",
			ArtifactId: "hello-ocm",
			Version:    "0.0.1",
			Extension:  "jar",
		}
		Expect(artifact.GAV()).To(Equal("ocm.software:hello-ocm:0.0.1"))
		Expect(artifact.GroupPath()).To(Equal("ocm/software"))
		Expect(artifact.FileName()).To(Equal("ocm/software/hello-ocm/0.0.1/hello-ocm-0.0.1.jar"))
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
		Expect(artifact.FileName()).To(Equal("org/apache/commons/commons-compress/1.26.1/commons-compress-1.26.1-cyclonedx.xml"))
	})
})
