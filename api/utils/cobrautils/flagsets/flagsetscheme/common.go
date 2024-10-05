package flagsetscheme

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

type TypeInfo interface {
	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler
	Description() string
	Format() string
}

type typeInfoImpl struct {
	handler flagsets.ConfigOptionTypeSetHandler
}

func (i *typeInfoImpl) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return i.handler
}
