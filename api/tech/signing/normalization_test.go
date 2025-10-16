package signing_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/rules"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/norm/entry"
	"ocm.software/ocm/api/tech/signing/norm/jcs"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
)

var CDExcludes = signing.MapExcludes{
	"component": signing.MapExcludes{
		"repositoryContexts": nil,
		"resources": signing.DefaultedMapFields{
			Next: signing.DynamicArrayExcludes{
				ValueChecker: rules.IgnoreResourcesWithAccessType("localBlob"),
				Continue: signing.MapExcludes{
					"access":  nil,
					"srcRefs": nil,
					"labels": signing.DynamicArrayExcludes{
						ValueChecker: rules.IgnoreLabelsWithoutSignature,
						Continue:     signing.NoExcludes{},
					},
				},
			},
		}.EnforceNull("extraIdentity"),
		"sources": signing.DynamicArrayExcludes{
			ValueChecker: rules.IgnoreResourcesWithNoneAccess,
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
						Type:     "elemtype",
						Relation: "local",
						SourceRefs: []compdesc.SourceRef{
							{
								IdentitySelector: map[string]string{
									"name": "non-existent",
								},
								Labels: nil,
							},
						},
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
						Type:     "elemtype",
						Relation: "local",
						SourceRefs: []compdesc.SourceRef{
							{
								IdentitySelector: map[string]string{
									"name": "non-existent",
								},
								Labels: nil,
							},
						},
					},
					Access: ociartifact.New("blob"),
				},
			},
		},
	}

	cd = compdesc.DefaultComponent(cd)

	It("Normalizes struct without excludes", func() {
		entries, err := signing.PrepareNormalization(entry.New(), cd, signing.NoExcludes{})
		Expect(err).To(Succeed())

		_, err = entries.Marshal("  ")
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
        type: elemtype
        version: 1
      }
      {
        access: {
          imageReference: blob
          type: ` + ociartifact.Type + `
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
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
		entries, err := signing.PrepareNormalization(entry.New(), cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"repositoryContexts": nil,
			},
		})
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
        type: elemtype
        version: 1
      }
      {
        access: {
          imageReference: blob
          type: ` + ociartifact.Type + `
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
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
		entries, err := signing.PrepareNormalization(entry.New(), cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"resources": signing.ArrayExcludes{
					signing.MapExcludes{
						"access": nil,
					},
				},
			},
		})
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
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
		entries, err := signing.PrepareNormalization(entry.New(), cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"resources": signing.DynamicArrayExcludes{
					ValueChecker: rules.IgnoreResourcesWithAccessType("localBlob"),
					Continue: signing.MapExcludes{
						"access": nil,
					},
				},
			},
		})
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
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

	It("list of plain values", func() {
		v := struct {
			List    []string `json:"list"`
			Empty   []string `json:"empty"`
			Omitted []string `json:"omitted,omitempty"`
		}{
			Empty:   []string{},
			Omitted: []string{},
		}
		d, err := json.Marshal(v)
		Expect(err).To(Succeed())
		var m map[string]interface{}

		err = json.Unmarshal(d, &m)
		Expect(err).To(Succeed())

		entries, err := signing.PrepareNormalization(entry.New(), v, signing.ExcludeEmpty{})
		Expect(err).To(Succeed())
		Expect(entries.String()).To(Equal(`[]`))
	})

	It("list of plain values", func() {
		v := map[string]interface{}{
			"list": []interface{}{
				"alice", "bob",
			},
		}
		entries, err := signing.PrepareNormalization(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.String())
		Expect(entries.Formatted()).To(Equal(`[
  {
    "list": [
      "alice",
      "bob"
    ]
  }
]`))
	})

	It("list of complex values", func() {
		v := map[string]interface{}{
			"list": []map[string]interface{}{
				{
					"alice": 25,
				},
				{
					"bob": 26,
				},
			},
		}
		entries, err := signing.PrepareNormalization(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.String())
		Expect(entries.Formatted()).To(Equal(`[
  {
    "list": [
      [
        {
          "alice": 25
        }
      ],
      [
        {
          "bob": 26
        }
      ]
    ]
  }
]`))
	})

	It("simple map", func() {
		v := map[string]interface{}{
			"bob":   26,
			"alice": 25,
		}
		entries, err := signing.PrepareNormalization(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.String())
		Expect(entries.Formatted()).To(Equal(`[
  {
    "alice": 25
  },
  {
    "bob": 26
  }
]`))
	})

	It("map with maps", func() {
		v := map[string]interface{}{
			"people": map[string]interface{}{
				"bob":   26,
				"alice": 25,
			},
		}
		entries, err := signing.PrepareNormalization(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.String())
		Expect(entries.Formatted()).To(Equal(`[
  {
    "people": [
      {
        "alice": 25
      },
      {
        "bob": 26
      }
    ]
  }
]`))
	})

	It("simple lists", func() {
		v := []interface{}{
			"bob",
			"alice",
		}
		entries, err := signing.Prepare(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		data, err := entries.Marshal("")
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", string(data))
		Expect(string(data)).To(Equal(`["bob","alice"]`))
	})

	It("list of maps", func() {
		v := []interface{}{
			map[string]interface{}{
				"bob": 26,
			},
			map[string]interface{}{
				"alice": 25,
			},
		}
		entries, err := signing.Prepare(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		data, err := entries.Marshal("")
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", string(data))
		Expect(string(data)).To(Equal(`[[{"bob":26}],[{"alice":25}]]`))
	})

	It("list of maps", func() {
		in := `
resources:
- access:
    localReference: blob
    mediaType: text/plain
    referenceName: ref
    type: localBlob
  extraIdentity:
    additional: value
    other: othervalue
  name: elem1
  relation: local
  type: elemtype
  version: 1
`
		var v interface{}
		err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(in), &v)
		Expect(err).To(Succeed())
		entries, err := signing.PrepareNormalization(entry.New(), v, signing.NoExcludes{})
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.Formatted())
		Expect(entries.String()).To(Equal(`[{"resources":[[{"access":[{"localReference":"blob"},{"mediaType":"text/plain"},{"referenceName":"ref"},{"type":"localBlob"}]},{"extraIdentity":[{"additional":"value"},{"other":"othervalue"}]},{"name":"elem1"},{"relation":"local"},{"type":"elemtype"},{"version":1}]]}]`))
	})

	It("Normalizes struct without no-signing resource labels", func() {
		entries, err := signing.PrepareNormalization(entry.New(), cd, signing.MapExcludes{
			"component": signing.MapExcludes{
				"resources": signing.ArrayExcludes{
					Continue: signing.MapExcludes{
						"labels": signing.ExcludeEmpty{signing.DynamicArrayExcludes{
							ValueChecker: rules.IgnoreLabelsWithoutSignature,
							Continue:     signing.NoExcludes{},
						}},
					},
				},
			},
		})
		Expect(err).To(Succeed())
		fmt.Printf("%s\n", entries.String())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
        type: elemtype
        version: 1
      }
      {
        access: {
          imageReference: blob
          type: ociArtifact
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
        srcRefs: [
          {
            identitySelector: {
              name: non-existent
            }
          }
        ]
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
		entries, err := signing.PrepareNormalization(entry.New(), cd, CDExcludes)
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
        extraIdentity: null
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
		entries, err := signing.PrepareNormalization(entry.New(), cd, rules)
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
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
		entries, err := signing.PrepareNormalization(entry.New(), cd, rules)
		Expect(err).To(Succeed())
		Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  modified: {
    name: test
  }
}`))
	})

	Context("JCS", func() {
		It("list of plain values", func() {
			v := struct {
				List    []string `json:"list"`
				Empty   []string `json:"empty"`
				Omitted []string `json:"omitted,omitempty"`
			}{
				Empty:   []string{},
				Omitted: []string{},
			}
			d, err := json.Marshal(v)
			Expect(err).To(Succeed())
			var m map[string]interface{}

			err = json.Unmarshal(d, &m)
			Expect(err).To(Succeed())

			entries, err := signing.PrepareNormalization(jcs.New(), v, signing.ExcludeEmpty{})
			Expect(err).To(Succeed())
			Expect(entries.String()).To(Equal(`{}`))

			entries, err = signing.PrepareNormalization(jcs.New(), v, signing.NoExcludes{})
			Expect(err).To(Succeed())
			Expect(entries.String()).To(Equal(`{"empty":[],"list":null}`))
		})

		It("Normalizes cd", func() {
			entries, err := signing.PrepareNormalization(jcs.New(), cd, CDExcludes)
			Expect(err).To(Succeed())
			Expect(entries.String()).To(StringEqualTrimmedWithContext(`
 {"component":{"componentReferences":[],"labels":[{"name":"b","value":{"a1":"v1","a2":"v2"}},{"name":"a","value":{"a1":"v1","a2":"v2"}}],"name":"test","provider":{"name":"provider"},"resources":[{"extraIdentity":null,"labels":[{"name":"b","signing":true,"value":"signed"}],"name":"elem2","relation":"local","type":"elemtype","version":"1"}],"sources":[],"version":"1"},"meta":{"configuredSchemaVersion":"v2"}}
`))
		})
	})

	AssertBasics := func(desc string, n signing.Normalization) {
		When(desc, func() {
			data := map[string]interface{}{
				"map1": map[string]interface{}{
					"field1": "value1",
					"field2": "value2",
				},
			}

			It("injects fields", func() {
				rules := signing.DefaultedMapFields{
					Name: "map1",
					Fields: map[string]interface{}{
						"injectedfield": "injected",
					},
				}
				entries, err := signing.PrepareNormalization(entry.Type, data, rules)
				Expect(err).To(Succeed())
				Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  map1: {
    field1: value1
    field2: value2
    injectedfield: injected
  }
}
`))
			})
			It("injects fields and excludes other", func() {
				rules := signing.DefaultedMapFields{
					Name: "map1",
					Fields: map[string]interface{}{
						"injectedfield": "injected",
					},
					Continue: signing.MapExcludes{
						"field1": nil,
					},
				}
				entries, err := signing.PrepareNormalization(entry.Type, data, rules)
				Expect(err).To(Succeed())
				Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  map1: {
    field2: value2
    injectedfield: injected
  }
}
`))
			})

			It("injects empty map fields and excludes other", func() {
				rules := signing.DefaultedMapFields{
					Name: "map1",
					Fields: map[string]interface{}{
						"injectedfield": map[string]interface{}{},
					},
					Continue: signing.MapExcludes{
						"field1": nil,
					},
				}
				entries, err := signing.PrepareNormalization(n, data, rules)
				Expect(err).To(Succeed())
				Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  map1: {
    field2: value2
    injectedfield: {}
  }
}
`))
			})

			It("injects empty list fields and excludes other", func() {
				rules := signing.DefaultedMapFields{
					Name: "map1",
					Fields: map[string]interface{}{
						"injectedfield": []interface{}{},
					},
					Continue: signing.MapExcludes{
						"field1": nil,
					},
				}
				entries, err := signing.PrepareNormalization(n, data, rules)
				Expect(err).To(Succeed())
				Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  map1: {
    field2: value2
    injectedfield: []
  }
}
`))
			})

			It("injects nil field", func() {
				rules := signing.DefaultedMapFields{
					Name: "map1",
					Fields: map[string]interface{}{
						"injectedfield": signing.Null,
					},
				}
				entries, err := signing.PrepareNormalization(n, data, rules)
				Expect(err).To(Succeed())
				Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  map1: {
    field1: value1
    field2: value2
    injectedfield: null
  }
}
`))
			})

			It("injects nil field", func() {
				rules := signing.DefaultedMapFields{
					Name: "map1",
					Fields: map[string]interface{}{
						"injectedfield": nil,
					},
				}
				entries, err := signing.PrepareNormalization(n, data, rules)
				Expect(err).To(Succeed())
				Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  map1: {
    field1: value1
    field2: value2
    injectedfield: null
  }
}
`))
			})

			Context("conditional", func() {
				data := map[string]interface{}{
					"array": []interface{}{
						map[string]interface{}{
							"field1": "value11",
							"field2": "value12",
							"cond":   "value",
						},
						map[string]interface{}{
							"field1": "value21",
							"field2": "value22",
						},
					},
					"one": []interface{}{
						map[string]interface{}{
							"field2": "valueOne",
						},
					},
				}

				Context("array", func() {
					It("selects exclude rules", func() {
						rules := signing.MapExcludes{
							"array": signing.ConditionalArrayExcludes{
								ValueChecker: func(v interface{}) bool {
									return v.(map[string]interface{})["cond"] == "value"
								},
								ContinueTrue: signing.MapExcludes{
									"field1": nil,
								},
								ContinueFalse: signing.MapExcludes{
									"field2": nil,
								},
							},
						}
						entries, err := signing.PrepareNormalization(n, data, rules)
						Expect(err).To(Succeed())
						Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {
      cond: value
      field2: value12
    }
    {
      field1: value21
    }
  ]
  one: [
    {
      field2: valueOne
    }
  ]
}
`))
					})
				})

				Context("map", func() {
					It("selects exclude rules", func() {
						rules := signing.ConditionalMapExcludes{
							"array": &signing.ConditionalExclude{
								ValueChecker: func(v interface{}) bool {
									return len(v.([]interface{})) > 1
								},
								ContinueTrue: signing.ArrayExcludes{signing.MapExcludes{"field2": nil}},
							},
						}
						entries, err := signing.PrepareNormalization(n, data, rules)
						Expect(err).To(Succeed())
						Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {
      cond: value
      field1: value11
    }
    {
      field1: value21
    }
  ]
  one: [
    {
      field2: valueOne
    }
  ]
}
`))
					})
				})
			})

			Context("array", func() {
				data := map[string]interface{}{
					"array": []interface{}{
						map[string]interface{}{
							"field1": "value1",
							"field2": "value2",
						},
					},
				}

				It("injects field in map array", func() {
					rules := signing.MapExcludes{
						"array": signing.DefaultedMapFields{
							Fields: map[string]interface{}{
								"injectedfield": "injected",
							},
							Continue: signing.MapExcludes{
								"field1": nil,
							},
						},
					}
					entries, err := signing.PrepareNormalization(n, data, rules)
					Expect(err).To(Succeed())
					Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {
      field2: value2
      injectedfield: injected
    }
  ]
}
`))
				})

				It("injects empty map field in map array", func() {
					rules := signing.MapExcludes{
						"array": signing.DefaultedMapFields{
							Continue: signing.MapExcludes{
								"field1": nil,
							},
						}.EnforceEmptyMap("injectedfield"),
					}
					entries, err := signing.PrepareNormalization(n, data, rules)
					Expect(err).To(Succeed())
					Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {
      field2: value2
      injectedfield: {}
    }
  ]
}
`))
				})

				It("injects empty array field in map array", func() {
					rules := signing.MapExcludes{
						"array": signing.DefaultedMapFields{
							Continue: signing.MapExcludes{
								"field1": nil,
							},
						}.EnforceEmptyList("injectedfield"),
					}
					entries, err := signing.PrepareNormalization(n, data, rules)
					Expect(err).To(Succeed())
					Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {
      field2: value2
      injectedfield: []
    }
  ]
}
`))
				})

				It("injects null field in map array", func() {
					rules := signing.MapExcludes{
						"array": signing.DefaultedMapFields{
							Fields: map[string]interface{}{
								"injectedfield": signing.Null,
							},
							Continue: signing.MapExcludes{
								"field1": nil,
							},
						},
					}
					entries, err := signing.PrepareNormalization(n, data, rules)
					Expect(err).To(Succeed())
					Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {
      field2: value2
      injectedfield: null
    }
  ]
}
`))
				})

				It("injects empty map in array", func() {
					rules := signing.MapExcludes{
						"array": signing.DefaultedListEntries{
							Default: map[string]interface{}{},
							Continue: signing.MapExcludes{
								"field1": nil,
							},
						},
					}
					data := map[string]interface{}{
						"array": []interface{}{
							nil,
							map[string]interface{}{
								"field1": "value1",
								"field2": "value2",
							},
						},
					}
					entries, err := signing.PrepareNormalization(n, data, rules)
					Expect(err).To(Succeed())
					Expect(entries.ToString("")).To(StringEqualTrimmedWithContext(`
{
  array: [
    {}
    {
      field2: value2
    }
  ]
}
`))
				})
			})
		})
	}

	AssertBasics("entry normalization", entry.Type)
	AssertBasics("JCS normalization", jcs.Type)
})
