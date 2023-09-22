// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comp

import (
	"fmt"

	. "github.com/open-component-model/ocm/pkg/finalizer"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/refs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/rscs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/srcs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ComponentVersionTag = common.ComponentVersionTag
)

type ResourceSpecHandler struct {
	rschandler *rscs.ResourceSpecHandler
	version    string
	schema     string
}

var (
	_ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)
	_ options.Options            = (*ResourceSpecHandler)(nil)
)

func New(v string, schema string, opts ...ocm.ModificationOption) *ResourceSpecHandler {
	return &ResourceSpecHandler{rschandler: rscs.New(opts...), version: v, schema: schema}
}

func (h *ResourceSpecHandler) AddFlags(fs *pflag.FlagSet) {
	h.rschandler.AddFlags(fs)
}

func (*ResourceSpecHandler) Key() string {
	return "component"
}

func (*ResourceSpecHandler) RequireInputs() bool {
	return false
}

func (h *ResourceSpecHandler) Decode(data []byte) (addhdlrs.ElementSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	if desc.Version == "" {
		desc.Version = h.version
	}
	return &desc, nil
}

func (*ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r addhdlrs.Element, acc compdesc.AccessSpec) error {
	return fmt.Errorf("not supported for components")
}

func (h *ResourceSpecHandler) Add(ctx clictx.Context, ictx inputs.Context, elem addhdlrs.Element, repo ocm.Repository) (err error) {
	var final Finalizer
	defer final.FinalizeWithErrorPropagation(&err)

	r, ok := elem.Spec().(*ResourceSpec)
	if !ok {
		return fmt.Errorf("element spec is not a valid resource spec, failed to assert type %T to ResourceSpec", elem.Spec())
	}
	comp, err := repo.LookupComponent(r.Name)
	if err != nil {
		return errors.ErrNotFound(errors.KIND_COMPONENT, r.Name)
	}
	final.Close(comp)

	cv, err := comp.NewVersion(r.Version, true)
	if err != nil {
		return errors.Wrapf(err, "%s:%s", r.Name, r.Version)
	}
	final.Close(cv)

	cd := cv.GetDescriptor()

	schema := h.schema
	if r.Meta.ConfiguredVersion != "" {
		schema = r.Meta.ConfiguredVersion
	}
	if schema != "" {
		if compdesc.DefaultSchemes[schema] == nil {
			return errors.ErrUnknown(errors.KIND_SCHEMAVERSION, schema)
		}
		cd.Metadata.ConfiguredVersion = schema
	}

	cd.Labels = r.Labels
	cd.Provider = r.Provider
	if !compatattr.Get(ctx) {
		cd.CreationTime = metav1.NewTimestampP()
	}

	err = handle(ctx, ictx, elem.Source(), cv, r.Sources, srcs.ResourceSpecHandler{})
	if err != nil {
		return err
	}
	err = handle(ctx, ictx, elem.Source(), cv, r.Resources, h.rschandler)
	if err != nil {
		return err
	}
	err = handle(ctx, ictx, elem.Source(), cv, r.References, refs.ResourceSpecHandler{})
	if err != nil {
		return err
	}
	return comp.AddVersion(cv)
}

func handle[T addhdlrs.ElementSpec](ctx clictx.Context, ictx inputs.Context, si addhdlrs.SourceInfo, cv ocm.ComponentVersionAccess, specs []T, h common.ResourceSpecHandler) error {
	key := utils.Plural(h.Key(), 0)
	elems, err := addhdlrs.MapSpecsToElems(ctx, ictx, si.Sub(key), specs, h)
	if err != nil {
		return errors.Wrapf(err, key)
	}
	return common.ProcessElements(ictx, cv, elems, h)
}

// //////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	// Meta enabled to specify information for the serialization
	Meta compdesc.Metadata `json:"meta"`

	metav1.ObjectMeta `json:",inline"`
	// Sources defines sources that produced the component
	Sources []*srcs.ResourceSpec `json:"sources"`
	// References references component dependencies that can be resolved in the current context.
	References []*refs.ResourceSpec `json:"componentReferences"`
	// Resources defines all resources that are created by the component and by a third party.
	Resources []*rscs.ResourceSpec `json:"resources"`
}

var _ addhdlrs.ElementSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("component %s:%s", r.Name, r.Version)
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *addhdlrs.ResourceInput) error {
	cd := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: r.ObjectMeta,
		},
	}
	return compdesc.Validate(cd)
}
