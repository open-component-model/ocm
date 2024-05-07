package download_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/mime"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const COMP2 = "test.de/y"
const PROVIDER = "mandelsoft"
const OUT = "/tmp/res"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists single resource in ctf file", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "resources", "-O", OUT, ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: 8 byte(s) written
`))
		Expect(env.FileExists(OUT)).To(BeTrue())
		Expect(env.ReadFile(OUT)).To(Equal([]byte("testdata")))
	})

	Context("with closure", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
				})
				env.Component(COMP2, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("moredata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "moredata")
						})
						env.Resource("otherdata", "", "PlainText", metav1.LocalRelation, func() {
							env.ExtraIdentity("id", "test")
							env.BlobStringData(mime.MIME_TEXT, "otherdata")
						})
						env.Reference("base", COMP, VERSION)
					})
				})
			})
		})

		It("downloads multiple files", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("download", "resources", "-O", OUT, "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
/tmp/res/test.de/y/v1/moredata: 8 byte(s) written
/tmp/res/test.de/y/v1/otherdata-id=test: 9 byte(s) written
`))

			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(Equal([]byte("moredata")))
			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(Equal([]byte("otherdata")))
		})

		It("downloads closure", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("download", "resources", "-r", "-O", OUT, "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
/tmp/res/test.de/y/v1/moredata: 8 byte(s) written
/tmp/res/test.de/y/v1/otherdata-id=test: 9 byte(s) written
/tmp/res/test.de/y/v1/test.de/x/v1/testdata: 8 byte(s) written
`))

			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/moredata"))).To(Equal([]byte("moredata")))
			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/otherdata-id=test"))).To(Equal([]byte("otherdata")))
			Expect(env.FileExists(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/"+COMP+"/"+VERSION+"/testdata"))).To(BeTrue())
			Expect(env.ReadFile(vfs.Join(env.FileSystem(), OUT, COMP2+"/"+VERSION+"/"+COMP+"/"+VERSION+"/testdata"))).To(Equal([]byte("testdata")))
		})
	})
})
