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

package hostpath_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
)

func IdentityMatcher(pattern, cur, id core.ConsumerIdentity) bool {
	return hostpath.IdentityMatcher("OCIRegistry")(pattern, cur, id)
}

var _ = Describe("ctf management", func() {

	Context("with path", func() {
		pat := credentials.ConsumerIdentity{
			hostpath.ID_HOSTNAME:   "host",
			hostpath.ID_PATHPREFIX: "a/b",
			hostpath.ID_PORT:       "4711",
			hostpath.ID_SCHEME:     "scheme://",
		}

		It("complete", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
		})

		It("path prefix", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("different prefix", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "b",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("longer prefix", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a/b/c",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing path", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME: "host",
				hostpath.ID_PORT:     "4711",
				hostpath.ID_SCHEME:   "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing port", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())

			Expect(IdentityMatcher(id, nil, pat)).To(BeTrue()) // accept additional port as fallback
			Expect(IdentityMatcher(id, id, pat)).To(BeFalse()) // but not to replace more general match
		})
		It("different port", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_PORT:       "0815",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})

		It("different host", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "other",
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("no host", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(id, nil, pat)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, id, pat)).To(BeTrue())
		})

		It("different scheme", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "otherscheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("no scheme", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "a/b",
				hostpath.ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(id, nil, pat)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, id, pat)).To(BeTrue())
		})
	})

	Context("without path", func() {
		pat := credentials.ConsumerIdentity{
			hostpath.ID_HOSTNAME: "host",
			hostpath.ID_PORT:     "4711",
			hostpath.ID_SCHEME:   "scheme://",
		}

		It("complete", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME: "host",
				hostpath.ID_PORT:     "4711",
				hostpath.ID_SCHEME:   "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
		})

		It("different prefix", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_PATHPREFIX: "b",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing port", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME: "host",
				hostpath.ID_SCHEME:   "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("different port", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME: "host",
				hostpath.ID_PORT:     "0815",
				hostpath.ID_SCHEME:   "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})

		It("different scheme", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME: "host",
				hostpath.ID_PORT:     "4711",
				hostpath.ID_SCHEME:   "otherscheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("no scheme", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME: "host",
				hostpath.ID_PORT:     "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
	})

})
