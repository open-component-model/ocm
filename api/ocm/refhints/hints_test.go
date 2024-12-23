package refhints_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm/refhints"
)

var _ = Describe("Hints Test Environment", func() {
	Context("hint", func() {
		It("single attr", func() {
			CheckHint("test", refhints.New("", "test"))
			CheckHint(`te\st`, refhints.New("", "te\\st"))
			CheckHint(`"test"`, refhints.New("", "test"))
			CheckHint(`"te;st"`, refhints.New("", "te;st"))
			CheckHint(`"te,st"`, refhints.New("", "te,st"))
			CheckHint("typ::te\\st", refhints.New("typ", "te\\st"))
			CheckHint("typ::test", refhints.New("typ", "test"))
			CheckHint(`typ::"te\"st"`, refhints.New("typ", `te"st`))
			CheckHint(`typ::te\st`, refhints.New("typ", `te\st`))
			CheckHint(`typ::x="te\\;st"`, refhints.DefaultReferenceHint{
				refhints.HINT_TYPE: "typ",
				"x":                `te\;st`,
			})
			CheckHint(`typ::x=test`, refhints.DefaultReferenceHint{
				refhints.HINT_TYPE: "typ",
				"x":                `test`,
			})
			CheckHint(`typ::xy=test`, refhints.DefaultReferenceHint{
				refhints.HINT_TYPE: "typ",
				"xy":               `test`,
			})
			CheckHint(`typ::x,=test`, refhints.DefaultReferenceHint{
				refhints.HINT_TYPE: "typ",
				"x,":               `test`,
			})
			CheckHint("ref=test,x=y", refhints.DefaultReferenceHint{
				"ref": "test",
				"x":   "y",
			})
			CheckHint(`"test,x=y"`, refhints.DefaultReferenceHint{
				refhints.HINT_REFERENCE: "test,x=y",
			})
			CheckHint2("test,x=y", "reference=test,x=y", refhints.DefaultReferenceHint{
				refhints.HINT_REFERENCE: "test",
				"x":                     "y",
			})
		})

		It("multi attr", func() {
			CheckHint(`typ::x=te\st,xy=test`, refhints.DefaultReferenceHint{
				refhints.HINT_TYPE: "typ",
				"x":                `te\st`,
				"xy":               "test",
			})
		})
	})

	Context("hints", func() {
		It("regular", func() {
			CheckHint("x::y;test",
				refhints.New("x", `y`),
				refhints.New("", "test"))
			CheckHint("x::y;typ::test",
				refhints.New("x", `y`),
				refhints.New("typ", "test"))
			CheckHint(`x::y;typ::"te\"st"`,
				refhints.New("x", `y`),
				refhints.New("typ", `te"st`))
			CheckHint(`x::y;typ::"te;st"`,
				refhints.New("x", `y`),
				refhints.New("typ", `te;st`))
			CheckHint(`x::y;typ::"te,st"`,
				refhints.New("x", `y`),
				refhints.New("typ", `te,st`))
			CheckHint(`x::y;typ::te\st`,
				refhints.New("x", `y`),
				refhints.New("typ", `te\st`))
			CheckHint(`x::y;typ::x=te\st`,
				refhints.New("x", `y`),
				refhints.DefaultReferenceHint{
					refhints.HINT_TYPE: "typ",
					"x":                `te\st`,
				})
			CheckHint(`x::y;typ::x=test`,
				refhints.New("x", `y`),
				refhints.DefaultReferenceHint{
					refhints.HINT_TYPE: "typ",
					"x":                `test`,
				})
			CheckHint(`x::y;typ::xy=test`,
				refhints.New("x", `y`),
				refhints.DefaultReferenceHint{
					refhints.HINT_TYPE: "typ",
					"xy":               `test`,
				})
			CheckHint(`x::y;typ::x,=test`,
				refhints.New("x", `y`),
				refhints.DefaultReferenceHint{
					refhints.HINT_TYPE: "typ",
					"x,":               `test`,
				})
			CheckHint(`xy;typ::x,=test`,
				refhints.New("", `xy`),
				refhints.DefaultReferenceHint{
					refhints.HINT_TYPE: "typ",
					"x,":               `test`,
				})
			CheckHint(`"x;y";typ::x,=test`,
				refhints.New("", `x;y`),
				refhints.DefaultReferenceHint{
					refhints.HINT_TYPE: "typ",
					"x,":               `test`,
				})
		})

		It("special", func() {
			CheckHint(`typ::x;=test`,
				refhints.New("typ", "x"),
				refhints.DefaultReferenceHint{
					"": "test",
				},
			)
		})
	})

	Context("implicit", func() {
		It("regular", func() {
			CheckHintImplicit("", "test", "implicit=true,reference=test")
			CheckHintImplicit("typ", "test", "typ::implicit=true,reference=test")
		})
	})
})

func CheckHint(s string, h ...refhints.ReferenceHint) {
	r := refhints.ParseHints(s)
	if strings.HasPrefix(s, "\"") && !strings.ContainsAny(s, ",;") {
		s = s[1 : len(s)-1]
	}
	ExpectWithOffset(1, r).To(Equal(refhints.ReferenceHints(h)))
	ExpectWithOffset(1, r.Serialize()).To(Equal(s))
}

func CheckHint2(s string, ser string, h ...refhints.ReferenceHint) {
	r := refhints.ParseHints(s)
	ExpectWithOffset(1, r).To(Equal(refhints.ReferenceHints(h)))
	ExpectWithOffset(1, r.Serialize()).To(Equal(ser))
}

func CheckHintImplicit(typ, ref, ser string) {
	h := refhints.New(typ, ref, true)
	s := h.Serialize(true)
	ExpectWithOffset(1, h.Serialize(true)).To(Equal(ser))
	r := refhints.ParseHints(s)
	ExpectWithOffset(1, r).To(Equal(refhints.ReferenceHints{h}))
}
