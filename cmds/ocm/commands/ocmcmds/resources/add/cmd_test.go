// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"

func CheckTextResource(env *TestEnv, cd *compdesc.ComponentDescriptor, name string) {
	rblob := accessio.BlobAccessForFile("text/plain", "/testdata/testcontent", env)
	CheckTextResourceBlob(env, cd, name, rblob)
}

func CheckTextResourceWith(env *TestEnv, cd *compdesc.ComponentDescriptor, name, txt string) {
	rblob := accessio.BlobAccessForString("text/plain", txt)
	CheckTextResourceBlob(env, cd, name, rblob)
}

func CheckTextResourceBlob(env *TestEnv, cd *compdesc.ComponentDescriptor, name string, rblob accessio.BlobAccess) {
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
	Expect(r.Version).To(Equal(VERSION))
	Expect(r.Type).To(Equal("PlainText"))

	spec, err := env.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)
	Expect(err).To(Succeed())
	Expect(spec.GetType()).To(Equal(localblob.Type))
	Expect(spec.(*localblob.AccessSpec).LocalReference).To(Equal(common.DigestToFileName(dig)))
	Expect(spec.(*localblob.AccessSpec).MediaType).To(Equal("text/plain"))
}

func Get(blob accessio.BlobAccess, expected []byte) []byte {
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
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds simple text blob", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/resources.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob by cli env file", func() {
		Expect(env.Execute("add", "resources", ARCH, "--settings", "/testdata/settings", "/testdata/resources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds simple text blob by cli variable", func() {
		Expect(env.Execute("add", "resources", ARCH, "CONTENT=testcontent", "/testdata/resources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))

		CheckTextResource(env, cd, "testdata")
	})

	It("adds helm chart", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/helm.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(2))

		r, err := cd.GetResourceByIdentity(metav1.NewIdentity("chart"))
		Expect(err).To(Succeed())
		Expect(r.Type).To(Equal(consts.HelmChart))
		Expect(r.Version).To(Equal(VERSION))

		Expect(r.Access.GetType()).To(Equal(localblob.Type))

		acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
		Expect(err).To(Succeed())
		// sha is always different for helm artefact
		// Expect(acc.(*localblob.AccessSpec).LocalReference).To(Equal("sha256.817db2696ed23f7779a7f848927e2958d2236e5483ad40875274462d8fa8ef9a"))

		blobpath := env.Join(ARCH, comparch.BlobsDirectoryName, common.DigestToFileName(digest.Digest(acc.(*localblob.AccessSpec).LocalReference)))
		blob := accessio.BlobAccessForFile(mime.MIME_GZIP, blobpath, env)

		set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
		Expect(err).To(Succeed())
		defer set.Close()
		art, err := set.GetArtefact(set.GetMain().String())
		Expect(err).To(Succeed())
		m := art.ManifestAccess().GetDescriptor()
		Expect(len(m.Layers)).To(Equal(1))

		blob, err = set.GetBlob(m.Layers[0].Digest)
		Expect(err).To(Succeed())
		reader, err := blob.Reader()
		Expect(err).To(Succeed())

		_, err = loader.LoadArchive(reader)
		Expect(err).To(Succeed())
	})

	It("adds external image", func() {
		Expect(env.Execute("add", "resources", ARCH, "/testdata/image.yaml")).To(Succeed())
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

		Expect(r.Access.GetType()).To(Equal(ociartefact.Type))

		acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociartefact.AccessSpec)(nil))))
		Expect(acc.(*ociartefact.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
	})

	Context("resource by options", func() {
		It("adds simple text blob", func() {
			meta := `
name: testdata
type: PlainText
`
			input := `
type: file
path: testdata/testcontent
mediaType: text/plain
`
			Expect(env.Execute("add", "resources", ARCH, "--resource", meta, "--input", input)).To(Succeed())
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
type: PlainText
`
			input := `
type: file
path: ${CONTENT}
mediaType: text/plain
`
			Expect(env.Execute("add", "resources", ARCH, "CONTENT=testdata/testcontent", "--resource", meta, "--input", input)).To(Succeed())
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
type: ociArtefact
imageReference: ghcr.io/mandelsoft/pause:v0.1.0
`
			Expect(env.Execute("add", "resources", ARCH, "--resource", meta, "--access", access)).To(Succeed())
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

			Expect(r.Access.GetType()).To(Equal(ociartefact.Type))

			acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociartefact.AccessSpec)(nil))))
			Expect(acc.(*ociartefact.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
		})

		It("adds simple text blob with metadata via explicit options", func() {
			input := `
{ "type": "file", "path": "testdata/testcontent", "mediaType": "text/plain" }
`
			Expect(env.Execute("add", "resources", ARCH, "--name", "testdata", "--type", "PlainText", "--input", input)).To(Succeed())
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
type: PlainText
`
			Expect(env.Execute("add", "resources", ARCH, "--resource", meta, "--inputType", "file", "--inputPath", "testdata/testcontent", "--"+options.MediatypeOption.Name(), "text/plain")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResource(env, cd, "testdata")
		})

		It("adds spiff processed text blob by dedicated input options", func() {
			meta := `
name: testdata
type: PlainText
`
			Expect(env.Execute("add", "resources", ARCH, "--resource", meta, "--inputType", "spiff", "--inputPath", "testdata/spiffcontent", "--"+options.MediatypeOption.Name(), "text/plain", "IMAGE=test")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.Resources)).To(Equal(1))

			CheckTextResourceWith(env, cd, "testdata", "data: test\n")
		})

		It("adds external image by options", func() {
			Expect(env.Execute("add", "resources", ARCH,
				"--type", "ociImage",
				"--name", "image",
				"--version", "v0.1.0",
				//	"--external",
				"--accessType", "ociArtefact",
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

			Expect(r.Access.GetType()).To(Equal(ociartefact.Type))

			acc, err := env.OCMContext().AccessSpecForSpec(r.Access)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(acc)).To(Equal(reflect.TypeOf((*ociartefact.AccessSpec)(nil))))
			Expect(acc.(*ociartefact.AccessSpec).ImageReference).To(Equal("ghcr.io/mandelsoft/pause:v0.1.0"))
		})
	})
})
