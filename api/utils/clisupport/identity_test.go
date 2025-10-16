package clisupport_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/clisupport"
)

var _ = Describe("IdentityPath Parsing", func() {
	Context("", func() {
		It("handles simple identity", func() {
			value := `name=alice`
			flag := Must(clisupport.ParseIdentityPath(value))
			Expect(flag).To(Equal([]v1.Identity{{"name": "alice"}}))
		})

		It("handles simple path", func() {
			value1 := `name=alice`
			value2 := `husband=bob`
			flag := Must(clisupport.ParseIdentityPath(value1, value2))
			Expect(flag).To(Equal([]v1.Identity{{"name": "alice", "husband": "bob"}}))
		})

		It("handles mulki path", func() {
			value1 := `name=alice`
			value2 := `husband=bob`
			value3 := "name=bob"
			value4 := "wife=alice"
			value5 := "name=other"
			flag := Must(clisupport.ParseIdentityPath(value1, value2, value3, value4, value5))
			Expect(flag).To(Equal([]v1.Identity{{"name": "alice", "husband": "bob"}, {"name": "bob", "wife": "alice"}, {"name": "other"}}))
		})

		It("rejects invalid value", func() {
			value := `a=b`
			ExpectError(clisupport.ParseIdentityPath(value)).To(MatchError("first attribute must be the name attribute"))
		})

		It("rejects invalid assignment", func() {
			value := `a`
			ExpectError(clisupport.ParseIdentityPath(value)).To(MatchError("identity attribute \"a\" is invalid"))
		})
	})
})
