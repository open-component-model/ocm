// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/maven"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"

	me "github.com/open-component-model/ocm/pkg/blobaccess/maven"
)

const (
	mvnPATH  = "/testdata/.m2/repository"
	FAILPATH = "/testdata/.m2/fail"
)

var _ = Describe("blobaccess for maven", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(TestData())
	})

	AfterEach(func() {
		MustBeSuccessful(env.Cleanup())
	})

	It("blobaccess for artifact", func() {
		repoUrl := "file://" + mvnPATH
		coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")

		b := Must(me.BlobAccessForMaven(repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, me.WithFileSystem(env.FileSystem())))
		defer Close(b, "blobaccess")
		files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
		Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0.pom", "sdk-modules-bom-5.7.0-random-content.txt",
			"sdk-modules-bom-5.7.0-random-content.json"))
	})

	It("blobaccess for files with the same classifier", func() {
		repoUrl := "file://" + mvnPATH
		coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
			"random-content")

		b := Must(me.BlobAccessForMaven(repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, me.WithClassifier(coords.Classifier),
			me.WithFileSystem(env.FileSystem())))
		defer Close(b, "blobaccess")
		files := Must(tarutils.ListArchiveContentFromReader(Must(b.Reader())))
		Expect(files).To(ConsistOf("sdk-modules-bom-5.7.0-random-content.txt",
			"sdk-modules-bom-5.7.0-random-content.json"))
	})
	It("blobaccess for a single file with classifier and extension", func() {
		repoUrl := "file://" + mvnPATH
		coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
			"random-content", "json")

		b := Must(me.BlobAccessForMaven(repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, me.WithClassifier(coords.Classifier),
			me.WithExtension(coords.Extension), me.WithFileSystem(env.FileSystem())))
		defer Close(b, "blobaccess")
		Expect(string(Must(b.Get()))).To(Equal(`{"some": "test content"}`))
	})
})
