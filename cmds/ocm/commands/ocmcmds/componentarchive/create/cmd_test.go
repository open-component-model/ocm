package create_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	compdescv3 "ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
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

	It("creates comp arch with "+compdescv3.VersionName, func() {
		plabels := metav1.Labels{}
		plabels.Set("email", "info@mandelsoft.de")
		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", "v1", "--provider", "mandelsoft", "--file", "/tmp/ca",
			"l1=value", "l2={\"name\":\"value\"}", "-p", "email=info@mandelsoft.de", "-S", compdescv3.VersionName)).To(Succeed())
		Expect(env.DirExists("/tmp/ca")).To(BeTrue())
		data, err := env.ReadFile("/tmp/ca/" + comparch.ComponentDescriptorFileName)
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(cd.Metadata.ConfiguredVersion).To(Equal(compdescv3.SchemaVersion))
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
