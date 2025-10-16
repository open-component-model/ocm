//go:build redis_test

package redis_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/extensions/pubsub/providers/ocireg"
	"ocm.software/ocm/api/ocm/extensions/pubsub/types/redis"
	"ocm.software/ocm/api/ocm/extensions/pubsub/types/redis/identity"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	ARCH = "ctf"
	COMP = "acme.org/component"
	VERS = "v1"
)

var _ = Describe("Test Environment", func() {
	var env *Builder
	var repo ocm.Repository

	BeforeEach(func() {
		env = NewBuilder()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory)
		attr := pubsub.For(env)
		attr.ProviderRegistry.Register(ctf.Type, &ocireg.Provider{})

		env.CredentialsContext().SetCredentialsForConsumer(
			identity.GetConsumerId("localhost:6379", "ocm", 0),
			credentials.NewCredentials(common.Properties{identity.ATTR_PASSWORD: "redis-test-0815"}),
		)

		repo = Must(ctf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
	})

	AfterEach(func() {
		if repo != nil {
			MustBeSuccessful(repo.Close())
		}
		env.Cleanup()
	})

	Context("local redis server", func() {
		It("tests local server", func() {
			MustBeSuccessful(pubsub.SetForRepo(repo, Must(redis.New("localhost:6379", "ocm", 0))))

			cv := composition.NewComponentVersion(env, COMP, VERS)
			defer Close(cv)

			Expect(repo.GetSpecification().GetKind()).To(Equal(ctf.Type))
			MustBeSuccessful(repo.AddComponentVersion(cv))
		})
	})
})
