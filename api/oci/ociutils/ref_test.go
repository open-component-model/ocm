package ociutils_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci/ociutils"
	"ocm.software/ocm/api/oci/testhelper"
)

var _ = Describe("Ref Test Environment", func() {
	dig := "sha256:" + testhelper.H_OCIARCHMANIFEST1

	type expect struct {
		yaml        string
		versionSpec string
		isVersion   bool
		version     string
		isTag       bool
		tag         string
		isDigested  bool
		digest      string
	}

	DescribeTable("parsing", func(src string, e expect) {
		v := Must(ociutils.ParseVersion(src))
		Expect(v).NotTo(BeNil())
		Expect(v).To(YAMLEqual(e.yaml))
		Expect(v.VersionSpec()).To(Equal(e.versionSpec))
		Expect(v.IsVersion()).To(Equal(e.isVersion))
		Expect(v.Version()).To(Equal(e.version))
		Expect(v.IsTagged()).To(Equal(e.isTag))
		Expect(v.GetTag()).To(Equal(e.tag))
		Expect(v.IsDigested()).To(Equal(e.isDigested))
		Expect(v.GetDigest()).To(Equal(digest.Digest(e.digest)))
	},
		Entry("empty", "", expect{
			yaml:        "{}",
			versionSpec: "latest",
			version:     "latest",
		}),
		Entry("tag", "tag", expect{
			yaml:        "{\"tag\":\"tag\"}",
			versionSpec: "tag",
			isVersion:   true,
			version:     "tag",
			isTag:       true,
			tag:         "tag",
		}),
		Entry("digest", "@"+dig, expect{
			yaml:        "{\"digest\":\"" + dig + "\"}",
			versionSpec: "@" + dig,
			isVersion:   true,
			version:     "@" + dig,
			isDigested:  true,
			digest:      dig,
		}),
		Entry("tag@digest", "tag@"+dig, expect{
			yaml:        "{\"tag\":\"tag\",\"digest\":\"" + dig + "\"}",
			versionSpec: "tag@" + dig,
			isVersion:   true,
			version:     "@" + dig,
			isTag:       true,
			tag:         "tag",
			isDigested:  true,
			digest:      dig,
		}),
	)
})
