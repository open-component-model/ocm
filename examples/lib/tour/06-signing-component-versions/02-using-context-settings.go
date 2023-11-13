// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

func prepareComponentInRepo(ctx ocm.Context, cfg *helper.Config) error {
	// now we compose a component version without a repository, again.
	// see tour02 example b.
	cv := composition.NewComponentVersion(ctx, "acme.org/example6", "v0.1.0")

	// just use the same component version setup again
	err := setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "version composition")
	}

	fmt.Printf("*** composition version ***\n")
	err = describeVersion(cv)

	// now add the component to a target repository.
	// here, we just reuse the code from tour05
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	target, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer target.Close()

	err = target.AddComponentVersion(cv, true)
	if err != nil {
		return errors.Wrapf(err, "cannot store signed version")
	}
	return nil
}

func SigningComponentVersionInRepo(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := prepareComponentInRepo(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot prepare component version in target repo")
	}

	// every context features a signing registry, which provides available
	// signers and  hashers, but also keys for various purposes.
	// It is always asked, if a key is required for a purpose, which is
	// not explicitly given to a signing/verification call.
	siginfo := signingattr.Get(ctx)

	// to sign a component version we need a private key.
	// for this example, we just create a local keypair.
	// to be able to verify later, we should save the public key,
	// but here we do all this in a single program.

	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create keypair")
	}

	// now we add the key to our context.
	// this can be done, for example by adding an appropriate
	// config object to your config file (see tour04).
	// here, we do it manually, just for demonstration
	siginfo.RegisterPrivateKey("acme.org", privkey)
	siginfo.RegisterPublicKey("acme.org", pubkey)

	// now, we are prepared and can sign any component version
	// in any repository for the signature name acme.org.

	// just get a the component version from the prepared repo.
	fmt.Printf("repository is %s\n", string(cfg.Target))
	repo, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer repo.Close()

	cv, err := repo.LookupComponentVersion("acme.org/example6", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "version not found")
	}
	defer cv.Close()

	// we don't need to present they key, here. It is taken from the
	// context.
	_, err = signing.SignComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}

	// please be aware that the signature should be stored.
	fmt.Printf("*** signed composition version ***\n")
	err = describeVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}

	_, err = signing.VerifyComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("verification succeeded\n")
	}
	return nil
}
