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

package artefactset_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	testenv "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder(testenv.NewEnvironment())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("maps artefact set to repo", func() {
		env.ArtefactSet("/tmp/set", accessio.FormatDirectory, func() {
			env.Manifest("v1", func() {
				env.Config(func() {
					env.BlobStringData(mime.MIME_JSON, "{}")
				})
				env.Layer(func() {
					env.BlobStringData(mime.MIME_OCTET, "testdata")
				})
			})
		})

		spec := artefactset.NewRepositorySpec(accessobj.ACC_READONLY, "/tmp/set", accessio.PathFileSystem(env))

		r, err := cpi.DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		defer Close(r)
		ns, err := r.LookupNamespace("")
		Expect(err).To(Succeed())

		Expect(ns.ListTags()).To(Equal([]string{"v1"}))

		a, err := ns.GetArtefact("v1")
		Expect(err).To(Succeed())

		Expect(a.IsManifest()).To(BeTrue())
		m := a.ManifestAccess()

		cfg, err := m.GetConfigBlob()
		Expect(err).To(Succeed())
		Expect(cfg.Get()).To(Equal([]byte("{}")))

		Expect(len(m.GetDescriptor().Layers)).To(Equal(1))
		blob, err := m.GetBlob(m.GetDescriptor().Layers[0].Digest)
		Expect(err).To(Succeed())
		Expect(blob.Get()).To(Equal([]byte("testdata")))
	})
})
