// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocireg_test

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/testutils"
)

func mapversion(s *semver.Version) (*semver.Version, error) {
	meta := s.Metadata()
	if meta == "" {
		return s, nil
	}
	v, err := s.SetMetadata("")
	if err != nil {
		return s, err
	}
	v, err = v.SetPrerelease(s.Prerelease() + "-.-" + meta)
	return &v, err
}

var _ = Describe("ref parsing", func() {

	It("omit v", func() {

		s := Must(semver.NewVersion("1.0.0-rc.1+65"))

		v := Must(mapversion(s))
		Expect(v.Original()).To(Equal("1.0.0-rc.1-.-65"))
	})
	It("keep v", func() {

		s := Must(semver.NewVersion("v1.0.0-rc.1+65"))

		s.Metadata()
		v := Must(mapversion(s))
		Expect(v.Original()).To(Equal("v1.0.0-rc.1-.-65"))
	})
	It("no meta", func() {

		s := Must(semver.NewVersion("v1.0.0-rc.1"))

		s.Metadata()
		v := Must(mapversion(s))
		Expect(v.Original()).To(Equal("v1.0.0-rc.1"))
	})

})
