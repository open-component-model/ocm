// Copyright 2021 Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compdesc_test

import (
	"github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
)

var _ = Describe("helper", func() {

	It("should inject a new repository context if none is defined", func() {
		cd := &compdesc.ComponentDescriptor{}
		Expect(compdesc.DefaultComponent(cd)).To(Succeed())

		repoCtx := ocireg.NewOCIRegistryRepositorySpec("example.com", "")
		Expect(cd.AddRepositoryContext(repoCtx)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(1))

		Expect(cd.AddRepositoryContext(repoCtx)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(1))

		repoCtx2 := ocireg.NewOCIRegistryRepositorySpec("example.com/dev", "")
		Expect(cd.AddRepositoryContext(repoCtx2)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(2))
	})

})
