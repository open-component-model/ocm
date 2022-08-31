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

package transferhandler

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

type TransferOptions interface {
}

type TransferOption interface {
	ApplyTransferOption(TransferOptions) error
}

type TransferHandler interface {
	OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error)

	TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.ElementMeta) (ocm.ComponentVersionAccess, TransferHandler, error)
	TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error)
	TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error)

	HandleTransferResource(r ocm.ResourceAccess, m ocm.AccessMethod, hint string, t ocm.ComponentVersionAccess) error
	HandleTransferSource(r ocm.SourceAccess, m ocm.AccessMethod, hint string, t ocm.ComponentVersionAccess) error
}

func ApplyOptions(set TransferOptions, opts ...TransferOption) error {
	list := errors.ErrListf("transfer options")
	for _, o := range opts {
		list.Add(o.ApplyTransferOption(set))
	}
	return list.Result()
}
