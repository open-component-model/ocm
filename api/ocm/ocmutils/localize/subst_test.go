package localize_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/helper/builder"
	envhelper "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/ocm/ocmutils/localize"
)

var _ = Describe("value substitution in filesystem", func() {
	var (
		env       *builder.Builder
		payloadfs vfs.FileSystem
	)

	BeforeEach(func() {
		env = builder.NewBuilder(envhelper.ModifiableTestData())
		fs, err := projectionfs.New(env.FileSystem(), "testdata")
		Expect(err).To(Succeed())
		payloadfs = fs
	})

	AfterEach(func() {
		vfs.Cleanup(payloadfs)
		vfs.Cleanup(env)
	})

	It("handles simple values substitution", func() {
		subs := UnmarshalSubstitutions(`
- name: test1
  file: dir/manifest1.yaml
  path: manifest.value1
  value: config1
- name: test2
  file: dir/manifest2.yaml
  path: manifest.value2
  value: config2
`)
		err := localize.Substitute(subs, payloadfs)
		Expect(err).To(Succeed())

		CheckYAMLFile("dir/manifest1.yaml", payloadfs, `
manifest:
  value1: config1
  value2: orig2
`)
		CheckYAMLFile("dir/manifest2.yaml", payloadfs, `
manifest:
  value1: orig1
  value2: config2
`)
	})

	It("handles multiple values substitution", func() {
		subs := UnmarshalSubstitutions(`
- name: test1
  file: dir/manifest1.yaml
  path: manifest.value1
  value: config1
- name: test2
  file: dir/manifest1.yaml
  path: manifest.value2
  value: config2
`)
		err := localize.Substitute(subs, payloadfs)
		Expect(err).To(Succeed())

		CheckYAMLFile("dir/manifest1.yaml", payloadfs, `
manifest:
  value1: config1
  value2: config2
`)
	})

	It("handles json substitution", func() {
		subs := UnmarshalSubstitutions(`
- name: test1
  file: dir/some.json
  path: manifest.value1
  value:
    some:
      value: 1
`)
		err := localize.Substitute(subs, payloadfs)
		Expect(err).To(Succeed())

		CheckJSONFile("dir/some.json", payloadfs, `
{"manifest": {"value1": {"some": {"value": 1}}, "value2": "orig2"}}

`)
	})
})
