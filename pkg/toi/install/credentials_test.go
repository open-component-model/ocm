// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/toi/install"
)

var _ = Describe("credential mapping", func() {
	memspec := memory.NewRepositorySpec("default")
	memcred := credentials.DirectCredentials{
		"username": "mandelsoft",
		"password": "secret",
	}
	cfgdata := `
configurations:
- credentials:
  - credentials:
      password: secret
      username: mandelsoft
    credentialsName: other
  - credentials:
      token: XXX
    credentialsName: target
  repoName: default
  type: memory.credentials.config.ocm.software
- consumers:
  - credentials:
    - credentialsName: target
      repoName: default
      type: Memory
    identity:
      name: target
      type: KubernetesCLuster
  type: credentials.config.ocm.software
type: generic.config.ocm.software
`
	It("creates config data", FlakeAttempts(50), func() {
		ctx := credentials.New()
		mem, err := ctx.RepositoryForSpec(memory.NewRepositorySpec("memory"))
		Expect(err).To(Succeed())

		_, err = mem.WriteCredentials("configured", memcred)
		Expect(err).To(Succeed())
		request := `
credentials:
  target:
    description: some kube config
    consumerId:
      type: KubernetesCLuster
      name: target
  other:
    description: some other stuff
`
		req, err := install.ParseCredentialRequest([]byte(request))
		Expect(err).To(Succeed())

		input := `
credentials:
  target:
     credentials:
        token: XXX
  other:
     reference:
       credentialsName: configured
       type: Memory
       repoName: memory
`
		spec, err := install.ParseCredentialSpecification([]byte(input), "settings")
		Expect(err).To(Succeed())
		c, err := install.GetCredentials(ctx, spec, req.Credentials, nil)
		Expect(err).To(Succeed())
		output, err := runtime.DefaultYAMLEncoding.Marshal(c)
		Expect(err).To(Succeed())
		Expect(string(output)).To(StringEqualTrimmedWithContext(cfgdata))
	})

	It("reads config data", func() {
		cfgctx := config.New()
		ctx := credentials.WithConfigs(cfgctx).New()

		_, err := cfgctx.ApplyData([]byte(cfgdata), runtime.DefaultYAMLEncoding, "config data")
		Expect(err).To(Succeed())

		mem, err := ctx.RepositoryForSpec(memspec)
		Expect(err).To(Succeed())
		Expect(mem.LookupCredentials("target")).To(Equal(credentials.DirectCredentials{
			"token": "XXX",
		}))
		Expect(mem.LookupCredentials("other")).To(Equal(memcred))
	})
})
