package install_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/directcreds"
	"ocm.software/ocm/api/credentials/extensions/repositories/memory"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	"ocm.software/ocm/api/tech/oci/identity"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("credential mapping", func() {
	consumerid := credentials.NewConsumerIdentity("CT", identity.ID_HOSTNAME, "github.com", identity.ID_PATHPREFIX, "open-component-model")
	ccreds := common.Properties{
		"user": "open-component-model",
		"pass": "mypass",
	}
	memspec := memory.NewRepositorySpec("default")
	memcred := credentials.DirectCredentials{
		"username": "open-component-model",
		"password": "secret",
	}
	cfgdata := `
configurations:
- credentials:
  - credentials:
      password: secret
      username: open-component-model
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
  - credentials:
    - credentialsName: Credentials
      properties:
        pass: mypass
        user: open-component-model
      type: Credentials
    identity:
      hostname: github.com
      pathprefix: open-component-model/testrepo
      type: CT
  type: credentials.config.ocm.software
type: generic.config.ocm.software
`
	It("creates config data", FlakeAttempts(50), func() {
		ctx := credentials.New()
		ctx.SetCredentialsForConsumer(consumerid, directcreds.NewCredentials(ccreds))
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
forwardedConsumers:
- consumerId:
    type: CT
    hostname: github.com
    pathprefix: open-component-model/testrepo
  consumerType: hostpath
`
		spec, err := install.ParseCredentialSpecification([]byte(input), "settings")
		Expect(err).To(Succeed())
		c, _, err := install.GetCredentials(ctx, spec, req.Credentials, nil)
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

		creq := consumerid.Copy()
		creq[identity.ID_PATHPREFIX] = "open-component-model/testrepo/bla"
		props := Must(credentials.CredentialsForConsumer(ctx, creq, hostpath.Matcher))
		Expect(props.Properties()).To(Equal(ccreds))
	})
})
