package flagsets

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/utils"
)

const (
	KIND_OPTIONTYPE = "option type"
	KIND_OPTION     = "option"
)

type ConfigOptionTypeCreator func(name string, description string) ConfigOptionType

type ValueTypeInfo struct {
	ConfigOptionTypeCreator
	Description string
}

func (i ValueTypeInfo) GetDescription() string {
	return i.Description
}

type ConfigOptionTypeRegistry = *registry

type registry struct {
	lock        sync.RWMutex
	valueTypes  map[string]ValueTypeInfo
	optionTypes map[string]ConfigOptionType
}

func NewConfigOptionTypeRegistry() ConfigOptionTypeRegistry {
	return &registry{
		valueTypes:  map[string]ValueTypeInfo{},
		optionTypes: map[string]ConfigOptionType{},
	}
}

func (r *registry) RegisterOptionType(t ConfigOptionType) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.optionTypes[t.GetName()] = t
}

func (r *registry) RegisterValueType(name string, c ConfigOptionTypeCreator, desc string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.valueTypes[name] = ValueTypeInfo{ConfigOptionTypeCreator: c, Description: desc}
}

func (r *registry) GetValueType(name string) *ValueTypeInfo {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if t, ok := r.valueTypes[name]; ok {
		return &t
	}
	return nil
}

func (r *registry) GetOptionType(name string) ConfigOptionType {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.optionTypes[name]
}

func (r *registry) CreateOptionType(typ, name, desc string) (ConfigOptionType, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	t, ok := r.valueTypes[typ]
	if !ok {
		return nil, errors.ErrUnknown(KIND_OPTIONTYPE, typ)
	}

	n := t.ConfigOptionTypeCreator(name, desc)
	o := r.optionTypes[name]
	if o != nil {
		if o.ValueType() != n.ValueType() {
			return nil, errors.ErrAlreadyExists(KIND_OPTION, name)
		}
		return o, nil
	}
	return n, nil
}

func (r *registry) Usage() string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	tinfo := utils.FormatMap("", r.valueTypes)
	oinfo := utils.FormatMap("", r.optionTypes)

	return `
The following predefined option types can be used:

` + oinfo + `

The following predefined value types are supported:

` + tinfo
}
