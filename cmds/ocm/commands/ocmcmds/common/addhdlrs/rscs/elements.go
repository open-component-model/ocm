// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rscs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv2 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/v2/pkg/errors"
	"github.com/open-component-model/ocm/v2/pkg/runtime"
)

const (
	ComponentVersionTag = common.ComponentVersionTag
)

type ResourceSpecHandler struct{}

var _ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)

func (ResourceSpecHandler) Key() string {
	return "resource"
}

func (ResourceSpecHandler) RequireInputs() bool {
	return true
}

func (ResourceSpecHandler) Decode(data []byte) (addhdlrs.ElementSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func (ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r addhdlrs.Element, acc compdesc.AccessSpec) error {
	spec, ok := r.Spec().(*ResourceSpec)
	if !ok {
		return fmt.Errorf("element spec is not a valid resource spec, failed to assert type %T to ResourceSpec", r.Spec())
	}
	vers := spec.Version
	if spec.Relation == metav1.LocalRelation {
		if vers == "" || vers == ComponentVersionTag {
			vers = v.GetVersion()
		}
	}
	if vers == ComponentVersionTag {
		vers = v.GetVersion()
	}
	if vers == "" {
		return errors.Newf("resource %q (%s): missing version", spec.Name, r.Source())
	}

	meta := &compdesc.ResourceMeta{
		ElementMeta: compdesc.ElementMeta{
			Name:          spec.Name,
			Version:       vers,
			ExtraIdentity: spec.ExtraIdentity,
			Labels:        spec.Labels,
		},
		Type:      spec.Type,
		Relation:  spec.Relation,
		SourceRef: compdescv2.ConvertSourcerefsTo(spec.SourceRef),
	}
	return v.SetResource(meta, acc)
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	compdescv2.ElementMeta `json:",inline"`

	// Type describes the type of the object.
	Type string `json:"type"`

	// Relation describes the relation of the resource to the component.
	// Can be a local or external resource
	Relation metav1.ResourceRelation `json:"relation,omitempty"`

	// SourceRef defines a list of source names.
	// These names reference the sources defines in `component.sources`.
	SourceRef []compdescv2.SourceRef `json:"srcRef"`

	addhdlrs.ResourceInput `json:",inline"`
}

var _ addhdlrs.ElementSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("resource %s: %s", r.Type, r.GetRawIdentity())
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *addhdlrs.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if r.Relation == "" {
		if input.Input != nil {
			r.Relation = metav1.LocalRelation
		}
		if r.Access != nil {
			r.Relation = metav1.ExternalRelation
		}
	}
	if r.Version == "" && r.Relation == metav1.LocalRelation {
		r.Version = ComponentVersionTag
	}
	rsc := compdescv2.Resource{
		ElementMeta: r.ElementMeta,
		Type:        r.Type,
		Relation:    r.Relation,
		SourceRef:   r.SourceRef,
	}
	if err := compdescv2.ValidateResource(fldPath, rsc, false); err != nil {
		allErrs = append(allErrs, err...)
	}

	if input.Access != nil {
		if r.Relation == metav1.LocalRelation {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("relation"), "access requires external relation"))
		}
	}
	if input.Input != nil {
		if r.Relation != metav1.LocalRelation {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("relation"), "input requires local relation"))
		}
	}
	return allErrs.ToAggregate()
}
