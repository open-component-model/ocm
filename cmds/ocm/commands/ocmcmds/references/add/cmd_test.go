// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
)

const ARCH = "/tmp/ca"
const VERSION = "v1.1.1"
const REF = "github.com/mandelsoft/ref"

func CheckReference(env *TestEnv, cd *compdesc.ComponentDescriptor, name string, add ...func(compdesc.ComponentReference)) {
	r, err := cd.GetComponentReferenceByIdentity(metav1.NewIdentity(name))
	ExpectWithOffset(1, err).To(Succeed())
	Expect(r.Version).To(Equal(VERSION))
	Expect(r.ComponentName).To(Equal(REF))
	for _, a := range add {
		a(r)
	}
}

var _ = Describe("Add references", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds simple ref", func() {
		Expect(env.Execute("add", "references", ARCH, "/testdata/references.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds simple ref by cli env file", func() {
		Expect(env.Execute("add", "references", ARCH, "--settings", "/testdata/settings", "/testdata/references.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds simple ref by cli variable", func() {
		Expect(env.Execute("add", "references", ARCH, "VERSION=v1.1.1", "/testdata/references.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	Context("reference by options", func() {
		It("adds simple ref", func() {
			Expect(env.Execute("add", "references", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			CheckReference(env, cd, "testdata")
		})

		It("mixed specification", func() {
			spec := `
labels:
- name: test
  value: value
`
			Expect(env.Execute("add", "references", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION, "--reference", spec)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			labels := metav1.Labels{}
			labels.Set("test", "value")
			CheckReference(env, cd, "testdata", func(r compdesc.ComponentReference) {
				ExpectWithOffset(2, r.GetLabels()).To(Equal(labels))
			})
		})

		It("overrides in mixed specification", func() {
			spec := `
name: bla
labels:
- name: test
  value: value
`
			Expect(env.Execute("add", "references", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION, "--reference", spec)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			labels := metav1.Labels{}
			labels.Set("test", "value")
			CheckReference(env, cd, "testdata", func(r compdesc.ComponentReference) {
				ExpectWithOffset(2, r.GetLabels()).To(Equal(labels))
			})
		})

	})
})
