// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		cctx := env.CLI.CredentialsContext()

		ids := credentials.ConsumerIdentity{
			identity.ID_TYPE:     "test",
			identity.ID_HOSTNAME: "ghcr.io",
		}
		creds := credentials.DirectCredentials{
			"user": "testuser",
			"pass": "testpass",
		}

		cctx.SetCredentialsForConsumer(ids, creds)

		ids = credentials.ConsumerIdentity{
			identity.ID_TYPE:       identity.CONSUMER_TYPE,
			identity.ID_HOSTNAME:   "ghcr.io",
			identity.ID_PATHPREFIX: "a",
		}
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
		Expect(env.CatchOutput(buf).Execute("get", "credentials", identity.ID_TYPE+"=test", identity.ID_HOSTNAME+"=ghcr.io")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ATTRIBUTE VALUE
pass      testpass
user      testuser
`))
	})
	It("fail with partial matcher", func() {
		buf := bytes.NewBuffer(nil)
		err := env.CatchOutput(buf).Execute("get", "credentials", identity.ID_TYPE+"=test", identity.ID_HOSTNAME+"=gcr.io")
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("consumer \"{\"hostname\":\"gcr.io\",\"type\":\"test\"}\" is unknown"))
	})

	It("get oci type with oci matcher", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "credentials", identity.ID_TYPE+"="+identity.CONSUMER_TYPE, identity.ID_HOSTNAME+"=ghcr.io", identity.ID_PATHPREFIX+"=a/b")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ATTRIBUTE VALUE
password  testpass
username  testuser
`))
	})
})
