// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociartifact

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
)

func Hint(nv common.NameVersion, locator, repo, version string) string {
	if i := strings.LastIndex(version, "@"); i >= 0 {
		version = version[:i] // remove digest
	}
	repository := repoName(nv, locator)
	if repo != "" {
		if strings.HasPrefix(repo, grammar.RepositorySeparator) {
			repository = repo[1:]
		} else {
			repository = repoName(nv, repo)
		}
	}
	if repository != "" && version != "" {
		if !strings.Contains(repository, ":") {
			repository = fmt.Sprintf("%s:%s", repository, version)
		}
	}
	return repository
}

func repoName(nv common.NameVersion, locator string) string {
	if nv.GetName() == "" {
		return locator
	} else {
		if locator == "" {
			return nv.GetName()
		}
		return fmt.Sprintf("%s/%s", nv.GetName(), locator)
	}
}
