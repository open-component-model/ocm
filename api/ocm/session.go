package ocm

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/internal"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

type ComponentContainer interface {
	LookupComponent(name string) (ComponentAccess, error)
}
type ComponentVersionContainer interface {
	LookupVersion(version string) (ComponentVersionAccess, error)
}

type EvaluationResult struct {
	Ref        RefSpec
	Repository Repository
	Component  ComponentAccess
	Version    ComponentVersionAccess
}

type Session interface {
	datacontext.Session

	Finalize(Finalizer)
	LookupRepository(Context, RepositorySpec) (Repository, error)
	LookupRepositoryForConfig(octx Context, data []byte, unmarshaler ...runtime.Unmarshaler) (Repository, error)
	LookupComponent(ComponentContainer, string) (ComponentAccess, error)
	LookupComponentVersion(r ComponentVersionResolver, comp, vers string) (ComponentVersionAccess, error)
	GetComponentVersion(ComponentVersionContainer, string) (ComponentVersionAccess, error)
	EvaluateRef(ctx Context, ref string) (*EvaluationResult, error)
	EvaluateComponentRef(ctx Context, ref string) (*EvaluationResult, error)
	EvaluateVersionRef(ctx Context, ref string) (*EvaluationResult, error)
	DetermineRepository(ctx Context, ref string) (Repository, UniformRepositorySpec, error)
	DetermineRepositoryBySpec(ctx Context, spec *UniformRepositorySpec) (Repository, error)
}

type session struct {
	datacontext.Session
	base         datacontext.SessionBase
	repositories *internal.RepositoryCache
	components   map[datacontext.ObjectKey]ComponentAccess
	versions     map[datacontext.ObjectKey]ComponentVersionAccess
}

var _ Session = (*session)(nil)

var key = reflect.TypeOf(session{})

func NewSession(s datacontext.Session) Session {
	return datacontext.GetOrCreateSubSession(s, key, newSession).(Session)
}

func newSession(s datacontext.SessionBase) datacontext.Session {
	return &session{
		Session:      s.Session(),
		base:         s,
		repositories: internal.NewRepositoryCache(),
		components:   map[datacontext.ObjectKey]ComponentAccess{},
		versions:     map[datacontext.ObjectKey]ComponentVersionAccess{},
	}
}

type Finalizer interface {
	Finalize() error
}

type finalizer struct {
	finalizer Finalizer
}

func (f *finalizer) Close() error {
	return f.finalizer.Finalize()
}

func (s *session) Finalize(f Finalizer) {
	s.Session.AddCloser(&finalizer{f})
}

func (s *session) Close() error {
	return s.Session.Close()
	// TODO: cleanup cache
}

func (s *session) LookupRepositoryForConfig(octx Context, data []byte, unmarshaler ...runtime.Unmarshaler) (Repository, error) {
	spec, err := octx.RepositorySpecForConfig(data, general.Optional(unmarshaler...))
	if err != nil {
		return nil, err
	}
	return s.LookupRepository(octx, spec)
}

func (s *session) LookupRepository(ctx Context, spec RepositorySpec) (Repository, error) {
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}

	repo, cached, err := s.repositories.LookupRepository(ctx, spec)
	if err != nil {
		return nil, err
	}

	// The repo's closer function should only be added once with add closer. Otherwise, it would be attempted to close
	// an already closed object. Thus, we only want to add the repo's closer function, if it was not already cached
	// (and thus, consequently already added to the sessions close).
	// Session has to take over responsibility for open repositories for the Repository Cache because the objects
	// opened during a session have to be closed in the reverse order they were opened (e.g. components opened based
	// on a previously opened repository have to be closed first).
	if !cached {
		s.base.AddCloser(repo)
	}

	return repo, nil
}

