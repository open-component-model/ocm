package refhints_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "ocm.software/ocm/api/ocm/refhints"
)

var _ = Describe("Hints Test Environment", func() {
	Context("hint", func() {
		It("single attr", func() {
			CheckHint("test", v1.New("", "test"))
			CheckHint(`"test"`, v1.New("", "test"))
			CheckHint("typ::test", v1.New("typ", "test"))
			CheckHint(`typ::"te\"st"`, v1.New("typ", `te"st`))
			CheckHint(`typ::"te\;st"`, v1.New("typ", `te;st`))
			CheckHint(`typ::"te\,st"`, v1.New("typ", `te,st`))
			CheckHint(`typ::"te\\st"`, v1.New("typ", `te\st`))
			CheckHint(`typ::x="te\\st"`, v1.DefaultReferenceHint{
				v1.HINT_TYPE: "typ",
				"x":          `te\st`,
			})
			CheckHint(`typ::x=test`, v1.DefaultReferenceHint{
				v1.HINT_TYPE: "typ",
				"x":          `test`,
			})
			CheckHint(`typ::xy=test`, v1.DefaultReferenceHint{
				v1.HINT_TYPE: "typ",
				"xy":         `test`,
			})
			CheckHint(`typ::x,=test`, v1.DefaultReferenceHint{
				v1.HINT_TYPE: "typ",
				"x,":         `test`,
			})
		})

		It("multi attr", func() {
			CheckHint(`typ::x="te\\st",xy=test`, v1.DefaultReferenceHint{
				v1.HINT_TYPE: "typ",
				"x":          `te\st`,
				"xy":         "test",
			})
		})
	})

	Context("hints", func() {
		It("regular", func() {
			CheckHint("x::y;test",
				v1.New("x", `y`),
				v1.New("", "test"))
			CheckHint("x::y;typ::test",
				v1.New("x", `y`),
				v1.New("typ", "test"))
			CheckHint(`x::y;typ::"te\"st"`,
				v1.New("x", `y`),
				v1.New("typ", `te"st`))
			CheckHint(`x::y;typ::"te\;st"`,
				v1.New("x", `y`),
				v1.New("typ", `te;st`))
			CheckHint(`x::y;typ::"te\,st"`,
				v1.New("x", `y`),
				v1.New("typ", `te,st`))
			CheckHint(`x::y;typ::"te\\st"`,
				v1.New("x", `y`),
				v1.New("typ", `te\st`))
			CheckHint(`x::y;typ::x="te\\st"`,
				v1.New("x", `y`),
				v1.DefaultReferenceHint{
					v1.HINT_TYPE: "typ",
					"x":          `te\st`,
				})
			CheckHint(`x::y;typ::x=test`,
				v1.New("x", `y`),
				v1.DefaultReferenceHint{
					v1.HINT_TYPE: "typ",
					"x":          `test`,
				})
			CheckHint(`x::y;typ::xy=test`,
				v1.New("x", `y`),
				v1.DefaultReferenceHint{
					v1.HINT_TYPE: "typ",
					"xy":         `test`,
				})
			CheckHint(`x::y;typ::x,=test`,
				v1.New("x", `y`),
				v1.DefaultReferenceHint{
					v1.HINT_TYPE: "typ",
					"x,":         `test`,
				})
		})

		It("special", func() {
			CheckHint(`typ::x;=test`,
				v1.New("typ", "x"),
				v1.New("", "=test"),
			)
		})
	})
})

func CheckHint(s string, h ...v1.ReferenceHint) {
	r := v1.ParseHints(s)
	if strings.HasPrefix(s, "\"") {
		s = s[1 : len(s)-1]
	}
	ExpectWithOffset(1, r).To(Equal(v1.ReferenceHints(h)))
	ExpectWithOffset(1, r.Serialize()).To(Equal(s))
}
