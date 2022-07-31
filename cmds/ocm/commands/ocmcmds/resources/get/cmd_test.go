// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package get_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
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
			Expect("\n" + buf.String()).To(Equal(
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
			Expect("\n" + buf.String()).To(Equal(
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
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH, "-c")).To(Succeed())
			fmt.Printf("%s", buf)
			Expect("\n" + buf.String()).To(Equal(
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
			Expect(env.CatchOutput(buf).Execute("get", "resources", ARCH, "-c", "-o", "tree")).To(Succeed())
			fmt.Printf("%s", buf)
			Expect("\n" + buf.String()).To(Equal(
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
			Expect("\n" + buf.String()).To(Equal(
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
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-c", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect("\n" + buf.String()).To(Equal(
					`
REFERENCEPATH              NAME     VERSION IDENTITY TYPE      RELATION
test.de/y:v1               moredata v1               PlainText local
test.de/y:v1->test.de/x:v1 testdata v1               PlainText local
`))
			})

			It("lists flat tree in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect("\n" + buf.String()).To(Equal(
					`
COMPONENTVERSION    NAME     VERSION IDENTITY TYPE      RELATION
└─ test.de/y:v1                                         
   └─               moredata v1               PlainText local
`))
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-c", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect("\n" + buf.String()).To(Equal(
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
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-c", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				fmt.Printf("%s", buf)
				Expect("\n" + buf.String()).To(Equal(
					`
REFERENCEPATH              NAME     VERSION IDENTITY TYPE      RELATION
test.de/y:v1->test.de/x:v1 testdata v1               PlainText local
`))
			})

			It("lists flat in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect("\n" + buf.String()).To(Equal(
					`
no elements found
`))
			})

			It("lists flat tree in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				Expect("\n" + buf.String()).To(Equal(
					`
COMPONENTVERSION    NAME VERSION IDENTITY TYPE RELATION
└─ test.de/y:v1                                
`))
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-c", "-o", "tree", "--repo", ARCH, COMP3+":"+VERSION)).To(Succeed())
				fmt.Printf("%s", buf)
				Expect("\n" + buf.String()).To(Equal(
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
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-c", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				fmt.Printf("%s", buf)
				Expect("\n" + buf.String()).To(Equal(
					`
REFERENCEPATH NAME     VERSION IDENTITY TYPE      RELATION
test.de/y:v1  moredata v1               PlainText local
`))
			})

			It("lists resource closure in ctf file", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "resources", "-c", "-o", "tree", "--repo", ARCH, COMP2+":"+VERSION)).To(Succeed())
				fmt.Printf("%s", buf)
				Expect("\n" + buf.String()).To(Equal(
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
