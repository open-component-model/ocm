package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/examples/lib/helper"
)

func Verify(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot register credentials")
	}
	err = RegisterPublicKeyFromFile(ctx)
	if err != nil {
		return errors.Wrapf(err, "cannot register public key")
	}

	err = VerifyComponentVersion(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "transport failed")
	}
	return nil
}

func VerifyComponentVersion(ctx ocm.Context, cfg *helper.Config) error {
	fmt.Printf("*** verifying component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

	// use the generic form here to enable the specification of any
	// supported repository type as target.
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	repo, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer repo.Close()

	// lookup component version to be transported
	cv, err := repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION)
	if err != nil {
		return errors.Wrapf(err, "cannot get component version from %s", cfg.Target)
	}
	defer cv.Close()

	_, err = signing.VerifyComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("*** verification successful\n")
	}
	return nil
}
