package add_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/goutils/testutils"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/selectors/refsel"
)

const (
	ARCH    = "/tmp/ca"
	VERSION = "v1.1.1"
	REF     = "github.com/mandelsoft/ref"
)

func CheckReference(env *TestEnv, cd *compdesc.ComponentDescriptor, name string, add ...func(compdesc.Reference)) {
	rs, _ := cd.SelectReferences(refsel.Name(name))
	if len(rs) != 1 {
		Fail(fmt.Sprintf("%d reference(s) with name %s found", len(rs), name), 1)
	}
	r := rs[0]
	ExpectWithOffset(1, r.Version).To(Equal(VERSION))
	ExpectWithOffset(1, r.ComponentName).To(Equal(REF))
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
		Expect(env.Execute("add", "references", "--file", ARCH, "/testdata/references.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds simple ref with extra identity", func() {
		Expect(env.Execute("add", "references", "--file", ARCH, "/testdata/referenceswithid.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata", func(r compdesc.Reference) {
			Expect(r.ExtraIdentity).To(Equal(metav1.Identity{"purpose": "test", "label": "local"}))
		})
	})

	It("adds simple ref by cli env file", func() {
		Expect(env.Execute("add", "references", "--file", ARCH, "--settings", "/testdata/settings", "/testdata/references.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds simple ref by cli variable", func() {
		Expect(env.Execute("add", "references", "--file", ARCH, "VERSION=v1.1.1", "/testdata/references.tmpl")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(1))

		CheckReference(env, cd, "testdata")
	})

	It("adds duplicate references", func() {
		Expect(env.Execute("add", "references", "--file", ARCH, "/testdata/dupreferences.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.References)).To(Equal(2))
	})

	Context("reference by options", func() {
		It("adds simple ref", func() {
			Expect(env.Execute("add", "references", "--file", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION)).To(Succeed())
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
			Expect(env.Execute("add", "references", "--file", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION, "--reference", spec)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			labels := metav1.Labels{}
			labels.Set("test", "value")
			CheckReference(env, cd, "testdata", func(r compdesc.Reference) {
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
			Expect(env.Execute("add", "references", "--file", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION, "--reference", spec)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			labels := metav1.Labels{}
			labels.Set("test", "value")
			CheckReference(env, cd, "testdata", func(r compdesc.Reference) {
				Expect(r.GetLabels()).To(Equal(labels))
			})
		})

		It("completely specified by options with extra identity", func() {
			Expect(env.Execute("add", "references", "--file", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION, "--extra", "purpose=test", "--extra", "label=local")).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			labels := metav1.Labels{}
			labels.Set("test", "value")
			CheckReference(env, cd, "testdata", func(r compdesc.Reference) {
				Expect(r.ExtraIdentity).To(Equal(metav1.Identity{"purpose": "test", "label": "local"}))
			})
		})

		It("completely specified by options with labels", func() {
			Expect(env.Execute("add", "references", "--file", ARCH, "--name", "testdata", "--component", REF, "--version", VERSION, "--label", "*purpose=test", "--label", `label@v1={"local": true}`)).To(Succeed())
			data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())
			Expect(len(cd.References)).To(Equal(1))

			labels := metav1.Labels{}
			labels.Set("purpose", "test", metav1.WithSigning())
			labels.Set("label", map[string]interface{}{"local": true}, metav1.WithVersion("v1"))
			CheckReference(env, cd, "testdata", func(r compdesc.Reference) {
				Expect(r.GetLabels()).To(Equal(labels))
			})
		})
	})

	Context("failures", func() {
		It("rejects adding duplicate ref", func() {
			testutils.ExpectError(env.Execute("add", "references", "--file", ARCH, "/testdata/references-dup.yaml")).To(
				MatchError(`duplicate reference identity "name"="testdata","version"="v1.1.1" (/testdata/references-dup.yaml[1][2] and /testdata/references-dup.yaml[1][1])`),
			)
		})
	})
})
