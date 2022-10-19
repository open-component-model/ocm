// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const COMP2 = "test.de/y"
const COMP3 = "test.de/z"
const PROVIDER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("component archive", func() {
		It("lists single resource in component archive", func() {
			env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
				})
			})

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NAME     VERSION IDENTITY TYPE      RELATION
testdata v1               PlainText local
`))
		})

		It("lists ambigious resource in component archive", func() {
			env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
					env.ExtraIdentity("platform", "a")
				})
				env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
					env.ExtraIdentity("platform", "b")
				})
			})

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NAME     VERSION IDENTITY       TYPE      RELATION
testdata v1      "platform"="a" PlainText local
testdata v1      "platform"="b" PlainText local
`))
		})

		It("lists single resource in component archive with ref", func() {
			env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
				})
				env.Reference("ref", COMP2, VERSION)
			})

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH, "-r")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REFERENCEPATH NAME     VERSION IDENTITY TYPE      RELATION
test.de/x:v1  testdata v1               PlainText local
`))
		})
		It("tree lists single resource in component archive with ref", func() {
			env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
				})
				env.Reference("ref", COMP2, VERSION)
			})

			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH, "-r", "-o", "tree")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENTVERSION    NAME     VERSION IDENTITY TYPE      RELATION
└─ test.de/x:v1                                         
   └─               testdata v1               PlainText local
`))
		})

	})

	Context("ctf", func() {
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
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NAME     VERSION IDENTITY TYPE      RELATION
testdata v1               PlainText local
`))
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
							env.Reference("base", COMP, VERSION)
						})
					})
				})
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-r", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
REFERENCEPATH              NAME     VERSION IDENTITY TYPE      RELATION
test.de/y:v1               moredata v1               PlainText local
test.de/y:v1->test.de/x:v1 testdata v1               PlainText local
`))
			})

			It("lists flat tree in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
COMPONENTVERSION    NAME     VERSION IDENTITY TYPE      RELATION
└─ test.de/y:v1                                         
   └─               moredata v1               PlainText local
`))
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-r", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
COMPONENTVERSION       NAME     VERSION IDENTITY TYPE      RELATION
└─ test.de/y:v1                                            
   ├─                  moredata v1               PlainText local
   └─ test.de/x:v1                                         
      └─               testdata v1               PlainText local
`))
			})
		})

		Context("with closure and intermediate empty version", func() {
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
							env.Reference("base", COMP, VERSION)
						})
					})
					env.Component(COMP3, func() {
						env.Version(VERSION, func() {
							env.Provider(PROVIDER)
							env.Resource("moredata", "", "PlainText", metav1.LocalRelation, func() {
								env.BlobStringData(mime.MIME_TEXT, "moredata")
							})
							env.Reference("base", COMP2, VERSION)
						})
					})
				})
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-r", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
REFERENCEPATH              NAME     VERSION IDENTITY TYPE      RELATION
test.de/y:v1->test.de/x:v1 testdata v1               PlainText local
`))
			})

			It("lists flat in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
no elements found
`))
			})

			It("lists flat tree in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
COMPONENTVERSION    NAME VERSION IDENTITY TYPE RELATION
└─ test.de/y:v1                                
`))
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-r", "-o", "tree", "--repo", ARCH, COMP3+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
COMPONENTVERSION          NAME     VERSION IDENTITY TYPE      RELATION
└─ test.de/z:v1                                               
   ├─                     moredata v1               PlainText local
   └─ test.de/y:v1                                            
      └─ test.de/x:v1                                         
         └─               testdata v1               PlainText local
`))
			})
		})

		Context("with closure and empty leaf version", func() {
			BeforeEach(func() {
				env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
					env.Component(COMP, func() {
						env.Version(VERSION, func() {
							env.Provider(PROVIDER)
						})
					})
					env.Component(COMP2, func() {
						env.Version(VERSION, func() {
							env.Provider(PROVIDER)
							env.Resource("moredata", "", "PlainText", metav1.LocalRelation, func() {
								env.BlobStringData(mime.MIME_TEXT, "moredata")
							})
							env.Reference("base", COMP, VERSION)
						})
					})
				})
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-r", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
REFERENCEPATH NAME     VERSION IDENTITY TYPE      RELATION
test.de/y:v1  moredata v1               PlainText local
`))
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-r", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
COMPONENTVERSION       NAME     VERSION IDENTITY TYPE      RELATION
└─ test.de/y:v1                                            
   ├─                  moredata v1               PlainText local
   └─ test.de/x:v1                                         
`))
			})
		})
	})
})
