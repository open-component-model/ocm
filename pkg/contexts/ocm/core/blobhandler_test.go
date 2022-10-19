// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package core_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/mime"
)

const REPO = "repo"

var IMPL = core.ImplementationRepositoryType{core.CONTEXT_TYPE, REPO}

type BlobHandler struct {
	name string
}

var _ core.BlobHandler = (*BlobHandler)(nil)

func (b BlobHandler) StoreBlob(blob core.BlobAccess, hint string, global core.AccessSpec, ctx core.StorageContext) (core.AccessSpec, error) {
	return nil, fmt.Errorf(b.name)
}

var _ = Describe("blob handler registry test", func() {
	var reg core.BlobHandlerRegistry

	BeforeEach(func() {
		reg = core.NewBlobHandlerRegistry()
	})

	It("priortizes complete specs", func() {
		reg.Register(&BlobHandler{"mine"}, core.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repo"}, core.ForRepo(core.CONTEXT_TYPE, REPO))
		reg.Register(&BlobHandler{"all"}, core.ForRepo(core.CONTEXT_TYPE, REPO), core.ForMimeType(mime.MIME_TEXT))

		h := reg.GetHandler(IMPL, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("all")))
	})

	It("priortizes complete specs", func() {
		reg.Register(&BlobHandler{"mine"}, core.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"repo"}, core.ForRepo(core.CONTEXT_TYPE, REPO))
		reg.Register(&BlobHandler{"all"}, core.ForRepo(core.CONTEXT_TYPE, REPO), core.ForMimeType(mime.MIME_TEXT))
		reg.Register(&BlobHandler{"high"}, core.WithPrio(core.DEFAULT_BLOBHANDLER_PRIO+1))

		h := reg.GetHandler(IMPL, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("high")))
	})

	It("copies registries", func() {
		mine := &BlobHandler{"mine"}
		repo := &BlobHandler{"repo"}
		reg.Register(mine, core.ForMimeType(mime.MIME_TEXT))
		reg.Register(repo, core.ForRepo(core.CONTEXT_TYPE, REPO))

		h := reg.GetHandler(core.ImplementationRepositoryType{core.CONTEXT_TYPE, REPO}, mime.MIME_OCTET)
		Expect(h).To(Equal(core.MultiBlobHandler{repo}))

		copy := reg.Copy()
		new := &BlobHandler{"repo2"}
		copy.Register(new, core.ForRepo(core.CONTEXT_TYPE, REPO))

		h = reg.GetHandler(core.ImplementationRepositoryType{core.CONTEXT_TYPE, REPO}, mime.MIME_OCTET)
		Expect(h).To(Equal(core.MultiBlobHandler{repo}))

		h = copy.GetHandler(core.ImplementationRepositoryType{core.CONTEXT_TYPE, REPO}, mime.MIME_OCTET)
		Expect(h).To(Equal(core.MultiBlobHandler{new}))

	})

})
