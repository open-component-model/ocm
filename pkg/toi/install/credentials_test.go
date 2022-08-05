// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package install_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/toi/install"

	"github.com/open-component-model/ocm/pkg/contexts/config"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
	"github.com/open-component-model/ocm/pkg/runtime"
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
  type: memory.credentials.config.ocm.gardener.cloud
- consumers:
  - credentials:
    - credentialsName: target
      repoName: default
      type: Memory
    identity:
      name: target
      type: KubernetesCLuster
  type: credentials.config.ocm.gardener.cloud
type: generic.config.ocm.gardener.cloud
`
	It("creates config data", func() {
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
		c, err := install.GetCredentials(ctx, spec, req)
		Expect(err).To(Succeed())
		output, err := runtime.DefaultYAMLEncoding.Marshal(c)
		Expect(err).To(Succeed())
		fmt.Printf("%s", output)
		Expect("\n" + string(output)).To(Equal(cfgdata))
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
