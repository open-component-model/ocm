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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci/ocirepo"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg/componentmapping"
	"github.com/open-component-model/ocm/pkg/mime"
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

		ocispec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
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

		Expect(vers.Close()).To(Succeed())
		Expect(comp.Close()).To(Succeed())
		Expect(repo.Close()).To(Succeed())

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

		Expect(vers.Close()).To(Succeed())
		Expect(comp.Close()).To(Succeed())
		Expect(repo.Close()).To(Succeed())
	})

	It("imports blobs", func() {

		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(ocirepo.NewBlobHandler(base))).New()

		blob := accessio.BlobAccessForString(mime.MIME_OCTET, "anydata")

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
		Expect(acc.GetKind()).To(Equal(localblob.Type))
		l, ok := acc.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(l.LocalReference).To(Equal(blob.Digest().String()))
		Expect(l.GlobalAccess).NotTo(BeNil())

		// check provided global access to be oci blob
		g, err := l.GlobalAccess.Evaluate(DefaultContext)
		Expect(err).To(Succeed())
		o, ok := g.(*ociblob.AccessSpec)
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
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(ocirepo.NewArtefactHandler(base), cpi.ForMimeType(mime))).New()

		// create artefactset
		opts, err := accessio.AccessOptions(nil, accessio.PathFileSystem(tempfs))
		Expect(err).To(Succeed())
		r, err := artefactset.FormatTGZ.Create("test.tgz", opts, 0700)
		Expect(err).To(Succeed())
		testhelper.DefaultManifestFill(r)
		r.Annotate(artefactset.MAINARTEFACT_ANNOTATION, "sha256:"+testhelper.DIGEST_MANIFEST)
		r.Annotate(artefactset.LEGACY_MAINARTEFACT_ANNOTATION, "sha256:"+testhelper.DIGEST_MANIFEST)
		Expect(r.Close()).To(Succeed())

		// create repository
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		defer repo.Close()

		ocirepo := repo.(*genericocireg.Repository).GetOCIRepository()

		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())
		defer comp.Close()
		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())
		defer vers.Close()
		blob := accessio.BlobAccessForFile(mime, "test.tgz", tempfs)

		acc, err := vers.AddBlob(blob, "artefact1", nil)
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(ociartefact.Type))
		o := acc.(*ociartefact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artefact1@sha256:" + testhelper.DIGEST_MANIFEST))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		acc, err = vers.AddBlob(blob, "artefact2:v1", nil)
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(ociartefact.Type))
		o = acc.(*ociartefact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artefact2:v1"))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		Expect(vers.Close()).To(Succeed())
		Expect(comp.Close()).To(Succeed())

		ns, err := ocirepo.LookupNamespace("artefact2")
		Expect(err).To(Succeed())
		defer ns.Close()
		art, err := ns.GetArtefact("v1")
		Expect(err).To(Succeed())
		defer art.Close()
		testhelper.CheckArtefact(art)
		Expect(art.Close()).To(Succeed())
		Expect(ns.Close()).To(Succeed())
		Expect(repo.(*genericocireg.Repository).Close()).To(Succeed())
	})

})
