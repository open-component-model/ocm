package template

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils"
)

const KIND_TEMPLATER = "templater"

type (
	TemplaterOptions map[string]interface{}
	TemplaterFactory func(system vfs.FileSystem, options TemplaterOptions) Templater
)

func (t TemplaterOptions) Get(name string) interface{} {
	if t == nil {
		return nil
	}
	return t[name]
}

type Registry interface {
	Register(name string, fac TemplaterFactory, desc string)
	Create(name string, fs vfs.FileSystem, options ...TemplaterOptions) (Templater, error)
	Describe(name string) (string, error)
	KnownTypeNames() []string
}

type templaterInfo struct {
	templater   TemplaterFactory
	description string
}

type registry struct {
	lock       sync.RWMutex
	templaters map[string]templaterInfo
}

func NewRegistry() Registry {
	return &registry{
		templaters: map[string]templaterInfo{},
	}
}

func (r *registry) Register(name string, fac TemplaterFactory, desc string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.templaters[name] = templaterInfo{
		templater:   fac,
		description: desc,
	}
}

func (r *registry) Create(name string, fs vfs.FileSystem, options ...TemplaterOptions) (Templater, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	t, ok := r.templaters[name]
	if !ok {
		return nil, errors.ErrNotSupported(KIND_TEMPLATER, name)
	}
	return t.templater(fs, general.Optional(options...)), nil
}

func (r *registry) Describe(name string) (string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	t, ok := r.templaters[name]
	if !ok {
		return "", errors.ErrNotSupported(KIND_TEMPLATER, name)
	}
	return t.description, nil
}

func (r *registry) KnownTypeNames() []string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return utils.StringMapKeys(r.templaters)
}

func Usage(scheme Registry) string {
	s := `
There are several templaters that can be selected by the <code>--templater</code> option:
`
	for _, t := range scheme.KnownTypeNames() {
		desc, err := scheme.Describe(t)
		if err == nil {
			var title string
			idx := strings.Index(desc, "\n")
			if idx >= 0 {
				title = desc[:idx]
				desc = desc[idx+1:]
			}
			if strings.TrimSpace(desc) == "" {
				s = fmt.Sprintf("%s- <code>%s</code> %s\n\n", s, t, title)
			} else {
				s = fmt.Sprintf("%s- <code>%s</code> %s\n\n%s", s, t, title, utils.IndentLines(desc, "  "))
			}
			if !strings.HasSuffix(s, "\n") {
				s += "\n"
			}
		}
	}
	return s
}

var _registry = NewRegistry()

func Register(name string, fac TemplaterFactory, desc string) {
	_registry.Register(name, fac, desc)
}

func DefaultRegistry() Registry {
	return _registry
}
