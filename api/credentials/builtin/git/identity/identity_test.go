package identity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/credentials/builtin/git/identity"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	common "ocm.software/ocm/api/utils/misc"
)

var _ = Describe("consumer id handling", func() {
	repo := "https://github.com/torvalds/linux.git"

	Context("id determination", func() {
		It("handles https repos", func() {
			id, err := GetConsumerId(repo)
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"port", "443",
				"hostname", "github.com",
				"scheme", "https",
			)))
		})

		It("handles http repos", func() {
			id, err := GetConsumerId("http://github.com/torvalds/linux.git")
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"port", "80",
				"hostname", "github.com",
				"scheme", "http",
			)))
		})

		It("handles ssh standard format repos", func() {
			id, err := GetConsumerId("ssh://github.com/torvalds/linux.git")
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"port", "22",
				"hostname", "github.com",
				"scheme", "ssh",
			)))
		})

		It("handles ssh git @ format repos", func() {
			id, err := GetConsumerId("git@github.com:torvalds/linux.git")
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"port", "22",
				"hostname", "github.com",
				"scheme", "ssh",
			)))
		})

		It("handles git format repos", func() {
			id, err := GetConsumerId("git://github.com/torvalds/linux.git")
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"port", "9418",
				"hostname", "github.com",
				"scheme", "git",
			)))
		})

		It("handles file format repos", func() {
			id, err := GetConsumerId("file:///path/to/linux/repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"scheme", "file",
				"hostname", "localhost",
				"path", "/path/to/linux/repo",
			)))
		})
	})

	Context("query credentials", func() {
		var ctx oci.Context
		var credctx credentials.Context

		BeforeEach(func() {
			ctx = oci.New(datacontext.MODE_EXTENDED)
			credctx = ctx.CredentialsContext()
		})

		It("Basic Auth", func() {
			user, pass := "linus", "torvalds"
			id, err := GetConsumerId(repo)
			Expect(err).ToNot(HaveOccurred())
			credctx.SetCredentialsForConsumer(id,
				credentials.CredentialsFromList(
					ATTR_USERNAME, user,
					ATTR_PASSWORD, pass,
				),
			)

			creds, err := GetCredentials(ctx, repo)
			Expect(err).ToNot(HaveOccurred())
			Expect(creds).To(BeEquivalentTo(common.Properties{
				ATTR_USERNAME: user,
				ATTR_PASSWORD: pass,
			}))
		})

		It("Token Auth", func() {
			token := "mytoken"
			id, err := GetConsumerId(repo)
			Expect(err).ToNot(HaveOccurred())
			credctx.SetCredentialsForConsumer(id,
				credentials.CredentialsFromList(
					ATTR_TOKEN, token,
				),
			)

			creds, err := GetCredentials(ctx, repo)
			Expect(err).ToNot(HaveOccurred())
			Expect(creds).To(BeEquivalentTo(common.Properties{
				ATTR_TOKEN: token,
			}))
		})

		It("Public Key Auth", func() {
			user, key := "linus", "path/to/my/id_rsa"
			id, err := GetConsumerId(repo)
			Expect(err).ToNot(HaveOccurred())
			credctx.SetCredentialsForConsumer(id,
				credentials.CredentialsFromList(
					ATTR_USERNAME, user,
					ATTR_PRIVATE_KEY, key,
				),
			)

			creds, err := GetCredentials(ctx, repo)
			Expect(err).ToNot(HaveOccurred())
			Expect(creds).To(BeEquivalentTo(common.Properties{
				ATTR_USERNAME:    user,
				ATTR_PRIVATE_KEY: key,
			}))
		})
	})
})
