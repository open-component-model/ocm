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

package signing_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociregistry"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
)

var CDExcludes = signing.MapExcludes{
	"component": signing.MapExcludes{
		"repositoryContexts": nil,
		"resources": signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreResourcesWithAccessType("localBlob"),
			Continue: signing.MapExcludes{
				"access": nil,
				"labels": nil,
				"srcRef": nil,
			},
		},
		"sources": signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreResourcesWithNoneAccess,
			Continue: signing.MapExcludes{
				"access": nil,
				"labels": nil,
			},
		},
		"references": signing.ArrayExcludes{
			signing.MapExcludes{
				"labels": nil,
			},
		},
		"signatures": nil,
	},
}

var _ = Describe("normalization", func() {

	data, err := json.Marshal(map[string]interface{}{
		"a1": "v1",
		"a2": "v2",
	})
	Expect(err).To(Succeed())
	labels := metav1.Labels{
		metav1.Label{"b", data},
		metav1.Label{"a", data},
	}

	data, err = json.Marshal(map[string]interface{}{
		"type": "t1",
		"attr": "value",
	})
	unstr := &runtime.UnstructuredTypedObject{}
	err = json.Unmarshal(data, unstr)
	Expect(err).To(Succeed())

	cd := &compdesc.ComponentDescriptor{
		Metadata: compdesc.Metadata{
			ConfiguredVersion: "v2",
		},
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: compdesc.ObjectMeta{
				Name:     "test",
				Version:  "1",
				Labels:   labels,
				Provider: compdesc.Provider{Name: "provider"},
			},
			RepositoryContexts: []*runtime.UnstructuredTypedObject{
				unstr,
			},
			Sources:    nil,
			References: nil,
			Resources: compdesc.Resources{
				compdesc.Resource{
					ResourceMeta: compdesc.ResourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:    "elem1",
							Version: "1",
							ExtraIdentity: metav1.Identity{
								"additional": "value",
								"other":      "othervalue",
							},
							Labels: labels,
						},
						Type:      "elemtype",
						Relation:  "local",
						SourceRef: nil,
					},
					Access: localblob.New("blob", "ref", mime.MIME_TEXT, nil),
				},
				compdesc.Resource{
					ResourceMeta: compdesc.ResourceMeta{
						ElementMeta: compdesc.ElementMeta{
							Name:          "elem1",
							Version:       "1",
							ExtraIdentity: nil,
							Labels:        nil,
						},
						Type:      "elemtype",
						Relation:  "local",
						SourceRef: nil,
					},
					Access: ociregistry.New("blob"),
				},
			},
		},
	}

	cd = compdesc.DefaultComponent(cd)

	It("Normalizes struct without excludes", func() {

		entries, err := signing.PrepareNormalization(cd, signing.NoExcludes{})
		Expect(err).To(Succeed())

		data, err := signing.Marshal("  ", entries)
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", string(data))

		fmt.Printf("******\n%s\n", entries.ToString(""))

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  component: {
    componentReferences: []
    labels: [
      {
        name: b
        value: {
          a1: v1
          a2: v2
        }
      }
      {
        name: a
        value: {
          a1: v1
          a2: v2
        }
      }
    ]
    name: test
    provider: {
      name: provider
    }
    repositoryContexts: [
      {
        attr: value
        type: t1
      }
    ]
    resources: [
      {
        access: {
          localReference: blob
          mediaType: text/plain
          referenceName: ref
          type: localBlob
        }
        extraIdentity: {
          additional: value
          other: othervalue
        }
        labels: [
          {
            name: b
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
        ]
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
      {
        access: {
          imageReference: blob
          type: ociRegistry
        }
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
    ]
    sources: []
    version: 1
  }
  meta: {
    configuredSchemaVersion: v2
  }
}`))
	})

	It("Normalizes struct without repositoryContexts", func() {

		entries, err := signing.PrepareNormalization(cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"repositoryContexts": nil,
			},
		})
		Expect(err).To(Succeed())

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  component: {
    componentReferences: []
    labels: [
      {
        name: b
        value: {
          a1: v1
          a2: v2
        }
      }
      {
        name: a
        value: {
          a1: v1
          a2: v2
        }
      }
    ]
    name: test
    provider: {
      name: provider
    }
    resources: [
      {
        access: {
          localReference: blob
          mediaType: text/plain
          referenceName: ref
          type: localBlob
        }
        extraIdentity: {
          additional: value
          other: othervalue
        }
        labels: [
          {
            name: b
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
        ]
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
      {
        access: {
          imageReference: blob
          type: ociRegistry
        }
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
    ]
    sources: []
    version: 1
  }
  meta: {
    configuredSchemaVersion: v2
  }
}`))
	})

	It("Normalizes struct without access", func() {

		entries, err := signing.PrepareNormalization(cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"resources": signing.ArrayExcludes{
					signing.MapExcludes{
						"access": nil,
					},
				},
			},
		})
		Expect(err).To(Succeed())

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  component: {
    componentReferences: []
    labels: [
      {
        name: b
        value: {
          a1: v1
          a2: v2
        }
      }
      {
        name: a
        value: {
          a1: v1
          a2: v2
        }
      }
    ]
    name: test
    provider: {
      name: provider
    }
    repositoryContexts: [
      {
        attr: value
        type: t1
      }
    ]
    resources: [
      {
        extraIdentity: {
          additional: value
          other: othervalue
        }
        labels: [
          {
            name: b
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
        ]
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
      {
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
    ]
    sources: []
    version: 1
  }
  meta: {
    configuredSchemaVersion: v2
  }
}`))
	})

	It("Normalizes struct without resources of type localBlob", func() {

		entries, err := signing.PrepareNormalization(cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"resources": signing.DynamicArrayExcludes{
					ValueChecker: signing.IgnoreResourcesWithAccessType("localBlob"),
					Continue: signing.MapExcludes{
						"access": nil,
					},
				},
			},
		})
		Expect(err).To(Succeed())
		Expect("\n" + entries.ToString("")).To(Equal(`
{
  component: {
    componentReferences: []
    labels: [
      {
        name: b
        value: {
          a1: v1
          a2: v2
        }
      }
      {
        name: a
        value: {
          a1: v1
          a2: v2
        }
      }
    ]
    name: test
    provider: {
      name: provider
    }
    repositoryContexts: [
      {
        attr: value
        type: t1
      }
    ]
    resources: [
      {
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
    ]
    sources: []
    version: 1
  }
  meta: {
    configuredSchemaVersion: v2
  }
}`))
	})

	It("Normalizes cd", func() {

		entries, err := signing.PrepareNormalization(cd, CDExcludes)
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.ToString(""))

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  component: {
    componentReferences: []
    labels: [
      {
        name: b
        value: {
          a1: v1
          a2: v2
        }
      }
      {
        name: a
        value: {
          a1: v1
          a2: v2
        }
      }
    ]
    name: test
    provider: {
      name: provider
    }
    resources: [
      {
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
    ]
    sources: []
    version: 1
  }
  meta: {
    configuredSchemaVersion: v2
  }
}`))
	})

	It("Normalizes with recursive includes", func() {

		rules := signing.MapIncludes{
			"component": signing.MapIncludes{
				"name": nil,
			},
		}
		entries, err := signing.PrepareNormalization(cd, rules)
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.ToString(""))

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  component: {
    name: test
  }
}`))
	})

	It("Normalizes with recursive modifying includes", func() {

		rules := signing.DynamicMapIncludes{
			"component": &signing.DynamicInclude{
				Continue: signing.DynamicMapIncludes{
					"name": nil,
				},
				Name: "modified",
			},
		}
		entries, err := signing.PrepareNormalization(cd, rules)
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.ToString(""))

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  modified: {
    name: test
  }
}`))
	})
})
