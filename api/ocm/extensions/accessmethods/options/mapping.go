package options

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
)

func MapRepository(in any) (any, error) {
	uni, err := cpi.ParseRepo(in.(string))
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, cpi.KIND_REPOSITORYSPEC, in.(string))
	}

	// TODO: basically a context is required, here.
	spec, err := cpi.DefaultContext().MapUniformRepositorySpec(&uni)
	if err != nil {
		return nil, err
	}
	return cpi.ToGenericRepositorySpec(spec)
}

func MapResourceRef(in any) (any, error) {
	list := in.([]v1.Identity)
	if len(list) == 0 {
		return nil, fmt.Errorf("empty resource reference")
	}
	return v1.NewResourceRef(list[0], list[1:]...), nil
}
