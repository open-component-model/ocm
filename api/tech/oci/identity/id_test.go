package identity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/oci/identity"
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
