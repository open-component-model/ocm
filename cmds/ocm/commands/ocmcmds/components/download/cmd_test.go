package download_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/grammar"
	. "ocm.software/ocm/api/ocm/testhelper"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH      = "/tmp/ctf"
	PROVIDER  = "mandelsoft"
	VERSION   = "v1"
	COMPONENT = "github.com/mandelsoft/test"
	OUT       = "/tmp/res"
)

var _ = Describe("Download Component Version", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("download single component version from ctf file", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, S_TESTDATA)
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "component", "-O", OUT, "--repo", ARCH, COMPONENT+grammar.VersionSeparator+VERSION)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: downloaded
`))
		Expect(env.DirExists(OUT)).To(BeTrue())
		Expect(env.ReadFile(vfs.Join(env, OUT, comparch.BlobsDirectoryName, "sha256."+D_TESTDATA))).To(Equal([]byte(S_TESTDATA)))

		cd := `
component:
  componentReferences: []
  name: github.com/mandelsoft/test
  provider: mandelsoft
  repositoryContexts: []
  resources:
  - access:
      localReference: sha256.${value}
      mediaType: text/plain
      type: localBlob
    digest:
      value: ${value}
      normalisationAlgorithm: ${normalisationAlgorithm}
      hashAlgorithm: ${hashAlgorithm}
    name: testdata
    relation: local
    type: ${type}
    version: v1
  sources: []
  version: v1
meta:
  schemaVersion: v2
`
		Expect(env.ReadFile(vfs.Join(env, OUT, comparch.ComponentDescriptorFileName))).To(YAMLEqual(cd,
			MergeSubst(SubstFrom(DS_TESTDATA), SubstList("type", resourcetypes.PLAIN_TEXT))))
	})
})
