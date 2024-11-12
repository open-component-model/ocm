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
	DescribeTable("parsing", func(src, yaml, vspec string, isvers bool, vers string, istag bool, tag string, isdig bool, dig string) {
		v := Must(ociutils.ParseVersion(src))
		Expect(v).NotTo(BeNil())
		Expect(v).To(YAMLEqual(yaml))
		Expect(v.VersionSpec()).To(Equal(vspec))
		Expect(v.IsVersion()).To(Equal(isvers))
		Expect(v.Version()).To(Equal(vers))
		Expect(v.IsTagged()).To(Equal(istag))
		Expect(v.GetTag()).To(Equal(tag))
		Expect(v.IsDigested()).To(Equal(isdig))
		Expect(v.GetDigest()).To(Equal(digest.Digest(dig)))
	},
		Entry("empty", "", "{}", "latest", false, "latest", false, "", false, ""),
		Entry("tag", "tag", "{\"tag\":\"tag\"}", "tag", true, "tag", true, "tag", false, ""),
		Entry("digest", "@"+dig, "{\"digest\":\""+dig+"\"}", "@"+dig, true, "@"+dig, false, "", true, dig),
		Entry("tag@digest", "tag@"+dig, "{\"tag\":\"tag\",\"digest\":\""+dig+"\"}", "tag@"+dig, true, "@"+dig, true, "tag", true, dig),
	)
})
