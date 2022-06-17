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

package comparch

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/errors"
)

func init() {
	cpi.RegisterBlobHandler(NewBlobHandler(), cpi.ForRepo(cpi.CONTEXT_TYPE, comparch.CTFComponentArchiveType))
}

////////////////////////////////////////////////////////////////////////////////

// blobHandler is the default handling to store local blobs as local blobs
type blobHandler struct {
}

func NewBlobHandler() cpi.BlobHandler {
	return &blobHandler{}
}

func (b *blobHandler) StoreBlob(blob cpi.BlobAccess, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	ocmctx := ctx.(storagecontext.StorageContext)

	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	err := ocmctx.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	path := common.DigestToFileName(blob.Digest())
	if compatattr.Get(ctx.GetContext()) {
		return localfsblob.New(path, blob.MimeType()), nil
	} else {
		return localblob.New(path, hint, blob.MimeType(), global), nil
	}
}
