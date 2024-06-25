package semverutils

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Masterminds/semver/v3"
)

var _ = Describe("filter", func() {
	V10, _ := semver.NewVersion("1.0")
	V19, _ := semver.NewVersion("1.9")
	V20, _ := semver.NewVersion("2.0")
	V21, _ := semver.NewVersion("2.0.1")

	C2x, _ := semver.NewConstraint(">=2.0")
	C19, _ := semver.NewConstraint("1.9")

	It("just sorts", func() {
		Expect(Must(MatchVersionStrings([]string{"2.0", "1.0", "1.9"}))).To(Equal(semver.Collection{V10, V19, V20}))
	})

	It("filters by constraints", func() {
		Expect(Must(MatchVersionStrings([]string{"2.0", "1.0", "2.0.1", "1.9"}, C2x))).To(Equal(semver.Collection{V20, V21}))
	})

	It("filters by multiple constraints", func() {
		Expect(Must(MatchVersionStrings([]string{"2.0", "1.0", "1.9"}, C2x, C19))).To(Equal(semver.Collection{V19, V20}))
	})

	It("filters invalid", func() {
		r, err := MatchVersionStrings([]string{"2.0", "1.0", "1.x", "1.9"})
		MustFailWithMessage(err, "invalid semver versions: 1.x")
		Expect(r).To(Equal(semver.Collection{V10, V19, V20}))
	})
})
