// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compdesc

import (
	"testing"

	"github.com/gardener/ocm/pkg/ocm/accessmethods/ociregistry"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Test Suite")
}

var _ = Describe("Validation", func() {

	var (
		comp *ComponentDescriptor

		ociImage1    *Resource
		ociRegistry1 *ociregistry.AccessSpec
		ociImage2    *Resource
		ociRegistry2 *ociregistry.AccessSpec
	)

	BeforeEach(func() {
		ociRegistry1 = ociregistry.New("docker/image1:1.2.3")

		unstrucOCIRegistry1, err := runtime.ToUnstructuredTypedObject(ociRegistry1)
		Expect(err).ToNot(HaveOccurred())

		ociImage1 = &Resource{
			ElementMeta: ElementMeta{
				Name:    "image1",
				Version: "1.2.3",
			},
			Relation: metav1.ExternalRelation,
			Access:   unstrucOCIRegistry1,
		}
		ociRegistry2 = ociregistry.New("docker/image1:1.2.3")
		unstrucOCIRegistry2, err := runtime.ToUnstructuredTypedObject(ociRegistry2)
		Expect(err).ToNot(HaveOccurred())
		ociImage2 = &Resource{
			ElementMeta: ElementMeta{
				Name:    "image2",
				Version: "1.2.3",
			},
			Relation: metav1.ExternalRelation,
			Access:   unstrucOCIRegistry2,
		}

		comp = &ComponentDescriptor{
			Metadata: metav1.Metadata{
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
				Metadata: metav1.Metadata{},
			}

			errList := validate(nil, &comp)
			Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("meta.schemaVersion"),
			}))))
		})

		It("should pass if the component schemaVersion is defined", func() {
			errList := validate(nil, comp)
			Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("meta.schemaVersion"),
			}))))
		})

	})

	Context("#ObjectMeta", func() {
		It("should forbid if the component's version is missing", func() {
			comp := ComponentDescriptor{}
			errList := validate(nil, &comp)
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
			errList := validate(nil, &comp)
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
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
					Relation: metav1.LocalRelation,
					Access:   runtime.NewEmptyUnstructured(ociregistry.Type),
				},
			}
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
			errList := validate(nil, comp)
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
						ExtraIdentity: metav1.Identity{
							"my-id": "some-id",
						},
					},
				},
				{
					ElementMeta: ElementMeta{
						Name: "test",
						ExtraIdentity: metav1.Identity{
							"my-id": "some-id",
						},
					},
				},
			}
			errList := validate(nil, comp)
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
						ExtraIdentity: metav1.Identity{
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
			errList := validate(nil, comp)
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
						Labels: []metav1.Label{
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

			errList := validate(nil, comp)
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
						Labels: []metav1.Label{
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

			errList := validate(nil, comp)
			Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeDuplicate),
				"Field": Equal("component.componentReferences[0].labels[1]"),
			}))))
		})
	})

	Context("#Identity", func() {
		It("should pass valid identity labels", func() {
			identity := metav1.Identity{
				"my-l1": "test",
				"my-l2": "test",
			}
			errList := metav1.ValidateIdentity(field.NewPath("identity"), identity)
			Expect(errList).To(HaveLen(0))
		})

		It("should forbid if a identity label define the name", func() {
			identity := metav1.Identity{
				"name": "test",
			}
			errList := metav1.ValidateIdentity(field.NewPath("identity"), identity)
			Expect(errList).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeForbidden),
				"Field": Equal("identity[name]"),
			}))))
		})

		It("should forbid if a identity label defines a key with invalid characters", func() {
			identity := metav1.Identity{
				"my-l1!": "test",
			}
			errList := metav1.ValidateIdentity(field.NewPath("identity"), identity)
			Expect(errList).ToNot(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeForbidden),
				"Field": Equal("identity[my-l1!]"),
			}))))
		})
	})
})
