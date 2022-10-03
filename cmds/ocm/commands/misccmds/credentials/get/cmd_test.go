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

package get_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		cctx := env.CLI.CredentialsContext()

		ids := credentials.ConsumerIdentity{
			credentials.CONSUMER_ATTR_TYPE: "test",
			identity.ID_HOSTNAME:           "ghcr.io",
		}
		creds := credentials.NewCredentials(common.Properties{
			"user": "testuser",
			"pass": "testpass",
		})

		cctx.SetCredentialsForConsumer(ids, creds)

		ids = credentials.ConsumerIdentity{
			credentials.CONSUMER_ATTR_TYPE: identity.CONSUMER_TYPE,
			identity.ID_HOSTNAME:           "ghcr.io",
			identity.ID_PATHPREFIX:         "a",
		}
		creds = credentials.NewCredentials(common.Properties{
			"username": "testuser",
			"password": "testpass",
		})

		cctx.SetCredentialsForConsumer(ids, creds)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get unknown type with partial matcher", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "credentials", credentials.CONSUMER_ATTR_TYPE+"=test", identity.ID_HOSTNAME+"=ghcr.io")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ATTRIBUTE VALUE
pass      testpass
user      testuser
`))
	})
	It("fail with partial matcher", func() {
		buf := bytes.NewBuffer(nil)
		err := env.CatchOutput(buf).Execute("get", "credentials", credentials.CONSUMER_ATTR_TYPE+"=test", identity.ID_HOSTNAME+"=gcr.io")
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("consumer \"{\"hostname\":\"gcr.io\",\"type\":\"test\"}\" is unknown"))
	})

	It("get oci type with oci matcher", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "credentials", credentials.CONSUMER_ATTR_TYPE+"="+identity.CONSUMER_TYPE, identity.ID_HOSTNAME+"=ghcr.io", identity.ID_PATHPREFIX+"=a/b")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ATTRIBUTE VALUE
password  testpass
username  testuser
`))
	})
})
