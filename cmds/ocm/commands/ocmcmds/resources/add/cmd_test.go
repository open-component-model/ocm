package add_test

import (
	"fmt"
	"net/http"
	"reflect"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart/loader"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH    = "/tmp/ca"
	VERSION = "v1"
	OCIPATH = "/tmp/oci"
	OCIHOST = "ghcr.io"
)

func CheckTextResource(env *TestEnv, cd *compdesc.ComponentDescriptor, name string, ff ...func(r compdesc.Resource)) {
	rblob := blobaccess.ForFile(mime.MIME_TEXT, "/testdata/testcontent", env)
	CheckTextResourceBlob(env, cd, name, rblob, ff...)
}

func CheckTextResourceWith(env *TestEnv, cd *compdesc.ComponentDescriptor, name, txt string) {
	rblob := blobaccess.ForString(mime.MIME_TEXT, txt)
	CheckTextResourceBlob(env, cd, name, rblob)
}

func CheckTextResourceBlob(env *TestEnv, cd *compdesc.ComponentDescriptor, name string, rblob blobaccess.BlobAccess, ff ...func(r compdesc.Resource)) {
	dig := rblob.Digest()
	data, err := rblob.Get()
	Expect(err).To(Succeed())

	list, err := vfs.ReadDir(env, env.Join(ARCH, comparch.BlobsDirectoryName))
	found := map[string]string{}
	blobname := common.DigestToFileName(dig)
	for _, e := range list {
		data, err := env.ReadFile(env.Join(ARCH, comparch.BlobsDirectoryName, e.Name()))
		Expect(err).To(Succeed())
		found[e.Name()] = string(data)
	}
	Expect(err).To(Succeed())
	if content, ok := found[blobname]; !ok {
		Fail(fmt.Sprintf("expected blob %s not found, but %v", blobname, found))
	} else {
		Expect(content).To(Equal(string(data)))
	}

	r, err := cd.GetResourceByIdentity(metav1.NewIdentity(name))
	Expect(err).To(Succeed())
	if len(ff) > 0 {
		for _, f := range ff {
			f(r)
		}
	} else {
		Expect(r.Version).To(Equal(VERSION))
		Expect(r.Type).To(Equal(resourcetypes.PLAIN_TEXT))
	}

	spec, err := env.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)
	Expect(err).To(Succeed())
	Expect(spec.GetType()).To(Equal(localblob.Type))
	Expect(spec.(*localblob.AccessSpec).LocalReference).To(Equal(common.DigestToFileName(dig)))
	Expect(spec.(*localblob.AccessSpec).MediaType).To(Equal("text/plain"))
}

func Get(blob blobaccess.BlobAccess, expected []byte) []byte {
	data, err := blob.Get()
	ExpectWithOffset(1, err).To(Succeed())
	if expected != nil {
		ExpectWithOffset(1, string(data)).To(Equal(string(expected)))
	}
	return data
}

