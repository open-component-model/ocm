package cpi

import (
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"ocm.software/ocm/api/ocm/grammar"
)

type ParseHandler func(u *UniformRepositorySpec) error

type registry struct {
	lock     sync.RWMutex
	handlers map[string]ParseHandler
}

func (r *registry) Register(ty string, h ParseHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.handlers[ty] = h
}

func (r *registry) Get(ty string) ParseHandler {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.handlers[ty]
}

func (r *registry) Handle(u UniformRepositorySpec) (UniformRepositorySpec, error) {
	h := r.Get(u.Type)
	if h != nil {
		err := h(&u)
		return u, err
	}
	return u, nil
}

var parseregistry = &registry{handlers: map[string]ParseHandler{}}

func RegisterRefParseHandler(ty string, h ParseHandler) {
	parseregistry.Register(ty, h)
}

func GetRefParseHandler(ty string, h ParseHandler) {
	parseregistry.Get(ty)
}

func HandleRef(u UniformRepositorySpec) (UniformRepositorySpec, error) {
	return parseregistry.Handle(u)
}

////////////////////////////////////////////////////////////////////////////////

// ParseRepo parses a standard ocm repository reference into a internal representation.
func ParseRepo(ref string) (UniformRepositorySpec, error) {
	create := false
	if strings.HasPrefix(ref, "+") {
		create = true
		ref = ref[1:]
	}
	if strings.HasPrefix(ref, ".") || strings.HasPrefix(ref, "/") {
		return HandleRef(UniformRepositorySpec{
			Info:            ref,
			CreateIfMissing: create,
		})
	}
	match := grammar.AnchoredRepositoryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		h := string(match[1])
		t, _ := grammar.SplitTypeSpec(h)
		return HandleRef(UniformRepositorySpec{
			Type:            t,
			TypeHint:        h,
			Scheme:          string(match[2]),
			Host:            string(match[3]),
			SubPath:         string(match[4]),
			CreateIfMissing: create,
		})
	}

	match = grammar.AnchoredSchemedHostPortRepositoryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		h := string(match[1])
		t, _ := grammar.SplitTypeSpec(h)
		return HandleRef(UniformRepositorySpec{
			Type:            t,
			TypeHint:        h,
			Scheme:          string(match[2]),
			Host:            string(match[3]),
			SubPath:         string(match[4]),
			CreateIfMissing: create,
		})
	}

	match = grammar.AnchoredHostWithPortRepositoryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		h := string(match[1])
		t, _ := grammar.SplitTypeSpec(h)
		return HandleRef(UniformRepositorySpec{
			Type:            t,
			TypeHint:        h,
			Scheme:          string(match[2]),
			Host:            string(match[3]),
			SubPath:         string(match[4]),
			CreateIfMissing: create,
		})
	}

	match = grammar.AnchoredGenericRepositoryRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		return UniformRepositorySpec{}, errors.ErrInvalid(KIND_OCM_REFERENCE, ref)
	}
	h := string(match[1])
	t, _ := grammar.SplitTypeSpec(h)
	return HandleRef(UniformRepositorySpec{
		Type:            t,
		TypeHint:        h,
		Info:            string(match[2]),
		CreateIfMissing: create,
	})
}

func ParseRepoToSpec(ctx Context, ref string, create ...bool) (RepositorySpec, error) {
	uni, err := ParseRepo(ref)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, KIND_REPOSITORYSPEC, ref)
	}
	if !uni.CreateIfMissing {
		uni.CreateIfMissing = general.Optional(create...)
	}
	repoSpec, err := ctx.MapUniformRepositorySpec(&uni)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, KIND_REPOSITORYSPEC, ref)
	}
	return repoSpec, nil
}
