package show_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH      = "/tmp/ctf"
	NAMESPACE = "mandelsoft/test"
	V13       = "v1.3"
	V131      = "v1.3.1"
	V132      = "v1.3.2"
	V132x     = "v1.3.2-beta.1"
	V14       = "v1.4"
	V2        = "v2.0"
	OTHERVERS = "sometag"
)

var _ = Describe("Show OCI Tags", func() {
	var env *TestEnv
	BeforeEach(func() {
		env = NewTestEnv()

		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NAMESPACE, func() {
				env.Manifest(V13, func() {
					env.Tags(V131, OTHERVERS)
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data131")
					})
				})
				env.Manifest(V132, func() {
					env.Tags(V132x)
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data132")
					})
				})
				env.Manifest(V14, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data14")
					})
				})
				env.Manifest(V2, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data2")
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists tags", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--repo", ARCH, NAMESPACE)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
sometag
v1.3
v1.3.1
v1.3.2
v1.3.2-beta.1
v1.4
v2.0
`))
	})

	It("lists tags for same artifact", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--repo", ARCH, NAMESPACE+":"+V13)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
sometag
v1.3
v1.3.1
`))
	})

	It("lists semver tags", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--semver", "--repo", ARCH, NAMESPACE)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3
v1.3.1
v1.3.2-beta.1
v1.3.2
v1.4
v2.0
`))
	})

	It("lists semver tags for same artifact", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--semver", "--repo", ARCH, NAMESPACE+":"+V13)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3
v1.3.1
`))
	})
})