var _ = Describe("Add resources", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())

		// fake OCI registry
		FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1For(env.Builder, "mandelsoft/pause", "v0.1.0")
		})

		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds simple text blob", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/resources.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob with explicit version info", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/resources2.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata", func(r compdesc.Resource) {
			Expect(r.Relation).To(Equal(metav1.LocalRelation))
			Expect(r.Version).To(Equal("3.3.3"))
			Expect(r.Type).To(Equal(resourcetypes.PLAIN_TEXT))
		})
	})

	It("adds simple text blob with direct archive file", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/resources.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob by cli env file", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "--settings", "/testdata/settings", "/testdata/resources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob by cli variable", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "CONTENT=testcontent", "/testdata/resources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds duplicate text blob", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/dupresources.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(2))
	})

	It("add helm chart from repo", func() {
		resp, err := http.Get("https://charts.helm.sh/stable")
		if err == nil { // only if connected to internet
			resp.Body.Close()
			fmt.Fprintf(GinkgoWriter, "helm executed\n")
			Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/helmref.yaml")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			r, err := cd.GetResourceByIdentity(metav1.NewIdentity("chart"))
			Expect(err).To(Succeed())
			Expect(r.Type).To(Equal(resourcetypes.HELM_CHART))
			Expect(r.Version).To(Equal("3.0.8"))
		} else {
			fmt.Fprintf(GinkgoWriter, "helm test skipped\n")
		}
	})

	DescribeTable("adds helm chart", func(rsc string) {
		Expect(env.Execute("add", "resources", "--skip-digest-generation", "--file", ARCH, rsc)).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(2))

		r, err := cd.GetResourceByIdentity(metav1.NewIdentity("chart"))
		Expect(err).To(Succeed())
		Expect(r.Type).To(Equal(resourcetypes.HELM_CHART))
		Expect(r.Version).To(Equal(VERSION))

		Expect(r.Access.GetType()).To(Equal(localblob.Type))

		acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
		Expect(err).To(Succeed())
		Expect(acc.(*localblob.AccessSpec).ReferenceName).To(Equal("test.de/x/mandelsoft/testchart:0.1.0"))
		// sha is always different for helm artifact
		// Expect(acc.(*localblob.AccessSpec).LocalReference).To(Equal("sha256.817db2696ed23f7779a7f848927e2958d2236e5483ad40875274462d8fa8ef9a"))

		blobpath := env.Join(ARCH, comparch.BlobsDirectoryName, common.DigestToFileName(digest.Digest(acc.(*localblob.AccessSpec).LocalReference)))
		blob := blobaccess.ForFile(mime.MIME_GZIP, blobpath, env)

		set, err := artifactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
		Expect(err).To(Succeed())
		defer Close(set, "artefactset")
		art, err := set.GetArtifact(set.GetMain().String())
		Expect(err).To(Succeed())
		defer Close(art, "artefact")
		m := art.ManifestAccess().GetDescriptor()
		Expect(len(m.Layers)).To(Equal(1))

		_, blobd, err := set.GetBlobData(m.Layers[0].Digest)
		Expect(err).To(Succeed())
		defer Close(blobd, "blob")
		reader, err := blobd.Reader()
		Expect(err).To(Succeed())
		defer reader.Close()
		_, err = loader.LoadArchive(reader)
		Expect(err).To(Succeed())
	},
		Entry("flat", "/testdata/helm.yaml"),
		Entry("up", "/testdata/nested/helm.yaml"),
	)

	It("adds external image", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/image.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		r, err := cd.GetResourceByIdentity(metav1.NewIdentity("image"))
		Expect(err).To(Succeed())
		Expect(r.Type).To(Equal("ociImage"))
		Expect(r.Version).To(Equal("v0.1.0"))
		Expect(r.Relation).To(Equal(metav1.ResourceRelation("external")))
		Expect(r.GetDigest()).To(Equal(DS_OCIMANIFEST1))
		Expect(r.Access.GetType()).To(Equal(ociartifact.Type))

		acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociartifact.AccessSpec)(nil))))
		Expect(acc.(*ociartifact.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
	})

	It("re-adds external image", func() {
		Expect(env.Execute("add", "resources", "--skip-digest-generation", "--file", ARCH, "/testdata/helm2.yaml")).To(Succeed())
		Expect(env.Execute("add", "resources", "-R", "--skip-digest-generation", "--file", ARCH, "/testdata/helm2.yaml")).To(Succeed())

		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))
	})

	It("rejects duplicate hints", func() {
		Expect(env.Execute("add", "resources", "--skip-digest-generation", "--file", ARCH, "/testdata/helm.yaml")).To(Succeed())
		ExpectError(env.Execute("add", "resources", "--skip-digest-generation", "--file", ARCH, "/testdata/helm2.yaml")).To(MatchError("cannot add resource \"chart2\"(/testdata/helm2.yaml[1][1]): reference name (hint) \"test.de/x/mandelsoft/testchart:0.1.0\" with base media type application/vnd.oci.image.manifest.v1 already used for resource chart:v1"))
	})

	Context("resource by options", func() {
		It("adds simple text blob", func() {
			meta := `
name: testdata
type: plainText
`
			input := `
type: file
path: testdata/testcontent
mediaType: text/plain
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--input", input)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResource(env, cd, "testdata")
		})

		It("adds simple text blob by cli variable", func() {
			meta := `
name: testdata
type: plainText
`
			input := `
type: file
path: ${CONTENT}
mediaType: text/plain
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "CONTENT=testdata/testcontent", "--resource", meta, "--input", input)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResource(env, cd, "testdata")
		})

		It("adds external image", func() {
			meta := `
type: ociImage
name: image
version: v0.1.0
relation: external
`
			access := `
type: ociArtifact
imageReference: ghcr.io/mandelsoft/pause:v0.1.0
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--access", access)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			r, err := cd.GetResourceByIdentity(metav1.NewIdentity("image"))
			Expect(err).To(Succeed())
			Expect(r.Type).To(Equal("ociImage"))
			Expect(r.Version).To(Equal("v0.1.0"))
			Expect(r.Relation).To(Equal(metav1.ResourceRelation("external")))

			Expect(r.Access.GetType()).To(Equal(ociartifact.Type))

			acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociartifact.AccessSpec)(nil))))
			Expect(acc.(*ociartifact.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
		})

		It("adds simple text blob with metadata via explicit options", func() {
			input := `
{ "type": "file", "path": "testdata/testcontent", "mediaType": "text/plain" }
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--name", "testdata", "--type", "plainText", "--input", input)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResource(env, cd, "testdata")
		})

		It("adds simple text blob by dedicated input options", func() {
			meta := `
name: testdata
type: plainText
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "file", "--inputPath", "testdata/testcontent", "--"+options.MediatypeOption.GetName(), "text/plain")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResource(env, cd, "testdata")
		})

		It("fail for non-matching input options", func() {
			meta := `
name: testdata
type: plainText
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "file", "--inputHelmRepository=x", "--inputPath", "testdata/testcontent", "--"+options.MediatypeOption.GetName(), "text/plain")).To(MatchError(`resource (by options): input specification: option "inputHelmRepository" given, but not possible for input type file`))
		})

		It("adds spiff processed text blob by dedicated input options", func() {
			meta := `
name: testdata
type: plainText
`
			Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "spiff", "--inputPath", "testdata/spiffcontent", "--"+options.MediatypeOption.GetName(), "text/plain", "IMAGE=test")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResourceWith(env, cd, "testdata", "data: test\n")
		})

		It("adds external image by options", func() {
			Expect(env.Execute("add", "resources", "--file", ARCH,
				"--type", "ociImage",
				"--name", "image",
				"--version", "v0.1.0",
				//	"--external",
				"--accessType", "ociArtifact",
				"--reference", "ghcr.io/mandelsoft/pause:v0.1.0")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			r, err := cd.GetResourceByIdentity(metav1.NewIdentity("image"))
			Expect(err).To(Succeed())
			Expect(r.Type).To(Equal("ociImage"))
			Expect(r.Version).To(Equal("v0.1.0"))
			Expect(r.Relation).To(Equal(metav1.ResourceRelation("external")))

			Expect(r.Access.GetType()).To(Equal(ociartifact.Type))

			acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociartifact.AccessSpec)(nil))))
			Expect(acc.(*ociartifact.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
		})
	})
})
