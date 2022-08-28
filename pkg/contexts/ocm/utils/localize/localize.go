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

package localize

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Localize(mappings []Localization, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (Substitutions, error) {
	var result Substitutions
	ctx := cv.GetContext()

	for i, v := range mappings {
		name := "image mapping"
		if v.Name != "" {
			name = fmt.Sprintf("%s %q", name, v.Name)
		}
		acc, rcv, err := utils.ResolveResourceReference(cv, v.ResourceReference, resolver)
		if err != nil {
			return nil, errors.ErrNotFoundWrap(err, "mapping", fmt.Sprintf("%d (%s)", i+1, &v.ResourceReference))
		}
		rcv.Close()
		ref, err := utils.GetOCIArtefactRef(ctx, acc)
		if err != nil {
			return nil, errors.Wrapf(err, "mapping %d: cannot resolve resource %s to an OCI Reference", i+1, v)
		}
		ix := strings.Index(ref, ":")
		if ix < 0 {
			ix = strings.Index(ref, "@")
			if ix < 0 {
				return nil, errors.Wrapf(err, "mapping %d: image tag or digest missing (%s)", i+1, ref)
			}
		}
		repo := ref[:ix]
		tag := ref[ix+1:]

		cnt := 0
		if v.Repository != "" {
			cnt++
		}
		if v.Tag != "" {
			cnt++
		}
		if v.Image != "" {
			cnt++
		}
		if cnt == 0 {
			return nil, fmt.Errorf("no substitution target given for %s", name)
		}

		if v.Repository != "" {
			result.Add(substName(v.Name, "repository", cnt), v.FilePath, v.Repository, repo)
		}
		if v.Tag != "" {
			result.Add(substName(v.Name, "tag", cnt), v.FilePath, v.Tag, tag)
		}
		if v.Image != "" {
			result.Add(substName(v.Name, "image", cnt), v.FilePath, v.Image, ref)
		}
	}
	return result, nil
}

func substName(name, sub string, cnt int) string {
	if name == "" {
		return ""
	}
	if cnt <= 1 {
		return name
	}
	return name + "-" + sub
}
