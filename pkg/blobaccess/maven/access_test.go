// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	me "github.com/open-component-model/ocm/pkg/blobaccess/maven"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/maven"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
	"time"
)

const (
	mvnPATH  = "/testdata/.m2/repository"
	FAILPATH = "/testdata/.m2/fail"

	MAVEN_CENTRAL     = "https://repo.maven.apache.org/maven2/"
	MAVEN_GROUP_ID    = "maven"
	MAVEN_ARTIFACT_ID = "maven"
	MAVEN_VERSION     = "1.1"
)

var _ = Describe("blobaccess for maven", func() {

	Context("maven filesystem repository", func() {
		var env *Builder

		BeforeEach(func() {
			env = NewBuilder(TestData())
		})

		AfterEach(func() {
			MustBeSuccessful(env.Cleanup())
		})

		It("blobaccess for gav", func() {
			repoUrl := "file://" + mvnPATH
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")

			b := Must(me.BlobAccessForMaven(repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, me.WithFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0.pom", "sdk-modules-bom-5.7.0.jar", "sdk-modules-bom-5.7.0-random-content.txt",
				"sdk-modules-bom-5.7.0-random-content.json", "sdk-modules-bom-5.7.0-sources.jar"))
		})

		It("blobaccess for files with the same classifier", func() {
			repoUrl := "file://" + mvnPATH
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier("random-content"))

			b := Must(me.BlobAccessForMavenCoords(repoUrl, coords, me.WithFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-random-content.txt",
				"sdk-modules-bom-5.7.0-random-content.json"))
		})

		It("blobaccess for files with empty classifier", func() {
			repoUrl := "file://" + mvnPATH
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier(""))

			b := Must(me.BlobAccessForMavenCoords(repoUrl, coords, me.WithFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0.pom",
				"sdk-modules-bom-5.7.0.jar"))
		})

		It("blobaccess for files with extension", func() {
			repoUrl := "file://" + mvnPATH
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithExtension("jar"))

			b := Must(me.BlobAccessForMavenCoords(repoUrl, coords, me.WithFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-sources.jar",
				"sdk-modules-bom-5.7.0.jar"))
		})

		It("blobaccess for files with extension", func() {
			repoUrl := "file://" + mvnPATH
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithExtension("txt"))

			b := Must(me.BlobAccessForMavenCoords(repoUrl, coords, me.WithFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
			Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-random-content.txt"))
		})

		It("blobaccess for a single file with classifier and extension", func() {
			repoUrl := "file://" + mvnPATH
			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier("random-content"), maven.WithExtension("json"))

			b := Must(me.BlobAccessForMavenCoords(repoUrl, coords, me.WithFileSystem(env.FileSystem())))
			defer Close(b, "blobaccess")
			Expect(string(Must(b.Get()))).To(Equal(`{"some": "test content"}`))
		})
	})

	Context("maven http repository", func() {
		if PingTCPServer("repo.maven.apache.org:443", time.Second) == nil {
			var coords *maven.Coordinates
			BeforeEach(func() {
				coords = maven.NewCoordinates(MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION)
			})
			It("blobaccess for gav", func() {
				b := Must(me.BlobAccessForMavenCoords(MAVEN_CENTRAL, coords))
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
