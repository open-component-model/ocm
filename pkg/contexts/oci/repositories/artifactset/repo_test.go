// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artifactset_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/common/accessobj"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/repositories/artifactset"
	testenv "github.com/open-component-model/ocm/v2/pkg/env"
	"github.com/open-component-model/ocm/v2/pkg/env/builder"
	"github.com/open-component-model/ocm/v2/pkg/mime"
)

var _ = Describe("", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder(testenv.NewEnvironment())
		// ocmlog.Context().AddRule(logging.NewConditionRule(logging.DebugLevel, accessio.ALLOC_REALM))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("maps artifact set to repo", func() {
		env.ArtifactSet("/tmp/set", accessio.FormatDirectory, func() {
			env.Manifest("v1", func() {
				env.Config(func() {
					env.BlobStringData(mime.MIME_JSON, "{}")
				})
				env.Layer(func() {
					env.BlobStringData(mime.MIME_OCTET, "testdata")
				})
			})
		})

		spec, err := artifactset.NewRepositorySpec(accessobj.ACC_READONLY, "/tmp/set", accessio.PathFileSystem(env))
		Expect(err).To(Succeed())

		r, err := cpi.DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		defer Close(r)
		ns, err := r.LookupNamespace("")
		Expect(err).To(Succeed())
		defer Close(ns)

		Expect(ns.ListTags()).To(Equal([]string{"v1"}))

		a, err := ns.GetArtifact("v1")
		Expect(err).To(Succeed())
		defer Close(a)

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
