// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
)

func GetRef(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot register credentials")
	}

	err = GetOCIReference(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "download failed")
	}
	return nil
}

func GetOCIReference(ctx ocm.Context, cfg *helper.Config) error {
	fmt.Printf("*** get OCI reference\n")

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

	// get resource and get blob content
	res, err := cv.GetResource(metav1.NewIdentity("podinfo-image"))
	if err != nil {
		return errors.Wrapf(err, "resource for podinfo-image not found")
	}

	acc, err := res.Access()
	if err == nil {
		data, _ := json.Marshal(acc)
		fmt.Printf("access: %s\n", string(data))
	}
	ref, err := utils.GetOCIArtifactRef(ctx, res)
	if err != nil {
		return errors.Wrapf(err, "cannot get OCI reference for resource")
	}
	if ref != "" {
		fmt.Printf("OCI reference: %s\n", ref)
	} else {
		fmt.Printf("no OCI reference found\n")
	}

	meth, err := res.AccessMethod()
	if err != nil {
		return err
	}
	defer meth.Close()
	PrintConsumerId(meth, "OCI image")
	return nil
}
