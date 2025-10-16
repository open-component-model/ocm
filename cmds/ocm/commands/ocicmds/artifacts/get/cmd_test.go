package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/common/handlers/artifacthdlr"
)

const (
	ARCH     = "/tmp/ctf"
	VERSION1 = "v1"
	VERSION2 = "v2"
	NS1      = "mandelsoft/test"
	NS2      = "mandelsoft/index"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	Context("without attached artifacts", func() {
		BeforeEach(func() {
			env = NewTestEnv()
			env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Namespace(NS1, func() {
					env.Manifest(VERSION1, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
					env.Manifest(VERSION2, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "otherdata")
						})
					})
				})

				env.Namespace(NS2, func() {
					env.Index(VERSION1, func() {
						env.Manifest("", func() {
							env.Config(func() {
								env.BlobStringData(mime.MIME_JSON, "{}")
							})
							env.Layer(func() {
								env.BlobStringData(mime.MIME_TEXT, "testdata")
							})
						})
						env.Manifest("", func() {
							env.Config(func() {
								env.BlobStringData(mime.MIME_JSON, "{}")
							})
							env.Layer(func() {
								env.BlobStringData(mime.MIME_TEXT, "otherdata")
							})
						})
					})
					env.Manifest(VERSION2, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "yetanotherdata")
						})
					})
				})
			})
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("get single artifacts", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", ARCH+"//"+NS1+":"+VERSION1)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REGISTRY REPOSITORY      KIND     TAG DIGEST
/tmp/ctf mandelsoft/test manifest v1  sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
`))
		})
		It("get all artifacts in namespace", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", ARCH+"//"+NS1)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REGISTRY REPOSITORY      KIND     TAG DIGEST
/tmp/ctf mandelsoft/test manifest v1  sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
/tmp/ctf mandelsoft/test manifest v2  sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
`))
		})

		It("get all artifacts in other namespace", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", ARCH+"//"+NS2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REGISTRY REPOSITORY       KIND     TAG DIGEST
/tmp/ctf mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
/tmp/ctf mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
		})

		It("get closure of all artifacts in other namespace", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", "-r", ARCH+"//"+NS2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REFERENCEPATH                                                           REGISTRY REPOSITORY       KIND     TAG DIGEST
                                                                        /tmp/ctf mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627 /tmp/ctf mandelsoft/index manifest -   sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627 /tmp/ctf mandelsoft/index manifest -   sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
                                                                        /tmp/ctf mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
		})
		It("get tree of all tagged artifacts in other namespace", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", "-o", "tree", ARCH+"//"+NS2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NESTING REGISTRY REPOSITORY       KIND     TAG DIGEST
├─      /tmp/ctf mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
└─      /tmp/ctf mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
		})

		It("get tree of all artifacts in other namespace", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", "-r", "-o", "tree", ARCH+"//"+NS2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NESTING REGISTRY REPOSITORY       KIND     TAG DIGEST
├─ ⊗    /tmp/ctf mandelsoft/index index    v1  sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
│  ├─   /tmp/ctf mandelsoft/index manifest -   sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
│  └─   /tmp/ctf mandelsoft/index manifest -   sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
└─      /tmp/ctf mandelsoft/index manifest v2  sha256:e51c2165e00ec22eba0b6d18fe7b136491edce1fa4d286549fb35bd5538c03df
`))
		})
	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("with attached artifacts", func() {
		BeforeEach(func() {
			env = NewTestEnv()
			env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Namespace(NS1, func() {
					env.Manifest(VERSION1, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
					})
					env.Manifest(VERSION2, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "otherdata")
						})
					})
				})

				env.Namespace(NS2, func() {
					var nested1 *artdesc.Descriptor
					desc := env.Index(VERSION1, func() {
						nested1 = env.Manifest("", func() {
							env.Config(func() {
								env.BlobStringData(mime.MIME_JSON, "{}")
							})
							env.Layer(func() {
								env.BlobStringData(mime.MIME_TEXT, "testdata")
							})
						})
						env.Manifest("", func() {
							env.Config(func() {
								env.BlobStringData(mime.MIME_JSON, "{}")
							})
							env.Layer(func() {
								env.BlobStringData(mime.MIME_TEXT, "otherdata")
							})
						})
					})
					env.Manifest(artifacthdlr.Attachment(desc.Digest, "test"), func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "attacheddata")
						})
					})
					env.Manifest(artifacthdlr.Attachment(nested1.Digest, "test"), func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "nestedattacheddata")
						})
					})
					env.Manifest(VERSION2, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "yetanotherdata")
						})
					})
				})
			})
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("get single artifact and attachment", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", "-a", ARCH+"//"+NS2+":"+VERSION1)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REGISTRY REPOSITORY       KIND     TAG                                                                          DIGEST
/tmp/ctf mandelsoft/index index    v1                                                                           sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
/tmp/ctf mandelsoft/index manifest sha256-d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627.test sha256:439d433cd85eac706b86e39d3d9dbbd5f1ff19acd1bcb7aa3549f5d7b11777d9
`))
		})

		It("get single artifact attachment tree", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", "-a", "-o", "tree", ARCH+"//"+NS2+":"+VERSION1)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NESTING    REGISTRY REPOSITORY       KIND     TAG                                                                          DIGEST
└─ ⊗       /tmp/ctf mandelsoft/index index    v1                                                                           sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
   └─ test /tmp/ctf mandelsoft/index manifest sha256-d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627.test sha256:439d433cd85eac706b86e39d3d9dbbd5f1ff19acd1bcb7aa3549f5d7b11777d9
`))
		})

		It("get single artifact attachment tree with closure", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "artifact", "-a", "-r", "-o", "tree", ARCH+"//"+NS2+":"+VERSION1)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NESTING       REGISTRY REPOSITORY       KIND     TAG                                                                          DIGEST
└─ ⊗          /tmp/ctf mandelsoft/index index    v1                                                                           sha256:d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627
   ├─ test    /tmp/ctf mandelsoft/index manifest sha256-d6c3ddc587296fd09f56d2f4c8482f05575306a64705b06fae1d5695cb88d627.test sha256:439d433cd85eac706b86e39d3d9dbbd5f1ff19acd1bcb7aa3549f5d7b11777d9
   ├─ ⊗       /tmp/ctf mandelsoft/index manifest -                                                                            sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9
   │  └─ test /tmp/ctf mandelsoft/index manifest sha256-2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9.test sha256:efbfe2c665fc93690911d74e8e7dcf7fb01524545c7b87cb14d5febf1613eaba
   └─         /tmp/ctf mandelsoft/index manifest -                                                                            sha256:60b245b3de64c43b18489e9c3cf177402f9bd18ab62f8cc6653e2fc2e3a5fc39
`))
		})
	})
})
