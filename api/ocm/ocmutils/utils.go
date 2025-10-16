package ocmutils

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
)

func GetOCIArtifactRef(ctxp ocm.ContextProvider, r ocm.ResourceAccess) (string, error) {
	ctx := ctxp.OCMContext()

	acc, err := r.Access()
	if err != nil || acc == nil {
		return "", err
	}

	var cv cpi.ComponentVersionAccess
	if p, ok := r.(cpi.ComponentVersionProvider); ok {
		cv, err = p.GetComponentVersion()
		if err != nil {
			return "", errors.Wrapf(err, "cannot access component version for re/source")
		}
		defer cv.Close()
	}

	return ociartifact.GetOCIArtifactReference(ctx, acc, cv)
}
