package api

import (
	"github.com/mandelsoft/goutils/stringutils"

	"ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

type _Object = runtime.ObjectVersionedTypedObject

type actionType struct {
	_Object
	spectype ActionSpecType
	restype  ActionResultType
}

var _ ActionType = (*actionType)(nil)

func NewActionType[IS ActionSpec, IR ActionResult](kind, version string) ActionType {
	return NewActionTypeByConverter[IS, IS, IR, IR](kind, version, runtime.IdentityConverter[IS]{}, runtime.IdentityConverter[IR]{})
}

func NewActionTypeByConverter[IS ActionSpec, VS runtime.TypedObject, IR ActionResult, VR runtime.TypedObject](kind, version string, specconv runtime.Converter[IS, VS], resconv runtime.Converter[IR, VR]) ActionType {
	name := runtime.TypeName(kind, version)
	st := runtime.NewVersionedTypedObjectTypeByConverter[ActionSpec, IS, VS](name, specconv)
	rt := runtime.NewVersionedTypedObjectTypeByConverter[ActionResult, IR, VR](name, resconv)
	return &actionType{
		_Object:  runtime.NewVersionedTypedObject(kind, version),
		spectype: st,
		restype:  rt,
	}
}

func (a *actionType) SpecificationType() ActionSpecType {
	return a.spectype
}

func (a *actionType) ResultType() ActionResultType {
	return a.restype
}

func Usage(reg ActionTypeRegistry) string {
	p, buf := misc.NewBufferedPrinter()
	for _, n := range reg.GetActionNames() {
		a := reg.GetAction(n)
		p.Printf("- Name: %s\n", n)
		if a.Description() != "" {
			p.Printf("%s\n", stringutils.IndentLines(a.Description(), "    "))
		}
		if a.Usage() != "" {
			p.Printf("\n%s\n", stringutils.IndentLines(a.Usage(), "    "))
		}
		p := p.AddGap("  ")

		if len(a.ConsumerAttributes()) > 0 {
			p.Printf("Possible Consumer Attributes:\n")
			for _, a := range a.ConsumerAttributes() {
				p.Printf("- <code>%s</code>\n", a)
			}
		}
	}
	return buf.String()
}
