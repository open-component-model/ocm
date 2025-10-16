package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/tech/oci/identity"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		cctx := env.CLI.CredentialsContext()

		ids := credentials.NewConsumerIdentity("test", identity.ID_HOSTNAME, "ghcr.io")
		creds := credentials.DirectCredentials{
			"user": "testuser",
			"pass": "testpass",
		}

		cctx.SetCredentialsForConsumer(ids, creds)

		ids = credentials.NewConsumerIdentity(identity.CONSUMER_TYPE,
			identity.ID_HOSTNAME, "ghcr.io",
			identity.ID_PATHPREFIX, "a",
		)
		creds = credentials.DirectCredentials{
			"username": "testuser",
			"password": "testpass",
		}

		cctx.SetCredentialsForConsumer(ids, creds)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get unknown type with partial matcher", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "credentials", cpi.ID_TYPE+"=test", identity.ID_HOSTNAME+"=ghcr.io")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ATTRIBUTE VALUE
pass      testpass
user      testuser
`))
	})
	It("fail with partial matcher", func() {
		buf := bytes.NewBuffer(nil)
		err := env.CatchOutput(buf).Execute("get", "credentials", cpi.ID_TYPE+"=test", identity.ID_HOSTNAME+"=gcr.io")
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("consumer \"{\"hostname\":\"gcr.io\",\"type\":\"test\"}\" is unknown"))
	})

	It("get oci type with oci matcher", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "credentials", cpi.ID_TYPE+"="+identity.CONSUMER_TYPE, identity.ID_HOSTNAME+"=ghcr.io", identity.ID_PATHPREFIX+"=a/b")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ATTRIBUTE VALUE
password  testpass
username  testuser
`))
	})
})
