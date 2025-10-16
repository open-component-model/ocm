package genericocireg_test

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
)

const META_SEPARATOR = genericocireg.META_SEPARATOR

func mapversion(s *semver.Version) (*semver.Version, error) {
	meta := s.Metadata()
	if meta == "" {
		return s, nil
	}
	v, err := s.SetMetadata("")
	if err != nil {
		return s, err
	}
	v, err = v.SetPrerelease(s.Prerelease() + META_SEPARATOR + meta)
	return &v, err
}

var _ = Describe("ref parsing", func() {
	It("omit v", func() {
		s := Must(semver.NewVersion("1.0.0-rc.1+65"))

		v := Must(mapversion(s))
		Expect(v.Original()).To(Equal("1.0.0-rc.1" + META_SEPARATOR + "65"))
	})
	It("keep v", func() {
		s := Must(semver.NewVersion("v1.0.0-rc.1+65"))

		s.Metadata()
		v := Must(mapversion(s))
		Expect(v.Original()).To(Equal("v1.0.0-rc.1" + META_SEPARATOR + "65"))
	})
	It("no meta", func() {
		s := Must(semver.NewVersion("v1.0.0-rc.1"))

		s.Metadata()
		v := Must(mapversion(s))
		Expect(v.Original()).To(Equal("v1.0.0-rc.1"))
	})
})
