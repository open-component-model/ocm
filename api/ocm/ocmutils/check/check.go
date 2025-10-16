package check

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	common "ocm.software/ocm/api/utils/misc"
)

type Result struct {
	Missing   Missing           `json:"missing,omitempty"`
	Resources []metav1.Identity `json:"resources,omitempty"`
	Sources   []metav1.Identity `json:"sources,omitempty"`
}

func newResult() *Result {
	return &Result{Missing: Missing{}}
}

func (r *Result) IsEmpty() bool {
	if r == nil {
		return true
	}
	return len(r.Missing) == 0 && len(r.Resources) == 0 && len(r.Sources) == 0
}

type Missing map[common.NameVersion]common.History

func (n Missing) MarshalJSON() ([]byte, error) {
	m := map[string]common.History{}
	for k, v := range n {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

type Cache = map[common.NameVersion]*Result

////////////////////////////////////////////////////////////////////////////////

// Check provides a check object for checking component versions
// to completely available in an ocm repository.
// By default, it only checks the component reference closure
// to be in the same repository.
// Optionally, it is possible to check for inlined
// resources and sources, also.
func Check(opts ...Option) *Options {
	return optionutils.EvalOptions(opts...)
}

func (a *Options) For(cv ocm.ComponentVersionAccess) (*Result, error) {
	cache := Cache{}
	return a.handle(cache, cv, common.History{common.VersionedElementKey(cv)})
}

func (a *Options) ForId(repo ocm.Repository, id common.NameVersion) (*Result, error) {
	cv, err := repo.LookupComponentVersion(id.GetName(), id.GetVersion())
	if err != nil {
		return nil, err
	}
	defer cv.Close()
	return a.For(cv)
}

func (a *Options) check(cache Cache, repo ocm.Repository, id common.NameVersion, h common.History) (*Result, error) {
	if r, ok := cache[id]; ok {
		return r, nil
	}

	err := h.Add(ocm.KIND_COMPONENTVERSION, id)
	if err != nil {
		return nil, err
	}
	cv, err := repo.LookupComponentVersion(id.GetName(), id.GetVersion())
	if err != nil {
		if !errors.IsErrNotFound(err) {
			return nil, err
		}
		err = nil
	}

	var r *Result
	if cv == nil {
		r = &Result{Missing: Missing{id: h}}
	} else {
		defer cv.Close()
		r, err = a.handle(cache, cv, h)
	}
	cache[id] = r
	return r, err
}

func (a *Options) handle(cache Cache, cv ocm.ComponentVersionAccess, h common.History) (*Result, error) {
	result := newResult()

	for _, r := range cv.GetDescriptor().References {
		id := common.NewNameVersion(r.ComponentName, r.Version)
		n, err := a.check(cache, cv.Repository(), id, h)
		if err != nil {
			return result, err
		}
		if n != nil && len(n.Missing) > 0 {
			for k, v := range n.Missing {
				result.Missing[k] = v
			}
		}
	}

	var err error

	list := errors.ErrorList{}
	if optionutils.AsBool(a.CheckLocalResources) {
		result.Resources, err = a.checkArtifacts(cv.GetContext(), cv.GetDescriptor().Resources)
		list.Add(err)
	}
	if optionutils.AsBool(a.CheckLocalSources) {
		result.Sources, err = a.checkArtifacts(cv.GetContext(), cv.GetDescriptor().Sources)
		list.Add(err)
	}
	if result.IsEmpty() {
		result = nil
	}
	return result, list.Result()
}

func (a *Options) checkArtifacts(ctx ocm.Context, accessor compdesc.ElementListAccessor) ([]metav1.Identity, error) {
	var result []metav1.Identity

	list := errors.ErrorList{}
	for i := 0; i < accessor.Len(); i++ {
		e := accessor.Get(i).(compdesc.ElementArtifactAccessor)

		m, err := ctx.AccessSpecForSpec(e.GetAccess())
		if err == nil {
			if !m.IsLocal(ctx) {
				result = append(result, e.GetMeta().GetIdentity(accessor))
			}
		} else {
			list.Add(err)
		}
	}
	return result, list.Result()
}
