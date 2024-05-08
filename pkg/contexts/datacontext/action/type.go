package action

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/api"
)

const KIND_ACTION = api.KIND_ACTION

type (
	Selector           = api.Selector
	Action             = api.Action
	ActionSpec         = api.ActionSpec
	ActionResult       = api.ActionResult
	ActionType         = api.ActionType
	ActionTypeRegistry = api.ActionTypeRegistry
)

func DefaultRegistry() ActionTypeRegistry {
	return api.DefaultRegistry()
}
