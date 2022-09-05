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

package oci

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
)

func AsTags(tag string) []string {
	if tag != "" {
		return []string{tag}
	}
	return nil
}

func StandardOCIRef(host, repository, version string) string {
	sep := grammar.TagSeparator
	if ok, _ := artdesc.IsDigest(version); ok {
		sep = grammar.DigestSeparator
	}
	return fmt.Sprintf("%s%s%s%s%s", host, grammar.RepositorySeparator, repository, sep, version)
}

func IsIntermediate(spec RepositorySpec) bool {
	if s, ok := spec.(IntermediateRepositorySpecAspect); ok {
		return s.IsIntermediate()
	}
	return false
}
