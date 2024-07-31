//go:build redis_test

package redis_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub/providers/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub/types/redis"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub/types/redis/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
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
