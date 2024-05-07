package download_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/grammar"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/mime"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const OUT = "/tmp/res"

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
