// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

func PrintPublicKey(ctx ocm.Context, name string) {
	info := signingattr.Get(ctx)
	key := info.GetPublicKey(name)
	if key == nil {
		fmt.Printf("public key for %s not found\n", name)
	} else {
		buf := bytes.NewBuffer(nil)
		err := rsa.WriteKeyData(key, buf)
		if err != nil {
			fmt.Printf("key error: %s\n", err)
		} else {
			fmt.Printf("public key for %s:\n%s\n", name, buf.String())
		}
	}
}

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

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** publishing component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

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
