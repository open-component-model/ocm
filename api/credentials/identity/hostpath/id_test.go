package hostpath_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
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
		It("path prefix with / prefix", func() {
			id := credentials.ConsumerIdentity{
				hostpath.ID_HOSTNAME:   "host",
				hostpath.ID_PATHPREFIX: "/a/b",
				hostpath.ID_PORT:       "4711",
				hostpath.ID_SCHEME:     "scheme://",
			}
			Expect(IdentityMatcher(pat, nil, id)).To(BeTrue())
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
