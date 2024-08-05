package localize_test

import (
	"bytes"
	"compress/gzip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/download/handlers/dirtree"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/ocmutils/localize"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
)

const RESOURCE_TYPE = "gitOpsTemplate"

var _ = Describe("image value mapping", func() {
	const (
		ARCHIVE   = "archive.ctf"
		COMPONENT = "github.com/comp"
		VERSION   = "1.0.0"
		IMAGE     = "image"
		TEMPLATE  = "template"
	)

	var (
		repo     ocm.Repository
		cv       ocm.ComponentVersionAccess
		env      *builder.Builder
		template *bytes.Buffer
	)

	func() {
		template = bytes.NewBuffer(nil)
		w := gzip.NewWriter(template)
		err := tarutils.PackFsIntoTar(osfs.New(), "testdata", w, tarutils.TarFileSystemOptions{})
		w.Close()
		if err != nil {
			panic(err)
		}
	}()

	BeforeEach(func() {
		env = builder.NewBuilder(nil)

		// register downloader for new archive type.
		download.For(env).Register(dirtree.New(), download.ForArtifactType(RESOURCE_TYPE))

		env.OCMCommonTransport(ARCHIVE, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider("mandelsoft")
					env.Resource(IMAGE, "", "Spiff", v1.LocalRelation, func() {
						env.ModificationOptions(ocm.SkipVerify())
						env.Digest("fake", "sha256", "fake")
						env.Access(ociartifact.New("ghcr.io/mandelsoft/test:v1"))
					})
					env.Resource(TEMPLATE, "", RESOURCE_TYPE, v1.LocalRelation, func() {
						env.BlobData(mime.MIME_TGZ, template.Bytes())
					})
				})
			})
		})

		var err error
		repo, err = ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCHIVE, 0, env)
		Expect(err).To(Succeed())

		cv, err = repo.LookupComponentVersion(COMPONENT, VERSION)
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		Expect(cv.Close()).To(Succeed())
		Expect(repo.Close()).To(Succeed())
		vfs.Cleanup(env)
	})

	It("uses image ref data from component version", func() {
		rules := UnmarshalInstRules(`
templateResource:
  resource:
    name: template
localizationRules:
  - file: dir/manifest1.yaml
    image: manifest.value1
    resource:
      name: image
configRules:
  - file: dir/manifest1.yaml
    path: manifest.value2
    value: (( settings.value ))
configTemplate:
  defaults:
    value: default
  settings: (( merge(defaults, values) ))
configScheme:
  type: object
  properties:
    values:
      type: object
      properties:
        value:
          type: string
      additionalProperties: false
  additionalProperties: false
`)
		config := []byte(`
values:
  value: mine
`)
		fs := memoryfs.New()
		err := localize.Instantiate(rules, cv, nil, config, fs, RESOURCE_TYPE)
		Expect(err).To(Succeed())
		CheckYAMLFile("dir/manifest1.yaml", fs, `
manifest:
  value1: ghcr.io/mandelsoft/test:v1
  value2: mine
`)
	})
})
