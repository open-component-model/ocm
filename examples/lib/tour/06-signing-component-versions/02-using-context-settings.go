package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	configcfg "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/examples/lib/helper"
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

	// to sign a component version we still need a private key.
	// for this example, we just create a local keypair.
	// to be able to verify later, we should save the public key,
	// but here we do all this in a single program.

	// --- begin create keypair ---
	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create keypair")
	}
	// --- end create keypair ---

	// --- begin setup ---
	err = prepareComponentInRepo(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot prepare component version in target repo")
	}
	// --- end setup ---

	// every context features a signing registry, which provides available
	// signers and  hashers, but also keys for various purposes.
	// It is always asked if a key is required, which is
	// not explicitly given to a signing/verification call.
	// This context part is implemented as additional attribute stored along
	// with the context. Attributes are always implemented as a separate package
	// containing the attribute structure, its deserialization and
	// a `Get(Context)` function to retrieve the attribute for the context.
	// This way new arbitrary attributes for various use cases can be added
	// without the need to change the context interface.
	// --- begin signing attribute ---
	siginfo := signingattr.Get(ctx)
	// --- end signing attribute ---

	// now we add the key manually to our context.
	// --- begin configure keys ---
	siginfo.RegisterPrivateKey("acme.org", privkey)
	siginfo.RegisterPublicKey("acme.org", pubkey)
	// --- end configure keys ---

	// now, we are prepared and can sign any component version
	// in any repository for the signature name acme.org.

	// just get the component version from the prepared repo.
	// --- begin lookup component version ---
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
	// --- end lookup component version ---

	// we don't need to present they key, here. It is taken from the
	// context.
	// --- begin sign ---
	_, err = signing.SignComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "cannot sign component version")
	}
	// --- end sign ---

	// please be aware that the signature should be stored.
	fmt.Printf("*** signed composition version ***\n")
	err = describeVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}

	// The same way we can just call `VerifyComponentVersion` to
	// verify the signature.
	// --- begin verify ---
	_, err = signing.VerifyComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("verification succeeded\n")
	}
	// --- end verify ---

	return createOCMConfig(privkey, pubkey)
}

func createOCMConfig(privkey signutils.GenericPrivateKey, pubkey signutils.GenericPublicKey) error {
	// manually adding keys to the signing attribute
	// might simplify the call to possibly multiple signing/verification
	// calls, but it does not help to provide keys via an external
	// configuration (for example for using the OCM CLI).
	// in tour05 we have seen how arbitrary configuration
	// possibilities can be added. The signing attribute uses
	// this mechanism to configure itself by providing an own
	// configuration object, which can be used to feed keys (and certificates)
	// into the signing attribute of an OCM context.

	// --- begin create signing config ---
	sigcfg := signingattr.New()
	// --- end create signing config ---

	// it provides methods to add elements
	// like keys and certificates. which convert
	// these elements into a (de-)serializable form.
	// --- begin add signing config ---
	sigcfg.AddPrivateKey("acme.org", privkey)
	sigcfg.AddPublicKey("acme.org", pubkey)

	ocmcfg := configcfg.New()
	ocmcfg.AddConfig(sigcfg)
	// --- end add signing config ---

	// --- begin print signing config ---
	data, err := runtime.DefaultYAMLEncoding.Marshal(ocmcfg)
	if err != nil {
		return err
	}
	fmt.Printf("ocm config file configuring standard keys:\n--- begin ocmconfig ---\n%s--- end ocmconfig ---\n", string(data))
	// --- end print signing config ---
	return nil
}
