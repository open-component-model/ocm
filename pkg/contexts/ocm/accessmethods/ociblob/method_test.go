// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociblob_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/contexts/oci/testhelper"
	. "github.com/open-component-model/ocm/v2/pkg/env"
	. "github.com/open-component-model/ocm/v2/pkg/env/builder"

	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/grammar"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/accessmethods/ociblob"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/cpi"
)

const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

var _ = Describe("Method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses artifact", func() {
		var desc *artdesc.Descriptor
		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			desc = OCIManifest1(env)
		})

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		spec := ociblob.New(OCIHOST+".alias"+grammar.RepositorySeparator+OCINAMESPACE, desc.Digest, "", -1)

		m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
		Expect(err).To(Succeed())

		blob, err := m.Get()
		Expect(err).To(Succeed())

		Expect(string(blob)).To(Equal("manifestlayer"))
	})
})
