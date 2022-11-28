// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
		ref, err := utils.GetOCIArtifactRef(ctx, acc)
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
			if err := result.Add(substitutionName(v.Name, "repository", cnt), v.FilePath, v.Repository, repo); err != nil {
				return nil, errors.Wrapf(err, "setting repository for %s", substitutionName(v.Name, "repository", cnt))
			}
		}
		if v.Tag != "" {
			if err := result.Add(substitutionName(v.Name, "tag", cnt), v.FilePath, v.Tag, tag); err != nil {
				return nil, errors.Wrapf(err, "setting tag for %s", substitutionName(v.Name, "tag", cnt))
			}
		}
		if v.Image != "" {
			if err := result.Add(substitutionName(v.Name, "image", cnt), v.FilePath, v.Image, ref); err != nil {
				return nil, errors.Wrapf(err, "setting image for %s", substitutionName(v.Name, "image", cnt))
			}
		}
	}
	return result, nil
}

func substitutionName(name, sub string, cnt int) string {
	if name == "" {
		return ""
	}
	if cnt <= 1 {
		return name
	}
	return name + "-" + sub
}
