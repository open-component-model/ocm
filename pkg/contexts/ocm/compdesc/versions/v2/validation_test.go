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

package compdesc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/testutils"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2/jsonscheme"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	meta "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Test Suite")
}

var _ = Describe("Validation", func() {
	testutils.TestCompName(jsonscheme.ResourcesComponentDescriptorV2SchemaYamlBytes())

	Context("validator", func() {
		var (
			comp *ComponentDescriptor

			ociImage1    *Resource
			ociRegistry1 *ociartefact.AccessSpec
			ociImage2    *Resource
			ociRegistry2 *ociartefact.AccessSpec
		)

		BeforeEach(func() {
			ociRegistry1 = ociartefact.New("docker/image1:1.2.3")

			unstrucOCIRegistry1, err := runtime.ToUnstructuredTypedObject(ociRegistry1)
			Expect(err).ToNot(HaveOccurred())

			ociImage1 = &Resource{
				ElementMeta: ElementMeta{
					Name:    "image1",
					Version: "1.2.3",
				},
				Relation: meta.ExternalRelation,
				Access:   unstrucOCIRegistry1,
			}
			ociRegistry2 = ociartefact.New("docker/image1:1.2.3")
			unstrucOCIRegistry2, err := runtime.ToUnstructuredTypedObject(ociRegistry2)
			Expect(err).ToNot(HaveOccurred())
			ociImage2 = &Resource{
				ElementMeta: ElementMeta{
					Name:    "image2",
					Version: "1.2.3",
				},
				Relation: meta.ExternalRelation,
				Access:   unstrucOCIRegistry2,
			}

			comp = &ComponentDescriptor{
				Metadata: meta.Metadata{
					Version: SchemaVersion,
				},
				ComponentSpec: ComponentSpec{
					ObjectMeta: ObjectMeta{
						Name:    "my-comp",
						Version: "1.2.3",
					},
					Provider:            "external",
					RepositoryContexts:  nil,
					Sources:             nil,
					ComponentReferences: nil,
					Resources:           []Resource{*ociImage1, *ociImage2},
				},
			}
		})

		Context("#Metadata", func() {

			It("should forbid if the component schemaVersion is missing", func() {
				comp := ComponentDescriptor{
					Metadata: meta.Metadata{},
				}

				errList := Validate(nil, &comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("meta.schemaVersion"),
				}))))
			})

			It("should pass if the component schemaVersion is defined", func() {
				errList := Validate(nil, comp)
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("meta.schemaVersion"),
				}))))
			})

		})

		Context("#ObjectMeta", func() {
			It("should forbid if the component's version is missing", func() {
				comp := ComponentDescriptor{}
				errList := Validate(nil, &comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.name"),
				}))))
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.version"),
				}))))
			})

			It("should forbid if the component's name is missing", func() {
				comp := ComponentDescriptor{}
				errList := Validate(nil, &comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.name"),
				}))))
			})

		})

		Context("#Sources", func() {
			It("should forbid if a duplicated component's source is defined", func() {
				comp.Sources = []Source{
					{
						SourceMeta: SourceMeta{
							ElementMeta: ElementMeta{
								Name: "a",
							},
						},
						Access: runtime.NewEmptyUnstructured("custom"),
					},
					{
						SourceMeta: SourceMeta{
							ElementMeta: ElementMeta{
								Name: "a",
							},
						},
						Access: runtime.NewEmptyUnstructured("custom"),
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.sources[1]"),
				}))))
			})
		})

		Context("#ComponentReferences", func() {
			It("should pass if a reference is set", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Name:    "test",
							Version: "1.2.3",
						},
						ComponentName: "test",
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.componentReferences[0].name"),
				}))))
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.componentReferences[0].version"),
				}))))
			})

			It("should forbid if a reference's name is missing", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Version: "1.2.3",
						},
						ComponentName: "test",
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.componentReferences[0].name"),
				}))))
			})

			It("should forbid if a reference's component name is missing", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Name:    "test",
							Version: "1.2.3",
						},
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.componentReferences[0].componentName"),
				}))))
			})

			It("should forbid if a reference's version is missing", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Name: "test",
						},
						ComponentName: "test",
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("component.componentReferences[0].version"),
				}))))
			})

			It("should forbid if a duplicated component reference is defined", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Name: "test",
						},
					},
					{
						ElementMeta: ElementMeta{
							Name: "test",
						},
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.componentReferences[1]"),
				}))))
			})
		})

		Context("#Resources", func() {
			It("should forbid if a local resource's version differs from the version of the parent", func() {
				comp.Resources = []Resource{
					{
						ElementMeta: ElementMeta{
							Name:    "locRes",
							Version: "0.0.1",
						},
						Relation: meta.LocalRelation,
						Access:   runtime.NewEmptyUnstructured(ociartefact.Type),
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("component.resources[0].version"),
				}))))
			})

			It("should forbid if a resource name contains invalid characters", func() {
				comp.Resources = []Resource{
					{
						ElementMeta: ElementMeta{
							Name: "test$",
						},
					},
					{
						ElementMeta: ElementMeta{
							Name: "test🙅",
						},
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("component.resources[0].name"),
				}))))
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("component.resources[1].name"),
				}))))
			})

			It("should forbid if a duplicated local resource is defined", func() {
				comp.Resources = []Resource{
					{
						ElementMeta: ElementMeta{
							Name: "test",
						},
					},
					{
						ElementMeta: ElementMeta{
							Name: "test",
						},
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.resources[1]"),
				}))))
			})

			It("should forbid if a duplicated resource with additional identity labels is defined", func() {
				comp.Resources = []Resource{
					{
						ElementMeta: ElementMeta{
							Name: "test",
							ExtraIdentity: meta.Identity{
								"my-id": "some-id",
							},
						},
					},
					{
						ElementMeta: ElementMeta{
							Name: "test",
							ExtraIdentity: meta.Identity{
								"my-id": "some-id",
							},
						},
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.resources[1]"),
				}))))
			})

			It("should pass if a duplicated resource has the same name but with different additional identity labels", func() {
				comp.Resources = []Resource{
					{
						ElementMeta: ElementMeta{
							Name: "test",
							ExtraIdentity: meta.Identity{
								"my-id": "some-id",
							},
						},
					},
					{
						ElementMeta: ElementMeta{
							Name: "test",
						},
					},
				}
				errList := Validate(nil, comp)
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.resources[1]"),
				}))))
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.resources[0]"),
				}))))
			})
		})

		Context("#labels", func() {

			It("should forbid if labels are defined multiple times in the same context", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Name:    "test",
							Version: "1.2.3",
							Labels: []meta.Label{
								{
									Name:  "l1",
									Value: []byte{},
								},
								{
									Name:  "l1",
									Value: []byte{},
								},
							},
						},
						ComponentName: "test",
					},
				}

				errList := Validate(nil, comp)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.componentReferences[0].labels[1]"),
				}))))
			})

			It("should pass if labels are defined multiple times in the same context with differnet names", func() {
				comp.ComponentReferences = []ComponentReference{
					{
						ElementMeta: ElementMeta{
							Name:    "test",
							Version: "1.2.3",
							Labels: []meta.Label{
								{
									Name:  "l1",
									Value: []byte{},
								},
								{
									Name:  "l2",
									Value: []byte{},
								},
							},
						},
						ComponentName: "test",
					},
				}

				errList := Validate(nil, comp)
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeDuplicate),
					"Field": Equal("component.componentReferences[0].labels[1]"),
				}))))
			})
		})

		Context("#Identity", func() {
			It("should pass valid identity labels", func() {
				identity := meta.Identity{
					"my-l1": "test",
					"my-l2": "test",
				}
				errList := meta.ValidateIdentity(field.NewPath("identity"), identity)
				Expect(errList).To(HaveLen(0))
			})

			It("should forbid if a identity label define the name", func() {
				identity := meta.Identity{
					"name": "test",
				}
				errList := meta.ValidateIdentity(field.NewPath("identity"), identity)
				Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("identity[name]"),
				}))))
			})

			It("should forbid if a identity label defines a key with invalid characters", func() {
				identity := meta.Identity{
					"my-l1!": "test",
				}
				errList := meta.ValidateIdentity(field.NewPath("identity"), identity)
				Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("identity[my-l1!]"),
				}))))
			})
		})
	})
})
