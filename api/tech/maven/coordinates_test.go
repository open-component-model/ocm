package maven_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/optionutils"

	me "ocm.software/ocm/api/tech/maven"
)

var _ = Describe("Maven Test Environment", func() {
	It("GAV, GroupPath, FilePath", func() {
		coords := me.NewCoordinates("ocm.software", "hello-ocm", "0.0.1", me.WithExtension("jar"))
		Expect(coords.GAV()).To(Equal("ocm.software:hello-ocm:0.0.1"))
		Expect(coords.GroupPath()).To(Equal("ocm/software"))
		Expect(coords.FilePath()).To(Equal("ocm/software/hello-ocm/0.0.1/hello-ocm-0.0.1.jar"))
	})

	It("SetClassifierExtensionBy", func() {
		coords := me.NewCoordinates("ocm.software", "hello-ocm", "0.0.1")
		MustBeSuccessful(coords.SetClassifierExtensionBy("hello-ocm-0.0.1.pom"))
		Expect(coords.Classifier).ToNot(BeNil())
		Expect(optionutils.AsValue(coords.Classifier)).To(Equal(""))
		Expect(optionutils.AsValue(coords.Extension)).To(Equal("pom"))

		MustBeSuccessful(coords.SetClassifierExtensionBy("hello-ocm-0.0.1-tests.jar"))
		Expect(optionutils.AsValue(coords.Classifier)).To(Equal("tests"))
		Expect(optionutils.AsValue(coords.Extension)).To(Equal("jar"))

		coords.ArtifactId = "apache-me"
		coords.Version = "3.9.6"
		MustBeSuccessful(coords.SetClassifierExtensionBy("apache-me-3.9.6-bin.tar.gz"))
		Expect(optionutils.AsValue(coords.Classifier)).To(Equal("bin"))
		Expect(optionutils.AsValue(coords.Extension)).To(Equal("tar.gz"))
	})

	It("parse GAV", func() {
		gav := "org.apache.commons:commons-compress:1.26.1:cyclonedx:xml"
		coords, err := me.Parse(gav)
		Expect(err).To(BeNil())
		Expect(coords.String()).To(Equal(gav))
		Expect(coords.GroupId).To(Equal("org.apache.commons"))
		Expect(coords.ArtifactId).To(Equal("commons-compress"))
		Expect(coords.Version).To(Equal("1.26.1"))
		Expect(optionutils.AsValue(coords.Classifier)).To(Equal("cyclonedx"))
		Expect(optionutils.AsValue(coords.Extension)).To(Equal("xml"))
		Expect(coords.FilePath()).To(Equal("org/apache/commons/commons-compress/1.26.1/commons-compress-1.26.1-cyclonedx.xml"))
	})
})
