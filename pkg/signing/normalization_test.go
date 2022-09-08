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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
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
				"srcRef": nil,
				"labels": signing.DynamicArrayExcludes{
					ValueChecker: signing.IgnoreLabelsWithoutSignature,
					Continue:     signing.NoExcludes{},
				},
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

	labeldata, err := json.Marshal(map[string]interface{}{
		"a1": "v1",
		"a2": "v2",
	})
	Expect(err).To(Succeed())
	signed, err := json.Marshal("signed")
	Expect(err).To(Succeed())
	labels := metav1.Labels{
		metav1.Label{Name: "b", Value: labeldata},
		metav1.Label{Name: "a", Value: labeldata},
	}

	data, err := json.Marshal(map[string]interface{}{
		"type": "t1",
		"attr": "value",
	})
	Expect(err).To(Succeed())
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
							Name:          "elem2",
							Version:       "1",
							ExtraIdentity: nil,
							Labels: metav1.Labels{
								metav1.Label{
									Name:  "a",
									Value: labeldata,
								},
								metav1.Label{
									Name:    "b",
									Value:   signed,
									Signing: true,
								},
							},
						},
						Type:      "elemtype",
						Relation:  "local",
						SourceRef: nil,
					},
					Access: ociartefact.New("blob"),
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
		logrus.Infof("%s\n", string(data))

		r := entries.ToString("")
		logrus.Infof("******\n%s\n", r)

		Expect("\n" + r).To(Equal(`
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
          type: ` + ociartefact.Type + `
        }
        labels: [
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: b
            signing: true
            value: signed
          }
        ]
        name: elem2
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
          type: ` + ociartefact.Type + `
        }
        labels: [
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: b
            signing: true
            value: signed
          }
        ]
        name: elem2
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
        labels: [
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: b
            signing: true
            value: signed
          }
        ]
        name: elem2
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
        labels: [
          {
            name: a
            value: {
              a1: v1
              a2: v2
            }
          }
          {
            name: b
            signing: true
            value: signed
          }
        ]
        name: elem2
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

	It("Normalizes struct without no-signing resource labels", func() {

		entries, err := signing.PrepareNormalization(cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"resources": signing.ArrayExcludes{
					Continue: signing.MapExcludes{
						"labels": signing.ExcludeEmpty{signing.DynamicArrayExcludes{
							ValueChecker: signing.IgnoreLabelsWithoutSignature,
							Continue:     signing.NoExcludes{},
						}},
					},
				},
			},
		})
		Expect(err).To(Succeed())
		r := entries.ToString("")
		Expect("\n" + r).To(Equal(`
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
        name: elem1
        relation: local
        type: elemtype
        version: 1
      }
      {
        access: {
          imageReference: blob
          type: ociArtefact
        }
        labels: [
          {
            name: b
            signing: true
            value: signed
          }
        ]
        name: elem2
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
		r := entries.ToString("")
		logrus.Infof("%s\n", r)

		Expect("\n" + r).To(Equal(`
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
        labels: [
          {
            name: b
            signing: true
            value: signed
          }
        ]
        name: elem2
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
		logrus.Infof("%s\n", entries.ToString(""))

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
		logrus.Infof("%s\n", entries.ToString(""))

		Expect("\n" + entries.ToString("")).To(Equal(`
{
  modified: {
    name: test
  }
}`))
	})
})
