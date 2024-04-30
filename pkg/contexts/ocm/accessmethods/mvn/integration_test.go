//go:build integration
// +build integration

package mvn_test

import (
	"crypto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("online accessmethods.mvn.AccessSpec integration tests", func() {
	var env *Builder
	var cv ocm.ComponentVersionAccess

	BeforeEach(func() {
		env = NewBuilder(TestData())
		cv = &cpi.DummyComponentVersionAccess{env.OCMContext()}
	})

	AfterEach(func() {
		env.Cleanup()
	})

	// https://repo1.maven.org/maven2/com/sap/cloud/sdk/sdk-modules-bom/5.7.0
	It("one single pom only", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		files, err := acc.GavFiles(cv.GetContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(files["sdk-modules-bom-5.7.0.pom"]).To(Equal(crypto.SHA1))
	})
	It("GetPackageMeta - com.sap.cloud.sdk", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		meta, err := acc.GetPackageMeta(ocm.DefaultContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(meta.Bin).To(HavePrefix("file://"))
		Expect(meta.Bin).To(ContainSubstring("mvn-sdk-modules-bom-5.7.0-"))
		Expect(meta.Bin).To(HaveSuffix(".tar.gz"))
		Expect(meta.Hash).To(Equal("345fe2e640663c3cd6ac87b7afb92e1c934f665f75ddcb9555bc33e1813ef00b"))
		Expect(meta.HashType).To(Equal(crypto.SHA256))
	})

	// https://repo1.maven.org/maven2/org/apache/maven/apache-maven/3.9.6
	It("apache-maven, with bin + tar.gz etc.", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "org.apache.maven", "apache-maven", "3.9.6")
		Expect(acc).ToNot(BeNil())
		Expect(acc.BaseUrl()).To(Equal("https://repo1.maven.org/maven2/org/apache/maven/apache-maven/3.9.6"))
		files, err := acc.GavFiles(cv.GetContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(8))
		Expect(files["apache-maven-3.9.6-src.zip"]).To(Equal(crypto.SHA512))
		Expect(files["apache-maven-3.9.6.pom"]).To(Equal(crypto.SHA1))
	})

	// https://repo1.maven.org/maven2/com/sap/cloud/environment/servicebinding/java-sap-vcap-services/0.10.4
	It("accesses local artifact", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "com.sap.cloud.environment.servicebinding", "java-sap-vcap-services", "0.10.4")
		meta, err := acc.GetPackageMeta(ocm.DefaultContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(meta.Bin).To(HavePrefix("file://"))
		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		Expect(m.MimeType()).To(Equal(mime.MIME_TGZ))
		/* manually also tested with repos:
		- https://repo1.maven.org/maven2/org/apache/commons/commons-compress/1.26.1/  // cyclonedx
		- https://repo1.maven.org/maven2/cn/afternode/commons/commons/1.6/ // gradle module!
		*/
	})
})
