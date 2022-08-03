package cpi

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
)

type ConfigType string

const (
	ContainerRegistry ConfigType = "container_registry"
)

type Credential interface {
	Name() string
	ConsumerIdentity() cpi.ConsumerIdentity
	Data() cpi.Credentials
}

type Handler interface {
	ConfigType() ConfigType
	ParseConfig([]byte) ([]Credential, error)
}

var handlers = map[ConfigType]Handler{}
var lock sync.RWMutex

func RegisterHandler(h Handler) {
	lock.Lock()
	defer lock.Unlock()
	handlers[h.ConfigType()] = h
}

func GetHandler(configType ConfigType) Handler {
	lock.RLock()
	defer lock.RUnlock()
	return handlers[configType]
}

func GetHandlers() map[ConfigType]Handler {
	lock.RLock()
	defer lock.RUnlock()

	m := map[ConfigType]Handler{}
	for k, v := range handlers {
		m[k] = v
	}
	return m
}
