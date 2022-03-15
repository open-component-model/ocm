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

package genericocireg_test

import (
	"reflect"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/artefactset"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/testhelper"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	storagecontext "github.com/gardener/ocm/pkg/ocm/blobhandler/oci"
	"github.com/gardener/ocm/pkg/ocm/blobhandler/oci/ocirepo"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg/componentmapping"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = ocm.New()

const COMPONENT = "github.com/mandelsoft/ocm"
const TESTBASE = "testbase.de"

var _ = Describe("component repository mapping", func() {
	var tempfs vfs.FileSystem

	var ocispec oci.RepositorySpec
	var spec *genericocireg.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		ocispec = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		spec = genericocireg.NewRepositorySpec(ocispec, nil)

	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("creates a dummy component", func() {
		repo, err := DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())

		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		Expect(repo.(*genericocireg.Repository).Close()).To(Succeed())

		// access it again
		repo, err = DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())

		ok, err := repo.ExistsComponentVersion(COMPONENT, "v1")
		Expect(err).To(Succeed())
		Expect(ok).To(BeTrue())

		comp, err = repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err = comp.LookupVersion("v1")
		Expect(err).To(Succeed())
		Expect(vers.GetDescriptor()).To(Equal(compdesc.New(COMPONENT, "v1")))

		Expect(repo.Close()).To(Succeed())
	})

	It("imports blobs", func() {

		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().RegisterBlobHandler(ocirepo.NewBlobHandler(base))).New()

		blob := accessio.BlobAccessForString(testhelper.MimeTypeOctetStream, "anydata")

		// create repository
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())

		acc, err := vers.AddBlob(blob, "", nil)
		Expect(err).To(Succeed())

		// check provided actual access to be local blob
		Expect(acc.GetKind()).To(Equal(accessmethods.LocalBlobType))
		l, ok := acc.(*accessmethods.LocalBlobAccessSpec)
		Expect(ok).To(BeTrue())
		Expect(l.LocalReference).To(Equal(blob.Digest().String()))
		Expect(l.GlobalAccess).NotTo(BeNil())

		// check provided global access to be oci blob
		o, ok := l.GlobalAccess.(*accessmethods.OCIBlobAccessSpec)
		Expect(ok).To(BeTrue())
		Expect(o.Digest).To(Equal(blob.Digest()))
		Expect(o.Reference).To(Equal(TESTBASE + "/" + componentmapping.ComponentDescriptorNamespace + "/" + COMPONENT))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())
	})

	It("imports artefact", func() {
		mime := artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest) + "+tar+gzip"
		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().RegisterBlobHandler(ocirepo.NewArtefactHandler(base), cpi.ForMimeType(mime))).New()

		// create artefactset
		r, err := artefactset.FormatTGZ.Create("test.tgz", accessio.AccessOptions(accessio.PathFileSystem(tempfs)), 0700)
		Expect(err).To(Succeed())
		testhelper.DefaultManifestFill(r)
		r.Annotate(artefactset.MAINARTEFACT_ANNOTATION, "sha256:"+testhelper.DIGEST_MANIFEST)
		Expect(r.Close()).To(Succeed())

		// create repository
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())

		blob := accessio.BlobAccessForFile(mime, "test.tgz", tempfs)

		acc, err := vers.AddBlob(blob, "artefact1", nil)
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(accessmethods.OCIRegistryType))
		o := acc.(*accessmethods.OCIRegistryAccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artefact1@sha256:" + testhelper.DIGEST_MANIFEST))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		acc, err = vers.AddBlob(blob, "artefact2:v1", nil)
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(accessmethods.OCIRegistryType))
		o = acc.(*accessmethods.OCIRegistryAccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artefact2:v1"))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		ocirepo := repo.(*genericocireg.Repository).GetOCIRepository()

		ns, err := ocirepo.LookupNamespace("artefact2")
		Expect(err).To(Succeed())
		art, err := ns.GetArtefact("v1")
		Expect(err).To(Succeed())
		testhelper.CheckArtefact(art)
		Expect(repo.(*genericocireg.Repository).Close()).To(Succeed())
	})

})
