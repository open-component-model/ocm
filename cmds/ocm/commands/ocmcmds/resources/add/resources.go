// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ComponentVersionTag = "<componentversion>"
)

type ResourceSpecHandler struct{}

var _ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)

func (ResourceSpecHandler) RequireInputs() bool {
	return true
}

func (ResourceSpecHandler) Decode(data []byte) (common.ResourceSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func (ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r common.Resource, acc compdesc.AccessSpec) error {
	spec := r.Spec().(*ResourceSpec)
	vers := spec.Version
	if spec.Relation == metav1.LocalRelation {
		if vers == "" || vers == ComponentVersionTag {
			vers = v.GetVersion()
		} else if vers != v.GetVersion() {
			return errors.Newf("local resource %q (%s) has non-matching version %q", spec.Name, r.Source(), vers)
		}
	}
	if vers == ComponentVersionTag {
		vers = v.GetVersion()
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
	compdescv2.Resource `json:",inline"`
}

var _ common.ResourceSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("resource %s: %s", r.Type, r.GetRawIdentity())
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *common.ResourceInput) error {
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
	if err := compdescv2.ValidateResource(fldPath, r.Resource, false); err != nil {
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

////////////////////////////////////////////////////////////////////////////////

type ResourceSpecificationsProvider struct {
	*common.ContentResourceSpecificationsProvider
	external bool
}

func NewResourceSpecificationsProvider(ctx clictx.Context, deftype ...string) common.ResourceSpecificationsProvider {
	return &ResourceSpecificationsProvider{
		ContentResourceSpecificationsProvider: common.NewContentResourceSpecificationProvider(ctx, "resource", deftype...),
	}
}

func (p *ResourceSpecificationsProvider) AddFlags(fs *pflag.FlagSet) {
	p.ContentResourceSpecificationsProvider.AddFlags(fs)
	fs.BoolVarP(&p.external, "external", "", false, "flag non-local resource")
}

func (p *ResourceSpecificationsProvider) Description() string {
	d := p.ContentResourceSpecificationsProvider.Description()
	return d + "Non-local resources can be indicated using the option <code>--external</code>."
}

func (a *ResourceSpecificationsProvider) IsSpecified() bool {
	return a.external || a.ContentResourceSpecificationsProvider.IsSpecified()
}
func (a *ResourceSpecificationsProvider) ParseMeta() (flagsets.Config, error) {
	data, err := a.ContentResourceSpecificationsProvider.ParsedMeta()
	if err != nil {
		return nil, err
	}

	if a.external {
		data["relation"] = metav1.ExternalRelation
	}
	return data, nil
}
