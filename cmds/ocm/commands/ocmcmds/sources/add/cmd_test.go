// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"

func CheckArchiveSource(env *TestEnv, cd *compdesc.ComponentDescriptor, name string) {

	r, err := cd.GetSourceByIdentity(metav1.NewIdentity(name))
	Expect(err).To(Succeed())
	Expect(r.Version).To(Equal(VERSION))
	Expect(r.Type).To(Equal("git"))
	spec, err := env.OCMContext().AccessSpecForSpec(r.Access)
	Expect(err).To(Succeed())
	Expect(spec.GetType()).To(Equal(localblob.Type))
	Expect(spec.(*localblob.AccessSpec).MediaType).To(Equal(mime.MIME_TGZ))

	local := spec.(*localblob.AccessSpec).LocalReference
	env.Logger().Info("local", "local", local)

	bpath := env.Join(ARCH, comparch.BlobsDirectoryName, local)
	Expect(env.FileExists(bpath)).To(BeTrue())
	file, err := env.Open(bpath)
	Expect(err).To(Succeed())
	defer file.Close()

	gz, err := gzip.NewReader(file)
	Expect(err).To(Succeed())
	tr := tar.NewReader(gz)
	files := []string{}
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			Expect(err).To(Succeed())
		}

		switch header.Typeflag {
		case tar.TypeReg, tar.TypeDir:
			files = append(files, header.Name)
		}
	}
	sort.Strings(files)
	Expect(files).To(Equal([]string{"settings", "testcontent"}))
}

func CheckTextSource(env *TestEnv, cd *compdesc.ComponentDescriptor, name string) {
	rblob := accessio.BlobAccessForFile(mime.MIME_TEXT, "/testdata/testcontent", env)
	dig := rblob.Digest()
	data, err := rblob.Get()
	Expect(err).To(Succeed())
	bpath := env.Join(ARCH, comparch.BlobsDirectoryName, common.DigestToFileName(dig))
	Expect(env.FileExists(bpath)).To(BeTrue())
	Expect(env.ReadFile(bpath)).To(Equal(data))

	r, err := cd.GetSourceByIdentity(metav1.NewIdentity(name))
	Expect(err).To(Succeed())
	Expect(r.Version).To(Equal(VERSION))
	Expect(r.Type).To(Equal("git"))
	spec, err := env.OCMContext().AccessSpecForSpec(r.Access)
	Expect(err).To(Succeed())
	Expect(spec.GetType()).To(Equal(localblob.Type))
	Expect(spec.(*localblob.AccessSpec).LocalReference).To(Equal(common.DigestToFileName(dig)))
	Expect(spec.(*localblob.AccessSpec).MediaType).To(Equal(mime.MIME_TEXT))
}

var _ = Describe("Add sources", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds simple text blob", func() {
		Expect(env.Execute("add", "sources", ARCH, "/testdata/sources.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Sources)).To(Equal(2))

		CheckTextSource(env, cd, "testdata")
		CheckArchiveSource(env, cd, "myothersrc")
	})

	It("adds simple text blob by cli env file", func() {
		Expect(env.Execute("add", "sources", ARCH, "--settings", "/testdata/settings", "/testdata/sources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Sources)).To(Equal(2))

		CheckTextSource(env, cd, "testdata")
	})

	It("adds simple text blob by cli variable", func() {
		Expect(env.Execute("add", "sources", ARCH, "CONTENT=testcontent", "/testdata/sources.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Sources)).To(Equal(2))

		CheckTextSource(env, cd, "testdata")
	})
})
