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

package core_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
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

	It("priotizes complete specs", func() {
		reg.RegisterBlobHandler(&BlobHandler{"mine"}, core.ForMimeType(mime.MIME_TEXT))
		reg.RegisterBlobHandler(&BlobHandler{"repo"}, core.ForRepo(core.CONTEXT_TYPE, REPO))
		reg.RegisterBlobHandler(&BlobHandler{"all"}, core.ForRepo(core.CONTEXT_TYPE, REPO), core.ForMimeType(mime.MIME_TEXT))

		h := reg.GetHandler(IMPL, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("all")))
	})

	It("priotizes complete specs", func() {
		reg.RegisterBlobHandler(&BlobHandler{"mine"}, core.ForMimeType(mime.MIME_TEXT))
		reg.RegisterBlobHandler(&BlobHandler{"repo"}, core.ForRepo(core.CONTEXT_TYPE, REPO))
		reg.RegisterBlobHandler(&BlobHandler{"all"}, core.ForRepo(core.CONTEXT_TYPE, REPO), core.ForMimeType(mime.MIME_TEXT))
		reg.RegisterBlobHandler(&BlobHandler{"high"}, core.WithPrio(core.DEFAULT_BLOBHANDLER_PRIO+1))

		h := reg.GetHandler(IMPL, mime.MIME_TEXT)
		Expect(h).NotTo(BeNil())
		_, err := h.StoreBlob(nil, "", nil, nil)
		Expect(err).To(MatchError(fmt.Errorf("high")))
	})
})
