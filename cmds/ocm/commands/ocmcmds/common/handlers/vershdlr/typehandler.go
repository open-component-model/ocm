package vershdlr

import (
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/goutils/sliceutils"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils/semverutils"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

func Elem(e interface{}) *Object {
	return e.(*Object)
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	Spec       ocm.RefSpec
	Repository ocm.Repository
	Component  string
	Version    string
}

func CompareObject(a, b output.Object) int {
	return Compare(a, b)
}

type Manifest struct {
	Component string `json:"component"`
	Version   string `json:"version"`
	Message   string `json:"error,omitempty"`
}

func (o *Object) AsManifest() interface{} {
	tag := ""
	msg := ""
	if o.Spec.Version != nil {
		tag = *o.Spec.Version
	}
	if o.Version == "" {
		msg = "<unknown component version>"
	}
	return &Manifest{
		o.Component,
		tag,
		msg,
	}
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx        clictx.OCM
	session     ocm.Session
	repobase    ocm.Repository
	resolver    ocm.ComponentVersionResolver
	constraints []*semver.Constraints
	latest      bool
}

func NewTypeHandler(octx clictx.OCM, session ocm.Session, repobase ocm.Repository, opts ...Option) utils.TypeHandler {
	h := &TypeHandler{
		octx:     octx,
		session:  session,
		repobase: repobase,
	}
	for _, o := range opts {
		o.ApplyToHandler(h)
	}
	return h
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	if h.repobase == nil {
		return nil, nil
	}
	return h.all(h.repobase)
}

func (h *TypeHandler) all(repo ocm.Repository) ([]output.Object, error) {
	lister := repo.ComponentLister()
	if lister == nil {
		return nil, nil
	}
	list, err := lister.GetComponents("", true)
	if err != nil {
		return nil, err
	}
	var result []output.Object
	sort.Strings(list)
	for _, l := range list {
		part, err := h.get(repo, utils.StringSpec(l))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
		result = append(result, part...)
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	return h.get(h.repobase, elemspec)
}

func (h *TypeHandler) filterVersions(vers []string) ([]string, error) {
	if len(h.constraints) == 0 && !h.latest {
		return vers, nil
	}
	versions, err := semverutils.MatchVersionStrings(vers, h.constraints...)
	if err != nil {
		return nil, err
	}
	if h.latest && len(versions) > 1 {
		versions = versions[len(versions)-1:]
	}
	vers = nil
	for _, v := range versions {
		vers = append(vers, v.Original())
	}
	return vers, nil
}

func (h *TypeHandler) getVersions(repo ocm.Repository, component ocm.ResolvedComponentVersionProvider, spec ocm.RefSpec) ([]output.Object, error) {
	var result []output.Object
	if component == nil {
		return h.all(repo)
	} else {
		versions, err := component.ListVersions()
		if err != nil {
			return nil, err
		}
		versions, err = h.filterVersions(versions)
		if err != nil {
			return nil, err
		}

		for _, vers := range versions {
			t := vers
			s := spec
			s.Version = &t
			result = append(result, &Object{
				Repository: repo,
				Spec:       s,
				Component:  component.GetName(),
				Version:    vers,
			})
		}
	}
	return result, nil
}

func (h *TypeHandler) get(repo ocm.Repository, elemspec utils.ElemSpec) ([]output.Object, error) {
	var component ocm.ComponentAccess
	var result []output.Object
	var err error

	name := elemspec.String()
	spec := ocm.RefSpec{}
	if repo == nil {
		evaluated, err := h.session.EvaluateComponentRef(h.octx.Context(), name)
		if err != nil {
			if !errors.IsErrNotFound(err) {
				if h.resolver != nil {
					comp, err := ocm.ParseComp(name)
					if err != nil {
						return nil, errors.Wrapf(err, "invalid component version reference %q", name)
					}
					if comp.IsVersion() {
						cv, err := h.resolver.LookupComponentVersion(comp.Component, *comp.Version)
						if err != nil {
							return nil, err
						}
						if cv != nil {
							evaluated = &ocm.EvaluationResult{}
							evaluated.Ref.UniformRepositorySpec = *cv.Repository().GetSpecification().AsUniformSpec(h.octx.Context())
							evaluated.Ref.CompSpec = comp
							evaluated.Version = cv
							evaluated.Repository = cv.Repository()
							h.session.Closer(cv)
						}
					} else {
						if r, ok := h.resolver.(ocm.ComponentResolver); ok {
							spec = evaluated.Ref
							spec.Component = comp.Component
							for _, p := range r.LookupComponentProviders(comp.Component) {
								c, err := p.LookupComponent(name)
								if err != nil {
									continue
								}

								list, err := h.getVersions(repo, c, spec)
								if err != nil {
									return nil, err
								}
								result = sliceutils.AppendUnique(result, list...)
							}
							if len(result) > 0 {
								return result, nil
							}
						}
						evaluated.Ref.Component = comp.Component
					}
				}
			}
			if evaluated == nil {
				return nil, errors.Wrapf(err, "%s: invalid component version reference", name)
			}
		}
		if evaluated.Version != nil {
			result = append(result, &Object{
				Spec:       evaluated.Ref,
				Repository: evaluated.Repository,
				Component:  evaluated.Component.GetName(),
				Version:    evaluated.Version.GetVersion(),
			})
			return result, nil
		}
		spec = evaluated.Ref
		component = evaluated.Component
		repo = evaluated.Repository
	} else {
		comp := ocm.CompSpec{Component: ""}
		if name != "" {
			comp, err = ocm.ParseComp(name)
			if err != nil {
				return nil, errors.Wrapf(err, "reference %q", name)
			}
		}
		component, err = h.session.LookupComponent(repo, comp.Component)
		if err != nil {
			return nil, errors.Wrapf(err, "reference %q", name)
		}
		spec.UniformRepositorySpec = *repo.GetSpecification().AsUniformSpec(h.octx.Context())
		spec.Component = comp.Component
		spec.Version = comp.Version
	}

	if spec.IsVersion() {
		vers := ""
		v, err := h.session.GetComponentVersion(component, *spec.Version)
		if err != nil {
			if !errors.IsErrNotFound(err) {
				return nil, err
			}
		} else {
			vers = v.GetVersion()
		}
		result = append(result, &Object{
			Repository: repo,
			Spec:       spec,
			Component:  component.GetName(),
			Version:    vers,
		})
	} else {
		if repo == nil {
			result = append(result, &Object{
				Spec:       spec,
				Repository: nil,
				Component:  spec.Component,
				Version:    optionutils.AsValue(spec.Version),
			})
			return result, nil
		}
		return h.getVersions(repo, component, spec)
	}
	return result, nil
}
