// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
)

func Check(env *TestEnv) {
	repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, "ctf", 0, env))
	defer Close(repo)
	cv := Must(repo.LookupComponentVersion("ocm.software/demo/test", "1.0.0"))
	defer Close(cv)
	cd := cv.GetDescriptor()

	var plabels metav1.Labels
	MustBeSuccessful(plabels.Set("city", "Karlsruhe"))

	var clabels metav1.Labels
	MustBeSuccessful(clabels.Set("purpose", "test"))

	Expect(string(cd.Provider.Name)).To(Equal("ocm.software"))
	Expect(cd.Provider.Labels).To(Equal(plabels))
	Expect(cd.Labels).To(Equal(clabels))

	r := Must(cv.GetResource(metav1.Identity{"name": "data"}))
	data := Must(ocm.ResourceData(r))
	Expect(string(data)).To(Equal("!stringdata"))
}

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates ctf and adds component", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", "ctf", "testdata/component.yaml")).To(Succeed())
		Expect(env.DirExists("ctf")).To(BeTrue())
		Check(env)
	})

	It("creates ctf and adds components", func() {
		Expect(env.Execute("add", "c", "-fc", "--file", "ctf", "--version", "1.0.0", "testdata/components.yaml")).To(Succeed())
		Expect(env.DirExists("ctf")).To(BeTrue())
		Check(env)
	})
})
