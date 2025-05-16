package ocmutils

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils/misc"
)

// WalkingStep is used to process a component version during graph traversal.
// If returned true, the traversal process follows local component references.
// If an error is returned the traversal is aborted with this error.
// Additionally, an info object of type T can be registered in the state for the
// component version.
type WalkingStep[T any] func(state misc.WalkingState[T, ocm.ComponentVersionAccess], cv ocm.ComponentVersionAccess) (bool, error)

// Walk traverses a component version graph using the WalkingStep to
// process found component version.
func Walk[T any](closure misc.NameVersionInfo[T], cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, step WalkingStep[T]) (misc.NameVersionInfo[T], error) {
	if closure == nil {
		closure = misc.NameVersionInfo[T]{}
	}
	state := misc.WalkingState[T, ocm.ComponentVersionAccess]{
		Closure: closure,
		Context: cv,
	}
	err := walk[T](state, cv, resolver, step)
	return closure, err
}

func walk[T any](state misc.WalkingState[T, ocm.ComponentVersionAccess], cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, step WalkingStep[T]) error {
	nv := misc.VersionedElementKey(cv)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok || err != nil {
		return err
	}
	c, err := step(state, cv)
	if err != nil {
		return errors.Wrapf(err, "%s", state.History)
	}
	if c {
		for _, ref := range cv.GetDescriptor().References {
			n, err := resolver.LookupComponentVersion(ref.ComponentName, ref.Version)
			if err != nil {
				return errors.Wrapf(err, "%s: cannot resolve ref %s", state.History, ref)
			}
			err = errors.Join(walk[T](state, n, resolver, step), n.Close())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
