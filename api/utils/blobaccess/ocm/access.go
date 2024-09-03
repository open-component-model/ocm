package ocm

import (
	"github.com/mandelsoft/goutils/finalizer"

	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/refmgmt"
)

func DataAccess(cvp ComponentVersionProvider, res ResourceProvider) (bpi.DataAccess, error) {
	return BlobAccess(cvp, res)
}

func BlobAccess(cvp ComponentVersionProvider, res ResourceProvider) (blob bpi.BlobAccess, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	cv, err := refmgmt.ToLazy(cvp.GetComponentVersionAccess())
	if err != nil {
		return nil, err
	}
	finalize.Close(cv)

	r, eff, err := res.GetResource(cv)
	if eff != nil {
		finalize.Close(refmgmt.AsLazy(eff))
	}
	if err != nil {
		return nil, err
	}
	return r.BlobAccess()
}

func Provider(cvp ComponentVersionProvider, res ResourceProvider) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		return BlobAccess(cvp, res)
	})
}
