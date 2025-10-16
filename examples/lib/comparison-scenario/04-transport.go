package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/examples/lib/helper"
)

func Transport(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot register credentials")
	}

	err = TransportComponentVersion(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "transport failed")
	}
	return nil
}

func TransportComponentVersion(ctx ocm.Context, cfg *helper.Config) error {
	fmt.Printf("*** transporting component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

	// get the access to the source repository
	fmt.Printf("source OCI repository is %s\n", string(cfg.Repository))
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return errors.Wrapf(err, "cannot get repository access for %s", cfg.Repository)
	}
	defer repo.Close()

	// lookup component version to be transported
	cv, err := repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION)
	if err != nil {
		return errors.Wrapf(err, "cannot get component version from %s", cfg.Repository)
	}
	defer cv.Close()
	PrintSignatures(cv)

	// use the generic form here to enable the specification of any
	// supported repository type as target.
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	target, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer target.Close()

	err = transfer.Transfer(cv, target,
		standard.ResourcesByValue(),
		standard.Overwrite(),
		transfer.WithPrinter(common.StdoutPrinter))
	if err != nil {
		return errors.Wrapf(err, "transfer failed")
	}
	return nil
}
