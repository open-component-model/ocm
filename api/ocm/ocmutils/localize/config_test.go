package localize_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/ocmutils/localize"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

var config = []byte(`
values:
  a: va
  b: vb
  c:
    a: vca
`)

var _ = Describe("config value mapping", func() {
	It("handles simple values substitution", func() {
		configs := UnmarshalConfigurations(`
- name: test1
  file: file1
  path: a.b.c
  value: value
`)
		subst, err := localize.Configure(configs, nil, nil, nil, nil, config, nil, nil)
		Expect(err).To(Succeed())
		Expect(subst).To(Equal(UnmarshalSubstitutions(`
- name: test1
  file: file1
  path: a.b.c
  value: value
`)))
	})

	It("handles simple expression substitution", func() {
		configs := UnmarshalConfigurations(`
- name: test1
  file: file1
  path: a.b.c
  value: (( values.a ))
`)
		subst, err := localize.Configure(configs, nil, nil, nil, nil, config, nil, nil)
		Expect(err).To(Succeed())
		Expect(subst).To(Equal(UnmarshalSubstitutions(`
- name: test1
  file: file1
  path: a.b.c
  value: va
`)))
	})

	It("fails for invalid expression substitution", func() {
		configs := UnmarshalConfigurations(`
- file: file1
  path: a.b.c
  value: (( values.x ))
`)
		_, err := localize.Configure(configs, nil, nil, nil, nil, config, nil, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("*'values.x' not found"))
	})

	It("handles expression substitution with substitution context", func() {
		context := `
- name: context
  file: file1
  path: a.b.c
  value: contextvalue
`
		configs := UnmarshalConfigurations(`
- name: test1
  file: file1
  path: a.b.c
  value: (( adjustments.context.value ))
`)
		subst, err := localize.Configure(configs, UnmarshalSubstitutions(context), nil, nil, nil, config, nil, nil)
		Expect(err).To(Succeed())
		Expect(subst).To(Equal(UnmarshalSubstitutions(context + `
- name: test1
  file: file1
  path: a.b.c
  value: contextvalue
`)))
	})
	It("handles expression substitution with template data", func() {
		template := []byte(`
helper:
  help: (( |x|->"helped " x ))
`)
		configs := UnmarshalConfigurations(`
- name: test1
  file: file1
  path: a.b.c
  value: (( helper.help(values.a) ))
`)
		subst, err := localize.Configure(configs, nil, nil, nil, template, config, nil, nil)
		Expect(err).To(Succeed())
		Expect(subst).To(Equal(UnmarshalSubstitutions(`
- name: test1
  file: file1
  path: a.b.c
  value: helped va
`)))
	})

	const (
		ARCHIVE   = "archive.ctf"
		COMPONENT = "github.com/comp"
		VERSION   = "1.0.0"
		LIB       = "lib"
	)

	Context("with libs", func() {
		var (
			repo ocm.Repository
			cv   ocm.ComponentVersionAccess
			env  *builder.Builder
		)

		BeforeEach(func() {
			env = builder.NewBuilder(nil)
			env.OCMCommonTransport(ARCHIVE, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider("mandelsoft")
						env.Resource(LIB, "", "Spiff", v1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_YAML, `
utilities:
  <<<: (( &inject &temporary(merge || ~) ))

  lib:
    help: (( |x|->"lib " x ))
`)
						})
					})
				})
			})

			var err error
			repo, err = ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, ARCHIVE, 0, env)
			Expect(err).To(Succeed())

			cv, err = repo.LookupComponentVersion(COMPONENT, VERSION)
			Expect(err).To(Succeed())
		})

		AfterEach(func() {
			Expect(cv.Close()).To(Succeed())
			Expect(repo.Close()).To(Succeed())
			vfs.Cleanup(env)
		})

		It("uses resolved library from component version", func() {
			configs := UnmarshalConfigurations(`
- name: test1
  file: file1
  path: a.b.c
  value: (( utilities.lib.help(values.a) ))
`)

			libs := []v1.ResourceReference{
				{
					Resource: v1.NewIdentity(LIB),
				},
			}
			subst, err := localize.Configure(configs, nil, cv, nil, nil, config, libs, nil)
			Expect(err).To(Succeed())
			Expect(subst).To(Equal(UnmarshalSubstitutions(`
- name: test1
  file: file1
  path: a.b.c
  value: lib va
`)))
		})

		It("uses templated configRules", func() {
			configs := UnmarshalConfigurations(`
- name: test1
  file: file1
  path: a.b.c
  value: (( values.a ))
`)

			template := `
list: [ "a", "b" ]
helper:
  <<<: (( &template ))
  name: (( "gen" k ))
  file: file1
  path: (( "some.path." k ))
  value: (( values.a ))

configRules: 
  - <<<: (( map[.list|k|->*.helper] )) 
`

			libs := []v1.ResourceReference{
				{
					Resource: v1.NewIdentity(LIB),
				},
			}
			subst, err := localize.Configure(configs, nil, cv, nil, []byte(template), config, libs, nil)
			Expect(err).To(Succeed())
			Expect(subst).To(Equal(UnmarshalSubstitutions(`
- name: gena
  file: file1
  path: some.path.a
  value: va
- name: genb
  file: file1
  path: some.path.b
  value: va
- name: test1
  file: file1
  path: a.b.c
  value: va
`)))
		})
	})
})
