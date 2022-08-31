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

package identity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
)

var _ = Describe("ctf management", func() {

	Context("with path", func() {
		pat := credentials.ConsumerIdentity{
			identity.ID_HOSTNAME:   "host",
			identity.ID_PATHPREFIX: "a/b",
			identity.ID_PORT:       "4711",
		}

		It("complete", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PATHPREFIX: "a/b",
				identity.ID_PORT:       "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, id, id)).To(BeFalse())
		})

		It("path prefix", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PATHPREFIX: "a",
				identity.ID_PORT:       "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("different prefix", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PATHPREFIX: "b",
				identity.ID_PORT:       "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("longer prefix", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PATHPREFIX: "a/b/c",
				identity.ID_PORT:       "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing path", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME: "host",
				identity.ID_PORT:     "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing port", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PATHPREFIX: "a/b",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())

			Expect(identity.IdentityMatcher(id, nil, pat)).To(BeTrue()) // accept additional port as fallback
			Expect(identity.IdentityMatcher(id, id, pat)).To(BeFalse()) // but not to replace more general match
		})
		It("different port", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PATHPREFIX: "a/b",
				identity.ID_PORT:       "0815",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})

		It("different host", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "other",
				identity.ID_PATHPREFIX: "a/b",
				identity.ID_PORT:       "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("no host", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_PATHPREFIX: "a/b",
				identity.ID_PORT:       "4711",
			}
			Expect(identity.IdentityMatcher(id, nil, pat)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, id, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, id, pat)).To(BeTrue())
		})
	})

	Context("without path", func() {
		pat := credentials.ConsumerIdentity{
			identity.ID_HOSTNAME: "host",
			identity.ID_PORT:     "4711",
		}

		It("complete", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME: "host",
				identity.ID_PORT:     "4711",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, id, id)).To(BeFalse())
		})

		It("different prefix", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME:   "host",
				identity.ID_PORT:       "4711",
				identity.ID_PATHPREFIX: "b",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing port", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME: "host",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("different port", func() {
			id := credentials.ConsumerIdentity{
				identity.ID_HOSTNAME: "host",
				identity.ID_PORT:     "0815",
			}
			Expect(identity.IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(identity.IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
	})

})
