package identity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/tech/wget/identity"

	"ocm.software/ocm/api/credentials"
)

var _ = Describe("wget credential management", func() {
	Context("with path", func() {
		pat := credentials.ConsumerIdentity{
			ID_HOSTNAME:   "host",
			ID_PATHPREFIX: "a/b",
			ID_PORT:       "4711",
		}

		It("complete", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PATHPREFIX: "a/b",
				ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
		})

		It("path prefix", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PATHPREFIX: "a",
				ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("different prefix", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PATHPREFIX: "b",
				ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("longer prefix", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PATHPREFIX: "a/b/c",
				ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing path", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME: "host",
				ID_PORT:     "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing port", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PATHPREFIX: "a/b",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())

			Expect(IdentityMatcher(id, nil, pat)).To(BeTrue()) // accept additional port as fallback
			Expect(IdentityMatcher(id, id, pat)).To(BeFalse()) // but not to replace more general match
		})
		It("different port", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PATHPREFIX: "a/b",
				ID_PORT:       "0815",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})

		It("different host", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "other",
				ID_PATHPREFIX: "a/b",
				ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("no host", func() {
			id := credentials.ConsumerIdentity{
				ID_PATHPREFIX: "a/b",
				ID_PORT:       "4711",
			}
			Expect(IdentityMatcher(id, nil, pat)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, id, pat)).To(BeTrue())
		})
	})

	Context("without path", func() {
		pat := credentials.ConsumerIdentity{
			ID_HOSTNAME: "host",
			ID_PORT:     "4711",
		}

		It("complete", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME: "host",
				ID_PORT:     "4711",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, id, id)).To(BeFalse())
		})

		It("different prefix", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME:   "host",
				ID_PORT:       "4711",
				ID_PATHPREFIX: "b",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("missing port", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME: "host",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
		It("different port", func() {
			id := credentials.ConsumerIdentity{
				ID_HOSTNAME: "host",
				ID_PORT:     "0815",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeFalse())
			Expect(IdentityMatcher(pat, pat, id)).To(BeFalse())
		})
	})
})
