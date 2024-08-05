package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/examples/lib/helper"
)

func GetDeployScript(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot register credentials")
	}

	err = DownloadDeployScript(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "deploy script failed")
	}
	return nil
}

func DownloadDeployScript(ctx ocm.Context, cfg *helper.Config) error {
	fmt.Printf("*** get deploy script\n")

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

	// get resource and blob content
	res, err := cv.GetResource(metav1.NewIdentity(RSC_DEPLOY))
	if err != nil {
		return errors.Wrapf(err, "resource for podinfo-image not found")
	}

	data, err := utils.GetResourceData(res)
	if err != nil {
		return errors.Wrapf(err, "cannot get deployscript")
	}

	fmt.Printf("deploy script:\n%s\n", string(data))
	return nil
}
