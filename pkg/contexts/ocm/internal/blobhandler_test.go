// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/mime"
)

const REPO = "repo"

var (
	IMPL = internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}
	ART  = "myType"
)

type BlobHandler struct {
	name string
}

var _ internal.BlobHandler = (*BlobHandler)(nil)

func (b BlobHandler) StoreBlob(blob internal.BlobAccess, hint string, global internal.AccessSpec, ctx internal.StorageContext) (internal.AccessSpec, error) {
	return nil, fmt.Errorf(b.name)
}

var _ = Describe("blob handler registry test", func() {
	var reg internal.BlobHandlerRegistry

	BeforeEach(func() {
		reg = internal.NewBlobHandlerRegistry()
	})

	It("priortizes complete specs", func() {
		reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
		reg.Register(&BlobHandler{"art"}, internal.ForArtefactType(ART))
		reg.Register(&BlobHandler{"all"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForArtefactType(ART), internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repomime"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))

		h := reg.GetHandler(IMPL, ART, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("all")))
	})

	It("priortizes mime", func() {
		reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
		reg.Register(&BlobHandler{"repomime"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))

		h := reg.GetHandler(IMPL, ART, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("repomime")))
	})

	It("priortizes mime", func() {
		reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
		reg.Register(&BlobHandler{"repomine"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repoart"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForArtefactType(ART))

		h := reg.GetHandler(IMPL, ART, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("repoart")))
	})

	It("priortizes prio", func() {
		reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
		reg.Register(&BlobHandler{"all"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"high"}, internal.WithPrio(internal.DEFAULT_BLOBHANDLER_PRIO+1))

		h := reg.GetHandler(IMPL, ART, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("high")))
	})

	It("copies registries", func() {
		mine := &BlobHandler{"mine"}
		repo := &BlobHandler{"repo"}
		reg.Register(mine, internal.ForMimeType(mime.MIME_TEXT))
		reg.Register(repo, internal.ForRepo(internal.CONTEXT_TYPE, REPO))

		h := reg.GetHandler(internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}, ART, mime.MIME_OCTET)
		Expect(h).To(Equal(internal.MultiBlobHandler{repo}))

		copy := reg.Copy()
		new := &BlobHandler{"repo2"}
		copy.Register(new, internal.ForRepo(internal.CONTEXT_TYPE, REPO))

		h = reg.GetHandler(internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}, ART, mime.MIME_OCTET)
		Expect(h).To(Equal(internal.MultiBlobHandler{repo}))

		h = copy.GetHandler(internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}, ART, mime.MIME_OCTET)
		Expect(h).To(Equal(internal.MultiBlobHandler{new}))

	})

})
