package download_test

import (
	"bytes"
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	env2 "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/tarutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	RESOURCE = "archive"
)

var _ = Describe("image dowanload with dirtree", func() {
	var env *TestEnv

	cfg := Must(json.Marshal(ociv1.ImageConfig{}))

	BeforeEach(func() {
		env = NewTestEnv(env2.TestData())

		MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, "testdata/layers/0")), "layer0.tgz", tarutils.Gzip, env))
		MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, "testdata/layers/1")), "layer1.tgz", tarutils.Gzip, env))

		env.ArtifactSet("image.set", accessio.FormatTGZ, func() {
			env.Manifest(VERSION, func() {
				env.Config(func() {
					env.BlobData(ociv1.MediaTypeImageConfig, cfg)
				})
				env.Layer(func() {
					env.BlobFromFile(ociv1.MediaTypeImageLayerGzip, "layer0.tgz")
				})
				env.Layer(func() {
					env.BlobFromFile(ociv1.MediaTypeImageLayerGzip, "layer1.tgz")
				})
			})
			env.Annotation(artifactset.MAINARTIFACT_ANNOTATION, VERSION)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Resource(RESOURCE, VERSION, resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
					env.BlobFromFile(artifactset.MediaType(ociv1.MediaTypeImageManifest), "image.set")
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("downloads as directory", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "resources", "--downloader", "ocm/dirtree:"+resourcetypes.OCI_IMAGE, "-O", OUT, "--repo", ARCH, COMP+":"+VERSION, RESOURCE)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: 2 file(s) with 25 byte(s) written
`))

		Expect(Must(vfs.IsDir(env, OUT))).To(BeTrue())
		data := Must(vfs.ReadFile(env, OUT+"/testfile"))
		Expect(string(data)).To(StringEqualWithContext("testdata\n"))
		data = Must(vfs.ReadFile(env, OUT+"/dir/nestedfile"))
		Expect(string(data)).To(StringEqualWithContext("other test data\n"))
	})

	It("downloads as archive", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "resources", "--downloader", "ocm/dirtree:"+resourcetypes.OCI_IMAGE+"=asArchive: true", "-O", OUT, "--repo", ARCH, COMP+":"+VERSION, RESOURCE)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: 3584 byte(s) written
`))

		MustBeSuccessful(env.MkdirAll("result", 0o700))
		resultfs := Must(projectionfs.New(env, "result"))
		MustBeSuccessful(tarutils.ExtractArchiveToFs(resultfs, OUT, env))

		data := Must(vfs.ReadFile(env, "result/testfile"))
		Expect(string(data)).To(StringEqualWithContext("testdata\n"))
		data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
		Expect(string(data)).To(StringEqualWithContext("other test data\n"))
	})
})
