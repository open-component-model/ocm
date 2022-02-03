// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ctf_test

import (
	"archive/tar"
	"compress/gzip"
	"io"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/artefactset"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const MimeTypeOctetStream = "application/octet-stream"

func defaultManifestFill(n cpi.NamespaceAccess) {
	art, err := n.NewArtefact()
	Expect(err).To(Succeed())
	Expect(art.AddLayer(accessio.BlobAccessForData(MimeTypeOctetStream, []byte("testdata")), nil)).To(Equal(0))
	desc, err := art.Manifest()
	Expect(err).To(Succeed())
	Expect(desc).NotTo(BeNil())

	Expect(desc.Layers[0].Digest).To(Equal(digest.FromString("testdata")))
	Expect(desc.Layers[0].MediaType).To(Equal(MimeTypeOctetStream))
	Expect(desc.Layers[0].Size).To(Equal(int64(8)))

	config := accessio.BlobAccessForData(MimeTypeOctetStream, []byte("{}"))
	Expect(n.AddBlob(config)).To(Succeed())
	desc.Config = *artdesc.DefaultBlobDescriptor(config)

	n.AddArtefact(art)
}

var _ = Describe("artefact management", func() {
	var tempfs vfs.FileSystem

	var spec *ctf.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		spec = ctf.NewRepositorySpec("test", accessobj.PathFileSystem(tempfs), accessobj.FormatDirectory)
		spec.PathFileSystem = tempfs
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("instantiate filesystem artefact", func() {
		r, err := spec.Repository(nil, nil)
		Expect(err).To(Succeed())
		Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())

		n, err := r.LookupNamespace("mandelsoft/test")
		defaultManifestFill(n)

		Expect(r.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, "test/"+ctf.ArtefactIndexFileName)).To(BeTrue())

		infos, err := vfs.ReadDir(tempfs, "test/"+artefactset.BlobsDirectoryName)
		Expect(err).To(Succeed())
		blobs := []string{}
		for _, fi := range infos {
			blobs = append(blobs, fi.Name())
		}
		Expect(blobs).To(ContainElements(
			"sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
	})

	It("instantiate tgz artefact", func() {
		ctf.FormatTGZ.ApplyOption(&spec.Options)
		spec.FilePath = "test.tgz"
		r, err := spec.Repository(nil, nil)
		Expect(err).To(Succeed())

		n, err := r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		defaultManifestFill(n)

		Expect(r.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, "test.tgz")).To(BeTrue())

		file, err := tempfs.Open("test.tgz")
		Expect(err).To(Succeed())
		defer file.Close()
		zip, err := gzip.NewReader(file)
		Expect(err).To(Succeed())
		defer zip.Close()
		tr := tar.NewReader(zip)

		files := []string{}
		for {
			header, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				Fail(err.Error())
			}

			switch header.Typeflag {
			case tar.TypeDir:
				Expect(header.Name).To(Equal(artefactset.BlobsDirectoryName))
			case tar.TypeReg:
				files = append(files, header.Name)
			}
		}
		Expect(files).To(ContainElements(
			ctf.ArtefactIndexFileName,
			"blobs/sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"blobs/sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"blobs/sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
	})

	Context("manifest", func() {
		It("read from filesystem artefact", func() {
			r, err := spec.Repository(nil, nil)
			Expect(err).To(Succeed())
			Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())
			n, err := r.LookupNamespace("mandelsoft/test")
			defaultManifestFill(n)
			Expect(r.Close()).To(Succeed())

			r, err = spec.Repository(nil, nil)
			Expect(err).To(Succeed())
			defer r.Close()

			n, err = r.LookupNamespace("mandelsoft/test")
			Expect(err).To(Succeed())

			art, err := n.GetArtefact("sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
			Expect(err).To(Succeed())
			Expect(art.IsManifest()).To(BeTrue())
			blob, err := art.GetBlob("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50")
			Expect(err).To(Succeed())
			Expect(blob.Get()).To(Equal([]byte("testdata")))
			Expect(blob.MimeType()).To(Equal(MimeTypeOctetStream))
		})
	})
	Context("index", func() {

	})
})
