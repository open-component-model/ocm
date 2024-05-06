package transfer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/mime"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const ARCH = "/tmp/ctf"
const VERSION = "v1"
const NS = "mandelsoft/test"
const OUT = "/tmp/res"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCICommonTransport(OUT, accessio.FormatDirectory)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers a named artifact", func() {
		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NS, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "artifact", ARCH+"//"+NS+":"+VERSION, "directory::"+OUT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
copying /tmp/ctf//mandelsoft/test:v1 to directory::` + OUT + `//mandelsoft/test:v1...
copied 1 from 1 artifact(s) and 1 repositories
`))
		Expect(env.ReadFile(OUT + "/" + ctf.ArtifactIndexFileName)).To(Equal([]byte("{\"schemaVersion\":1,\"artifacts\":[{\"repository\":\"mandelsoft/test\",\"tag\":\"v1\",\"digest\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\"}]}")))
	})

	It("transfers a named artifact to changed repository", func() {
		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NS, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "artifact", ARCH+"//"+NS+":"+VERSION, "directory::"+OUT+"//changed")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
copying /tmp/ctf//mandelsoft/test:v1 to directory::` + OUT + `//changed:v1...
copied 1 from 1 artifact(s) and 1 repositories
`))
		Expect(env.ReadFile(OUT + "/" + ctf.ArtifactIndexFileName)).To(Equal([]byte("{\"schemaVersion\":1,\"artifacts\":[{\"repository\":\"changed\",\"tag\":\"v1\",\"digest\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\"}]}")))
	})

	It("transfers a named artifact to sub repository", func() {
		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NS, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "artifact", "-R", ARCH+"//"+NS+":"+VERSION, "directory::"+OUT+"//sub")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
copying /tmp/ctf//mandelsoft/test:v1 to directory::` + OUT + `//sub/mandelsoft/test:v1...
copied 1 from 1 artifact(s) and 1 repositories
`))
		Expect(env.ReadFile(OUT + "/" + ctf.ArtifactIndexFileName)).To(Equal([]byte("{\"schemaVersion\":1,\"artifacts\":[{\"repository\":\"sub/mandelsoft/test\",\"tag\":\"v1\",\"digest\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\"}]}")))
	})

	It("transfers an unnamed artifact set", func() {
		env.ArtifactSet(ARCH, accessio.FormatDirectory, func() {
			env.Manifest(VERSION, func() {
				env.Config(func() {
					env.BlobStringData(mime.MIME_JSON, "{}")
				})
				env.Layer(func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("transfer", "artifact", ARCH, "directory::"+OUT+"//"+NS)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
copying ArtifactSet::/tmp/ctf//:v1 to directory::` + OUT + `//mandelsoft/test:v1...
copied 1 from 1 artifact(s) and 1 repositories
`))
		Expect(env.ReadFile(OUT + "/" + ctf.ArtifactIndexFileName)).To(Equal([]byte("{\"schemaVersion\":1,\"artifacts\":[{\"repository\":\"mandelsoft/test\",\"tag\":\"v1\",\"digest\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\"}]}")))
	})
})