func (s *session) LookupComponent(c ComponentContainer, name string) (ComponentAccess, error) {
	key := datacontext.ObjectKey{
		Object: c,
		Name:   name,
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if ns := s.components[key]; ns != nil {
		return ns, nil
	}
	ns, err := c.LookupComponent(name)
	if err != nil {
		return nil, err
	}
	s.components[key] = ns
	s.base.AddCloser(ns)
	return ns, err
}

func (s *session) LookupComponentVersion(r ComponentVersionResolver, comp, vers string) (ComponentVersionAccess, error) {
	if repo, ok := r.(Repository); ok {
		component, err := s.LookupComponent(repo, comp)
		if err != nil {
			return nil, err
		}
		return s.GetComponentVersion(component, vers)
	}

	key := datacontext.ObjectKey{
		Object: r,
		Name:   common.NewNameVersion(comp, vers).String(),
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if obj := s.versions[key]; obj != nil {
		return obj, nil
	}

	obj, err := r.LookupComponentVersion(comp, vers)
	if err != nil {
		return nil, err
	}

	s.versions[key] = obj
	s.base.AddCloser(obj)
	return obj, err
}

func (s *session) GetComponentVersion(c ComponentVersionContainer, version string) (ComponentVersionAccess, error) {
	if c == nil {
		return nil, fmt.Errorf("no container given")
	}
	key := datacontext.ObjectKey{
		Object: c,
		Name:   version,
	}
	s.base.Lock()
	defer s.base.Unlock()
	if s.base.IsClosed() {
		return nil, errors.ErrClosed("session")
	}
	if obj := s.versions[key]; obj != nil {
		return obj, nil
	}
	obj, err := c.LookupVersion(version)
	if err != nil {
		return nil, err
	}
	s.versions[key] = obj
	s.base.AddCloser(obj)
	return obj, err
}

func (s *session) EvaluateVersionRef(ctx Context, ref string) (*EvaluationResult, error) {
	evaluated, err := s.EvaluateComponentRef(ctx, ref)
	if err != nil {
		return nil, err
	}
	versions, err := evaluated.Component.ListVersions()
	if err != nil {
		return evaluated, errors.Wrapf(err, "%s[%s]: listing versions", ref, evaluated.Ref.Component)
	}
	if len(versions) != 1 {
		return evaluated, errors.Wrapf(err, "%s {%s]: found %d components", ref, evaluated.Ref.Component, len(versions))
	}
	evaluated.Version, err = s.GetComponentVersion(evaluated.Component, versions[0])
	if err != nil {
		return evaluated, errors.Wrapf(err, "%s {%s:%s]: listing components", ref, evaluated.Ref.Component, versions[0])
	}
	evaluated.Ref.Version = &versions[0]
	return evaluated, nil
}

func (s *session) EvaluateComponentRef(ctx Context, ref string) (*EvaluationResult, error) {
	evaluated, err := s.EvaluateRef(ctx, ref)
	if err != nil {
		return evaluated, err
	}
	if evaluated.Component == nil {
		lister := evaluated.Repository.ComponentLister()
		if lister == nil {
			return evaluated, errors.Newf("%s: no component specified", ref)
		}
		if n, err := lister.NumComponents(""); n != 1 {
			if err != nil {
				return evaluated, errors.Wrapf(err, "%s: listing components", ref)
			}
			// return evaluated, errors.Newf("%s: found %d components", ref, n)
			return evaluated, nil // return repo ref
		}
		list, err := lister.GetComponents("", true)
		if err != nil {
			return evaluated, errors.Wrapf(err, "%s: listing components", ref)
		}
		evaluated.Ref.Component = list[0]
		evaluated.Component, err = s.LookupComponent(evaluated.Repository, list[0])
		if err != nil {
			return evaluated, errors.Wrapf(err, "%s: listing components", ref)
		}
	}
	return evaluated, nil
}

func (s *session) EvaluateRef(ctx Context, ref string) (*EvaluationResult, error) {
	var err error
	result := &EvaluationResult{}
	result.Ref, err = ParseRef(ref)
	if err != nil {
		return nil, err
	}

	result.Repository, err = s.DetermineRepositoryBySpec(ctx, &result.Ref.UniformRepositorySpec)
	if err != nil {
		return result, err
	}
	if result.Ref.Component != "" {
		result.Component, err = s.LookupComponent(result.Repository, result.Ref.Component)
		if err != nil {
			return nil, err
		}
		if result.Ref.IsVersion() {
			result.Version, err = s.GetComponentVersion(result.Component, *result.Ref.Version)
		}
	}
	return result, err
}

func (s *session) DetermineRepository(ctx Context, ref string) (Repository, UniformRepositorySpec, error) {
	spec, err := ParseRepo(ref)
	if err != nil {
		return nil, spec, err
	}
	r, err := s.DetermineRepositoryBySpec(ctx, &spec)
	return r, spec, err
}

func (s *session) DetermineRepositoryBySpec(ctx Context, spec *UniformRepositorySpec) (Repository, error) {
	rspec, err := ctx.MapUniformRepositorySpec(spec)
	if err != nil {
		return nil, err
	}
	return s.LookupRepository(ctx, rspec)
}
