package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/examples/lib/helper"
)

func ReadConfiguration(ctx ocm.Context, cfg *helper.Config) error {
	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** applying config from %s\n", cfg.OCMConfig)

	_, err := utils.Configure(ctx, cfg.OCMConfig)
	if err != nil {
		return errors.Wrapf(err, "error in ocm config %s", cfg.OCMConfig)
	}
	return nil
}

func Provider(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot read ocm configuration")
	}

	cv, err := CreateComponentVersion(ctx)
	if err != nil {
		return errors.Wrapf(err, "cannot compose component version")
	}

	_, err = signing.SignComponentVersion(cv, SIGNATURE_NAME)
	if err != nil {
		return errors.Wrapf(err, "signing failed")
	}
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** verify signature\n")

	PrintPublicKey(ctx, "acme.org")
	_, err = signing.VerifyComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("*** verification successful\n")
	}

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** publishing component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

	// now get the access to the repository
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return errors.Wrapf(err, "cannot get repository access for %s", cfg.Repository)
	}
	defer repo.Close()

	err = repo.AddComponentVersion(cv, true)
	if err != nil {
		return errors.Wrapf(err, "cannot add version")
	}
	return nil
}
