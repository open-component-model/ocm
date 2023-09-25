// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package elemhdlr

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

type Object struct {
	History   common.History
	Version   ocm.ComponentVersionAccess
	VersionId metav1.Identity

	Spec    metav1.Identity
	Id      metav1.Identity
	Node    *common.NameVersion
	Element compdesc.ElementMetaAccessor
}

func (o *Object) String() string {
	return fmt.Sprintf("history: %s, id: %s, location: %s", o.History, o.Id, common.VersionedElementKey(o.Version))
}

var (
	_ common.HistorySource = (*Object)(nil)
	_ tree.Object          = (*Object)(nil)
)

type Manifest struct {
	History common.History               `json:"context"`
	Element compdesc.ElementMetaAccessor `json:"element"`
}

func (o *Object) AsManifest() interface{} {
	return &Manifest{
		History: o.History,
		Element: o.Element,
	}
}

func (o *Object) GetHistory() common.History {
	return o.History
}

func (o *Object) IsNode() *common.NameVersion {
	return o.Node
}

func (o *Object) IsValid() bool {
	return o.Element != nil
}

func (o *Object) Compare(b *Object) int {
	c := o.History.Compare(b.History)
	if c == 0 {
		if o.IsValid() {
			c = strings.Compare(o.Element.GetMeta().GetName(), b.Element.GetMeta().GetName())
			if c == 0 {
				c = strings.Compare(o.Id.String(), b.Id.String())
			}
		} else {
			c = 0
		}
	}
	return c
}

////////////////////////////////////////////////////////////////////////////////

type ElementFilter interface {
	Accept(e compdesc.ElementMetaAccessor) bool
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	repository ocm.Repository
	components []*comphdlr.Object
	session    ocm.Session
	kind       string
	forceEmpty bool
	filter     ElementFilter
	elemaccess func(ocm.ComponentVersionAccess) compdesc.ElementAccessor
}

func NewTypeHandler(octx clictx.OCM, oopts *output.Options, repobase ocm.Repository, session ocm.Session, kind string, compspecs []string, elemaccess func(ocm.ComponentVersionAccess) compdesc.ElementAccessor, hopts ...Option) (utils.TypeHandler, error) {
	components, err := comphdlr.Evaluate(octx, session, repobase, compspecs, oopts, MapToCompHandlerOptions(hopts...)...)
	if err != nil {
		return nil, err
	}

	t := &TypeHandler{
		components: components,
		repository: repobase,
		session:    session,
		elemaccess: elemaccess,
		kind:       kind,
	}
	for _, o := range hopts {
		o.ApplyToElemHandler(t)
	}
	return t, nil
}

func (h *TypeHandler) SetFilter(f ElementFilter) {
	h.filter = f
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	result := []output.Object{}
	for _, c := range h.components {
		sub, err := h.all(c)
		if err != nil {
			return nil, err
		}
		result = append(result, sub...)
	}
	return result, nil
}

func (h *TypeHandler) filterElement(e compdesc.ElementMetaAccessor) bool {
	if h.filter == nil {
		return true
	}
	return h.filter.Accept(e)
}

func (h *TypeHandler) all(c *comphdlr.Object) ([]output.Object, error) {
	result := []output.Object{}
	if c.ComponentVersion != nil {
		elemaccess := h.elemaccess(c.ComponentVersion)
		l := elemaccess.Len()
		for i := 0; i < l; i++ {
			e := elemaccess.Get(i)
			if h.filterElement(e) {
				result = append(result, &Object{
					History:   c.History.Append(common.VersionedElementKey(c.ComponentVersion)),
					Version:   c.ComponentVersion,
					VersionId: c.Identity,
					Id:        e.GetMeta().GetIdentity(elemaccess),
					Element:   e,
				})
			}
		}

		if len(result) == 0 && h.forceEmpty {
			result = append(result, &Object{
				History:   c.History.Append(common.VersionedElementKey(c.ComponentVersion)),
				Version:   c.ComponentVersion,
				VersionId: c.Identity,
				Id:        metav1.Identity{},
				Element:   nil,
			})
		}
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	var result []output.Object
	for _, c := range h.components {
		sub, err := h.get(c, elemspec)
		if err != nil {
			return nil, err
		}
		result = append(result, sub...)
	}
	return result, nil
}

func (h *TypeHandler) get(c *comphdlr.Object, elemspec utils.ElemSpec) ([]output.Object, error) {
	var result []output.Object

	selector, ok := elemspec.(metav1.Identity)
	if !ok {
		return nil, fmt.Errorf("element spec is not a valid identity, failed to assert type %T to metav1.Identity", elemspec)
	}
	elemaccess := h.elemaccess(c.ComponentVersion)
	l := elemaccess.Len()
	for i := 0; i < l; i++ {
		e := elemaccess.Get(i)
		if !h.filterElement(e) {
			continue
		}
		m := e.GetMeta()
		eid := m.GetMatchBaseIdentity()
		ok, _ := selector.Match(eid)
		if ok {
			result = append(result, &Object{
				History:   c.History.Append(common.VersionedElementKey(c.ComponentVersion)),
				Version:   c.ComponentVersion,
				VersionId: c.Identity,
				Spec:      selector,
				Id:        m.GetIdentity(elemaccess),
				Element:   e,
			})
		}
	}
	if len(result) == 0 && h.forceEmpty {
		result = append(result, &Object{
			History:   c.History.Append(common.VersionedElementKey(c.ComponentVersion)),
			Version:   c.ComponentVersion,
			VersionId: c.Identity,
			Id:        metav1.Identity{},
			Element:   nil,
		})
	}
	return result, nil
}

func MapToCompHandlerOptions(opts ...Option) comphdlr.Options {
	var copts []comphdlr.Option
	for _, o := range opts {
		if c, ok := o.(comphdlr.Option); ok {
			copts = append(copts, c)
		} else {
			if c, ok := o.(Options); ok {
				copts = append(copts, MapToCompHandlerOptions(c...))
			}
		}
	}
	return copts
}
