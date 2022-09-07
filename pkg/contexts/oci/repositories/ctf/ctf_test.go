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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/errors"
)

var _ = Describe("ctf management", func() {
	var tempfs vfs.FileSystem

	var spec *ctf.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		spec = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("instantiate filesystem ctf", func() {
		r, err := ctf.FormatDirectory.Create(oci.DefaultContext(), "test", spec.Options, 0700)
		Expect(err).To(Succeed())
		Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())

		n, err := r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		DefaultManifestFill(n)
		Expect(n.Close()).To(Succeed())

		Expect(r.ExistsArtefact("mandelsoft/test", TAG)).To(BeTrue())

		art, err := r.LookupArtefact("mandelsoft/test", TAG)
		Expect(err).To(Succeed())
		Close(art, "art")

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
		Expect(err).To(Succeed())
		DefaultManifestFill(n)

		Expect(n.Close()).To(Succeed())
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
		DefaultManifestFill(n)

		Expect(n.Close()).To(Succeed())
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
			Expect(err).To(Succeed())
			DefaultManifestFill(n)
			Expect(n.Close()).To(Succeed())
			Expect(r.Close()).To(Succeed())

			r, err = ctf.Open(nil, accessobj.ACC_READONLY, "test", 0, accessio.PathFileSystem(tempfs))
			Expect(err).To(Succeed())
			defer r.Close()

			n, err = r.LookupNamespace("mandelsoft/test")
			Expect(err).To(Succeed())

			art, err := n.GetArtefact("sha256:" + DIGEST_MANIFEST)
			Expect(err).To(Succeed())
			CheckArtefact(art)
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
