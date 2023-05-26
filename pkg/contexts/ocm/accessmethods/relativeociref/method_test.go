// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package relativeociref_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/relativeociref"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/finalizer"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

const COMP = "acme.org/compo"
const COMPVERS = "v1.0.0"
const RES = "ref"

var _ = Describe("Method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses artifact", func() {
		var finalize finalizer.Finalizer
		Defer(finalize.Finalize)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})
		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCMCommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, COMPVERS, func() {
				env.Resource(RES, COMPVERS, "testtyp", v1.LocalRelation, func() {
					env.Access(relativeociref.New(OCINAMESPACE + ":" + OCIVERSION))
				})
			})
		})

		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, OCIPATH, 0, env))
		finalize.Close(repo)
		vers := Must(repo.LookupComponentVersion(COMP, COMPVERS))
		finalize.Close(vers)
		res := Must(vers.GetResourceByIndex(0))
		m := Must(res.AccessMethod())
		finalize.Close(m)
		data := Must(m.Get())
		Expect(len(data)).To(Equal(628))
		Expect(m.(accessio.DigestSource).Digest().String()).To(Equal("sha256:0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f"))
	})
})
