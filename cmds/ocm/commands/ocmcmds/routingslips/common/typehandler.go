package common

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	utils2 "ocm.software/ocm/api/utils"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

type Object struct {
	Component    *comphdlr.Object
	Slip         string
	Error        string
	HistoryEntry *routingslip.HistoryEntry
	Payload      routingslip.Entry
}

var _ output.Manifest = (*Object)(nil)

func (o *Object) AsManifest() interface{} {
	return &Manifest{
		Component:   o.Component.ComponentVersion.GetName(),
		Version:     o.Component.ComponentVersion.GetVersion(),
		Error:       o.Error,
		RoutingSlip: o.Slip,
		Entry:       o.HistoryEntry,
	}
}

type Objects []*Object

type Manifest struct {
	Component string `json:"component"`
	Version   string `json:"version"`
	Error     string `json:"error,omitempty"`

	RoutingSlip string                    `json:"routingSlip,omitempty"`
	Entry       *routingslip.HistoryEntry `json:"entry,omitempty"`
}

func Elem(e interface{}) *Object {
	return e.(*Object)
}

type typeFilter struct {
	types []string
}

func (t typeFilter) ApplyToElemHandler(handler *TypeHandler) {
	if len(t.types) > 0 {
		handler.SetFilter(t)
	}
}

func (t typeFilter) Accept(e routingslip.Entry) bool {
	if len(t.types) == 0 {
		return true
	}
	typ := e.GetType()
	for _, a := range t.types {
		if a == typ {
			return true
		}
	}
	return false
}

func WithTypes(types []string) Option {
	return typeFilter{types}
}

////////////////////////////////////////////////////////////////////////////////

type ElementFilter interface {
	Accept(e routingslip.Entry) bool
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	components comphdlr.Objects
	filter     ElementFilter
	verify     bool
}

var _ utils.TypeHandler = (*TypeHandler)(nil)

func NewTypeHandler(octx clictx.OCM, opts *output.Options, repo ocm.Repository, session ocm.Session, compspecs []string, hopts ...Option) (utils.TypeHandler, error) {
	copts := *opts
	copts.StatusCheck = nil
	components, err := comphdlr.Evaluate(octx, session, repo, compspecs, &copts, MapToCompHandlerOptions(hopts...))
	if err != nil {
		return nil, err
	}
	if len(components) > 1 {
		return nil, fmt.Errorf("multiple component versions selected")
	}
	h := &TypeHandler{
		components: components,
	}
	for _, o := range hopts {
		o.ApplyToElemHandler(h)
	}
	return h, nil
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

func (h *TypeHandler) filterElement(e routingslip.Entry) bool {
	if h.filter == nil {
		return true
	}
	return h.filter.Accept(e)
}

func (h *TypeHandler) all(c *comphdlr.Object) ([]output.Object, error) {
	result := []output.Object{}
	if c.ComponentVersion != nil {
		slips, err := routingslip.Get(c.ComponentVersion)
		if err != nil {
			result = append(result, &Object{
				Component: c,
				Error:     err.Error(),
			})
		} else {
			for _, n := range utils2.StringMapKeys(slips) {
				s, err := slips.Get(n)
				if err != nil {
					return nil, errors.ErrInvalid(routingslip.KIND_ROUTING_SLIP, n)
				}
				h.addEntries(&result, c, s)
			}
		}
	}
	return result, nil
}

func (h *TypeHandler) addEntries(result *[]output.Object, c *comphdlr.Object, slip *routingslip.RoutingSlip) {
	err := slip.Verify(c.ComponentVersion.GetContext(), slip.GetName(), h.verify)
	if err != nil {
		o := &Object{
			Component: c,
			Slip:      slip.GetName(),
			Error:     err.Error(),
		}
		*result = append(*result, o)
	}
	for i := range slip.Entries() {
		h.add(result, c, slip.GetName(), slip.Get(i))
	}
}

func (h *TypeHandler) add(result *[]output.Object, c *comphdlr.Object, n string, he *routingslip.HistoryEntry) {
	if h.filterElement(he.Payload) {
		o := &Object{
			Component:    c,
			Slip:         n,
			HistoryEntry: he,
		}
		e, err := he.Payload.Evaluate(c.ComponentVersion.GetContext())
		if err != nil {
			o.Error = err.Error()
		} else {
			o.Payload = e
		}
		*result = append(*result, o)
	}
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

	slip, err := routingslip.GetSlip(c.ComponentVersion, elemspec.String())
	if err != nil {
		result = append(result, &Object{
			Component: c,
			Slip:      elemspec.String(),
			Error:     err.Error(),
		})
	} else {
		h.addEntries(&result, c, slip)
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
