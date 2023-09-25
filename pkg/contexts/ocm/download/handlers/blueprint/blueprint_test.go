// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprint_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/projectionfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/blueprint"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	tenv "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const (
	COMPONENT = "github.com/compa"
	VERSION   = "1.0.0"
	CTF       = "ctf"
	OCI       = "oci"

	OCIHOST      = "source"
	OCINAMESPACE = "ocm/value"
	OCIVERSION   = "v2.0"

	MIMETYPE            = "testmimetype"
	ARTIFACT_TYPE       = "testartifacttype"
	OCI_ARTIFACT_NAME   = "ociblueprint"
	LOCAL_ARTIFACT_NAME = "localblueprint"
	ARTIFACT_VERSION    = "v1.0.0"

	TESTDATA_PATH = "testdata/blueprint"
	ARCHIVE_PATH  = "archive"
	DOWNLOAD_PATH = "download"
)

var _ = Describe("download blueprint", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(tenv.TestData())

		MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, TESTDATA_PATH)), ARCHIVE_PATH, tarutils.Gzip, env))

		env.OCICommonTransport(OCI, accessio.FormatDirectory, func() {
			env.Namespace(OCINAMESPACE, func() {
				env.Manifest(OCIVERSION, func() {
					env.Config(func() {
						env.BlobStringData(blueprint.CONFIG_MIME_TYPE, "{}")
					})
					env.Layer(func() {
						env.BlobFromFile(blueprint.BLUEPRINT_MIMETYPE, ARCHIVE_PATH)
					})
				})
			})
		})

		testhelper.FakeOCIRepo(env, OCI, OCIHOST)
		env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENT, VERSION, func() {
				env.Resource(OCI_ARTIFACT_NAME, ARTIFACT_VERSION, blueprint.TYPE, v1.ExternalRelation, func() {
					env.Access(ociartifact.New(OCIHOST + ".alias/" + OCINAMESPACE + ":" + OCIVERSION))
				})
				env.Resource(LOCAL_ARTIFACT_NAME, ARTIFACT_VERSION, blueprint.TYPE, v1.LocalRelation, func() {
					env.BlobFromFile(blueprint.BLUEPRINT_MIMETYPE, ARCHIVE_PATH)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})
	DescribeTable("download blueprints", func(index int) {
		src := Must(ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, CTF, 0, env))
		defer Close(src, "source ctf")

		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)

		racc := Must(cv.GetResourceByIndex(index))

		p, buf := common.NewBufferedPrinter()
		ok, path := Must2(download.For(env).Download(p, racc, DOWNLOAD_PATH, env))
		Expect(ok).To(BeTrue())
		Expect(path).To(Equal(DOWNLOAD_PATH))
		Expect(env.FileExists(DOWNLOAD_PATH + "/blueprint.yaml")).To(BeTrue())
		Expect(env.FileExists(DOWNLOAD_PATH + "/test/README.md")).To(BeTrue())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(DOWNLOAD_PATH + ": 2 file(s) with 390 byte(s) written"))
	},
		Entry("oci artifact", 0),
		Entry("local resource", 1),
	)
})
