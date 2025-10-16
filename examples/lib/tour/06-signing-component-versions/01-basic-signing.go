package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/examples/lib/helper"
)

func SigningComponentVersions(cfg *helper.Config) error {
	// --- begin default context ---
	ctx := ocm.DefaultContext()
	// --- end default context ---

	// Configure context with optional ocm config.
	// See OCM config scenario in tour 04.
	// --- begin configure ---
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return err
	}
	// --- end configure ---

	// to sign a component version we need a private key.
	// for this example, we just create a local keypair.
	// to be able to verify later, we should save the public key,
	// but here we do all this in a single program.

	// --- begin create keypair ---
	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create keypair")
	}
	// --- end create keypair ---

	// now we compose a component version without a repository, again.
	// see tour02 example b.
	// --- begin compose ---
	cv := composition.NewComponentVersion(ctx, "acme.org/example6", "v0.1.0")

	// just use the same component version setup again
	err = setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "version composition")
	}

	fmt.Printf("*** composition version ***\n")
	err = describeVersion(cv)
	// --- end compose ---

	// now let's sign the component version.
	// There might be multiple signatures, therefore every signature
	// has a name (here acme.org). Keys are always specified for
	// a dedicated signature name.
	// --- begin sign ---
	_, err = signing.SignComponentVersion(cv, "acme.org", signing.PrivateKey("acme.org", privkey))
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}
	fmt.Printf("*** signed composition version ***\n")
	err = describeVersion(cv)
	// --- end sign ---

	// now add the signed component to a target repository.
	// here, we just reuse the code from tour02
	// --- begin add version ---
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
	// --- end add version ---

	// let's check the target for the new component version
	// --- begin lookup ---
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
	// --- end lookup ---

	// now lets verify the signature
	// --- begin verify ---
	_, err = signing.VerifyComponentVersion(cv, "acme.org", signing.PublicKey("acme.org", pubkey))
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("verification succeeded\n")
	}
	// --- end verify ---
	return nil
}
