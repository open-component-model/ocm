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

package utils

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/errors"
)

func GetOCIArtefactRef(ctx ocm.Context, r ocm.ResourceAccess) (string, error) {
	acc, err := r.Access()
	if err != nil {
		return "", err
	}

	if localblob.Is(acc) {
		g := acc.(*localblob.AccessSpec).GlobalAccess
		if g != nil {
			acc, err = ctx.AccessSpecForSpec(g)
			if err != nil {
				return "", errors.Wrapf(err, "global access spec")
			}
		}
	}
	if ociartefact.Is(acc) {
		return acc.(*ociartefact.AccessSpec).ImageReference, nil
	}
	return "", errors.Newf("cannot map access to external image reference")
}
