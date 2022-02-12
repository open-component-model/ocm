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
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
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

const TAG = "v1"
const DIGEST_MANIFEST = "3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"
const DIGEST_LAYER = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
const DIGEST_CONFIG = "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"

func defaultManifestFill(n cpi.NamespaceAccess) {
	art, err := n.NewArtefact()
	Expect(err).To(Succeed())
	Expect(art.AddLayer(accessio.BlobAccessForString(MimeTypeOctetStream, "testdata"), nil)).To(Equal(0))
	desc, err := art.Manifest()
	Expect(err).To(Succeed())
	Expect(desc).NotTo(BeNil())

	Expect(desc.Layers[0].Digest).To(Equal(digest.FromString("testdata")))
	Expect(desc.Layers[0].MediaType).To(Equal(MimeTypeOctetStream))
	Expect(desc.Layers[0].Size).To(Equal(int64(8)))

	config := accessio.BlobAccessForString(MimeTypeOctetStream, "{}")
	Expect(n.AddBlob(config)).To(Succeed())
	desc.Config = *artdesc.DefaultBlobDescriptor(config)

	blob, err := n.AddTaggedArtefact(art)
	Expect(err).To(Succeed())
	n.AddTags(blob.Digest(), TAG)
}

var _ = Describe("ctf management", func() {
	var tempfs vfs.FileSystem

	var spec *ctf.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		spec = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessobj.PathFileSystem(tempfs), accessobj.FormatDirectory)
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("instantiate filesystem ctf", func() {
		r, err := ctf.FormatDirectory.Create(oci.DefaultContext(), "test", spec.Options, 0700)
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
			"sha256."+DIGEST_MANIFEST,
			"sha256."+DIGEST_CONFIG,
			"sha256."+DIGEST_LAYER))
	})

	It("instantiate filesystem ctf", func() {
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
			"sha256."+DIGEST_MANIFEST,
			"sha256."+DIGEST_CONFIG,
			"sha256."+DIGEST_LAYER))
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
			"blobs/sha256."+DIGEST_MANIFEST,
			"blobs/sha256."+DIGEST_CONFIG,
			"blobs/sha256."+DIGEST_LAYER))
	})

	Context("manifest", func() {
		It("read from filesystem ctf", func() {
			r, err := spec.Repository(nil, nil)
			Expect(err).To(Succeed())
			Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())
			n, err := r.LookupNamespace("mandelsoft/test")
			defaultManifestFill(n)
			Expect(r.Close()).To(Succeed())

			r, err = ctf.Open(nil, accessobj.ACC_READONLY, "test", 0, accessobj.PathFileSystem(tempfs))
			Expect(err).To(Succeed())
			defer r.Close()

			n, err = r.LookupNamespace("mandelsoft/test")
			Expect(err).To(Succeed())

			art, err := n.GetArtefact("sha256:" + DIGEST_MANIFEST)
			Expect(err).To(Succeed())
			Expect(art.IsManifest()).To(BeTrue())
			blob, err := art.GetBlob("sha256:" + DIGEST_LAYER)
			Expect(err).To(Succeed())
			Expect(blob.Get()).To(Equal([]byte("testdata")))
			Expect(blob.MimeType()).To(Equal(MimeTypeOctetStream))
			blob, err = art.GetBlob("sha256:" + DIGEST_CONFIG)
			Expect(err).To(Succeed())
			Expect(blob.Get()).To(Equal([]byte("{}")))
			Expect(blob.MimeType()).To(Equal(MimeTypeOctetStream))

			art, err = n.GetArtefact(TAG)
			Expect(err).To(Succeed())
			b, err := art.Artefact().ToBlobAccess()
			Expect(err).To(Succeed())
			Expect(b.Digest()).To(Equal(digest.Digest("sha256:" + DIGEST_MANIFEST)))

			_, err = n.GetArtefact("dummy")
			Expect(err).To(Equal(errors.ErrNotFound(cpi.KIND_OCIARTEFACT, "dummy", "mandelsoft/test")))

			Expect(n.AddBlob(accessio.BlobAccessForString("", "dummy"))).To(Equal(accessobj.ErrReadOnly))

			n, err = r.LookupNamespace("mandelsoft/other")
			Expect(err).To(Succeed())
			_, err = n.GetArtefact("sha256:" + DIGEST_MANIFEST)
			Expect(err).To(Equal(errors.ErrNotFound(cpi.KIND_OCIARTEFACT, "sha256:"+DIGEST_MANIFEST, "mandelsoft/other")))
		})
	})
})
