package comphdlr

import (
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/semverutils"
)

func Elem(e interface{}) ocm.ComponentVersionAccess {
	return e.(*Object).ComponentVersion
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	History  common.History
	Identity metav1.Identity

	Spec       ocm.RefSpec
	Repository ocm.Repository
	// Component        ocm.ComponentAccess
	ComponentVersion ocm.ComponentVersionAccess
}

var (
	_ common.HistorySource = (*Object)(nil)
	_ tree.Object          = (*Object)(nil)
)

type Manifest struct {
	History common.History                `json:"context"`
	Element *compdesc.ComponentDescriptor `json:"element"`
}

func (o *Object) AsManifest() interface{} {
	h := o.History
	if h == nil {
		h = common.History{}
	}
	return &Manifest{
		h,
		o.ComponentVersion.GetDescriptor(),
	}
}

func (o *Object) GetHistory() common.History {
	return o.History
}

func (o *Object) IsNode() *common.NameVersion {
	var nv common.NameVersion
	if o.ComponentVersion == nil {
		nv = o.Spec.NameVersion()
	} else {
		nv = common.VersionedElementKey(o.ComponentVersion)
	}
	return &nv
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
		o.ApplyToCompHandler(h)
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
		return nil, fmt.Errorf("invalid constraints: %v", err)
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

func (h *TypeHandler) get(repo ocm.Repository, elemspec utils.ElemSpec) ([]output.Object, error) {
	var component ocm.ComponentAccess
	var result []output.Object
	var err error

	name := elemspec.String()
	spec := ocm.RefSpec{}
	if repo == nil {
		evaluated, err := h.session.EvaluateComponentRef(h.octx.Context(), name)
		if err != nil {
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
				// Component:        evaluated.Component,
				ComponentVersion: evaluated.Version,
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
		v, err := h.session.GetComponentVersion(component, *spec.Version)
		if err != nil {
			return nil, err
		}
		result = append(result, &Object{
			Repository: repo,
			Spec:       spec,
			// Component:        component,
			ComponentVersion: v,
		})
	} else {
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
				v, err := h.session.GetComponentVersion(component, vers)
				if err != nil {
					return nil, err
				}
				t := vers
				s := spec
				s.Version = &t
				result = append(result, &Object{
					Repository: repo,
					Spec:       s,
					// Component:        component,
					ComponentVersion: v,
				})
			}
		}
	}
	return result, nil
}
