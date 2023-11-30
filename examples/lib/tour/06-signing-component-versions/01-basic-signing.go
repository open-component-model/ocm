// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

func SigningComponentVersions(cfg *helper.Config) error {

	ctx := ocm.DefaultContext()

	// Configure context with optional ocm config.
	// See OCM config scenario in tour 04.
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return err
	}

	// siginfo := signingattr.Get(ctx)

	// to sign a component version we need a private key.
	// for this example, we just create a local keypair.
	// to be able to verify later, we should save the public key,
	// but here we do all this in a single program.

	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create keypair")
	}

	// now we compose a component version without a repository, again.
	// see tour02 example b.
	cv := composition.NewComponentVersion(ctx, "acme.org/example6", "v0.1.0")

	// just use the same component version setup again
	err = setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "version composition")
	}

	fmt.Printf("*** composition version ***\n")
	err = describeVersion(cv)

	// now let's sign the component version.
	// There might be multiple signatures, therefore every signature
	// has a name (here acme.org). Keys are always specified for
	// a dedicated signature name.
	_, err = signing.SignComponentVersion(cv, "acme.org", signing.PrivateKey("acme.org", privkey))
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}
	fmt.Printf("*** signed composition version ***\n")
	err = describeVersion(cv)

	// now add the signed component to a target repository.
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

	// let's check the target for the new component version
	tcv, err := target.LookupComponentVersion("acme.org/example6", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "transported version not found")
	}
	defer tcv.Close()

	// please be aware that the signature should be stored.
	fmt.Printf("*** target version in transportation target\n")
	err = describeVersion(tcv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}

	// new lets verify the signature
	_, err = signing.VerifyComponentVersion(cv, "acme.org", signing.PublicKey("acme.org", pubkey))
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("verification succeeded\n")
	}
	return nil
}
