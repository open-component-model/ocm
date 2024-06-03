package maven_test

import (
	"time"

	"github.com/open-component-model/ocm/pkg/maven/maventest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	me "github.com/open-component-model/ocm/pkg/blobaccess/maven"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const (
	MAVEN_PATH            = "/testdata/.m2/repository"
	FAIL_PATH             = "/testdata/.m2/fail"
	MAVEN_CENTRAL_ADDRESS = "repo.maven.apache.org:443"
	MAVEN_CENTRAL         = "https://repo.maven.apache.org/maven2/"
	MAVEN_GROUP_ID        = "maven"
	MAVEN_ARTIFACT_ID     = "maven"
	MAVEN_VERSION         = "1.1"
)

var _ = Describe("blobaccess for maven", func() {

	Context("maven filesystem repository", func() {
		var env *Builder
		var repo *maven.Repository

		BeforeEach(func() {
			env = NewBuilder(maventest.TestData())
			repo = maven.NewFileRepository(MAVEN_PATH, env.FileSystem())
		})

		AfterEach(func() {
			MustBeSuccessful(env.Cleanup())
		})

		It("blobaccess for gav", func() {
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")

			b := Must(me.BlobAccessForMaven(repo, coords.GroupId, coords.ArtifactId, coords.Version, me.WithCachingFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf(
				"sdk-modules-bom-5.7.0.pom",
				"sdk-modules-bom-5.7.0.jar",
				"sdk-modules-bom-5.7.0-random-content.txt",
				"sdk-modules-bom-5.7.0-random-content.json",
				"sdk-modules-bom-5.7.0-sources.jar"))
		})

		It("blobaccess for files with the same classifier", func() {
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier("random-content"))

			b := Must(me.BlobAccessForMavenCoords(repo, coords, me.WithCachingFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-random-content.txt",
				"sdk-modules-bom-5.7.0-random-content.json"))
		})

		It("blobaccess for files with empty classifier", func() {
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier(""))

			b := Must(me.BlobAccessForMavenCoords(repo, coords, me.WithCachingFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0.pom",
				"sdk-modules-bom-5.7.0.jar"))
		})

		It("blobaccess for files with extension", func() {
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithExtension("jar"))

			b := Must(me.BlobAccessForMavenCoords(repo, coords, me.WithCachingFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-sources.jar",
				"sdk-modules-bom-5.7.0.jar"))
		})

		It("blobaccess for files with extension", func() {
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithExtension("txt"))

			b := Must(me.BlobAccessForMavenCoords(repo, coords, me.WithCachingFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-random-content.txt"))
		})

		It("blobaccess for a single file with classifier and extension", func() {
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier("random-content"), maven.WithExtension("json"))

			b := Must(me.BlobAccessForMavenCoords(repo, coords, me.WithCachingFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			Expect(string(Must(b.Get()))).To(Equal(`{"some": "test content"}`))
		})
	})

	Context("maven http repository", func() {
		if PingTCPServer(MAVEN_CENTRAL_ADDRESS, time.Second) == nil {
			var coords *maven.Coordinates
			BeforeEach(func() {
				coords = maven.NewCoordinates(MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION)
			})
			It("blobaccess for gav", func() {
				repo := Must(maven.NewUrlRepository(MAVEN_CENTRAL))
				b := Must(me.BlobAccessForMavenCoords(repo, coords))
				defer Close(b, "blobaccess")
				files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
				Expect(files).To(ConsistOf(
					"maven-1.1-RC1.javadoc.javadoc.jar",
					"maven-1.1-sources.jar",
					"maven-1.1.jar",
					"maven-1.1.pom",
				))
			})
		}
	})
})
