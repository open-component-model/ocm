// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package create_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv3 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/ocm.software/v3alpha1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates default comp arch", func() {
		plabels := metav1.Labels{}
		plabels.Set("email", "info@mandelsoft.de")
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "--provider", "mandelsoft",
			"l1=value", "l2={\"name\":\"value\"}", "-p", "email=info@mandelsoft.de")).To(Succeed())
		Expect(env.DirExists("component-archive")).To(BeTrue())
		data, err := env.ReadFile("component-archive/" + comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Name).To(Equal("test.de/x"))
		Expect(cd.Version).To(Equal("v1"))
		Expect(string(cd.Provider.Name)).To(Equal("mandelsoft"))
		Expect(cd.Provider.Labels).To(Equal(plabels))
		Expect(cd.Labels).To(Equal(metav1.Labels{
			{
				Name:  "l1",
				Value: []byte("\"value\""),
			},
			{
				Name:  "l2",
				Value: []byte("{\"name\":\"value\"}"),
			},
		}))
	})

	It("creates comp arch", func() {

		plabels := metav1.Labels{}
		plabels.Set("email", "info@mandelsoft.de")
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "--provider", "mandelsoft", "--file", "/tmp/ca",
			"l1=value", "l2={\"name\":\"value\"}", "-p", "email=info@mandelsoft.de")).To(Succeed())
		Expect(env.DirExists("/tmp/ca")).To(BeTrue())
		data, err := env.ReadFile("/tmp/ca/" + comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Name).To(Equal("test.de/x"))
		Expect(cd.Version).To(Equal("v1"))
		Expect(string(cd.Provider.Name)).To(Equal("mandelsoft"))
		Expect(cd.Provider.Labels).To(Equal(plabels))
		Expect(cd.Labels).To(Equal(metav1.Labels{
			{
				Name:  "l1",
				Value: []byte("\"value\""),
			},
			{
				Name:  "l2",
				Value: []byte("{\"name\":\"value\"}"),
			},
		}))
	})

	It("creates comp arch with "+compdescv3.SchemaVersion, func() {

		plabels := metav1.Labels{}
		plabels.Set("email", "info@mandelsoft.de")
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "--provider", "mandelsoft", "--file", "/tmp/ca",
			"l1=value", "l2={\"name\":\"value\"}", "-p", "email=info@mandelsoft.de", "-S", compdescv3.SchemaVersion)).To(Succeed())
		Expect(env.DirExists("/tmp/ca")).To(BeTrue())
		data, err := env.ReadFile("/tmp/ca/" + comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Metadata.ConfiguredVersion).To(Equal(compdescv3.GroupVersion))
		Expect(cd.Name).To(Equal("test.de/x"))
		Expect(cd.Version).To(Equal("v1"))
		Expect(string(cd.Provider.Name)).To(Equal("mandelsoft"))
		Expect(cd.Provider.Labels).To(Equal(plabels))
		Expect(cd.Labels).To(Equal(metav1.Labels{
			{
				Name:  "l1",
				Value: []byte("\"value\""),
			},
			{
				Name:  "l2",
				Value: []byte("{\"name\":\"value\"}"),
			},
		}))
	})
})
