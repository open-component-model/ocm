package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/examples/lib/helper"
)

// --- begin read config ---
func ReadConfiguration(ctx ocm.Context, cfg *helper.Config) error {
	if cfg.OCMConfig != "" {
		fmt.Printf("*** applying config from %s\n", cfg.OCMConfig)

		_, err := utils.Configure(ctx, cfg.OCMConfig)
		if err != nil {
			return errors.Wrapf(err, "error in ocm config %s", cfg.OCMConfig)
		}
	}
	return nil
}

// --- end read config ---

func TransportingComponentVersions(cfg *helper.Config) error {
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

	// the context acts as factory for various model types based on
	// specification descriptor serialization formats in YAML or JSON.
	// Access method specifications and repository specification are
	// examples for this feature.
	//
	// Now, we use the repository specification serialization format to
	// determine the target repository for a transport from our yaml
	// configuration file.
	// --- begin target ---
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	target, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer target.Close()
	// --- end target ---

	// we just use the component version provided by the last examples
	// in a remote target repository.
	// Therefore, we set up the credentials context, again, as has
	// been shown in example 3.
	// --- begin set credentials ---
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	creds := ociidentity.SimpleCredentials(cfg.Username, cfg.Password)
	ctx.CredentialsContext().SetCredentialsForConsumer(id, creds)
	// --- end set credentials ---

	// now, we are ready to determine the transportation source.

	// For the transport, we first get access to the component version
	// we want to transport, by getting the source repository and looking up
	// the desired component version.
	// --- begin source ---
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec, creds)
	if err != nil {
		return err
	}
	defer repo.Close()

	cv, err := repo.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "added version not found")
	}
	defer cv.Close()
	// --- end source ---

	fmt.Printf("*** source version in source repository\n")
	err = describeVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}

	// transfer the component version with value mode.
	// Here, all resources are transported per value, all external
	// references will be inlined as `localBlob` and imported into
	// the target environment, applying blob upload handlers
	// where possible. For a CTF archive as target, there are no
	// configured handlers by default.
	// --- begin transfer ---
	err = transfer.Transfer(cv, target, standard.ResourcesByValue(), standard.Overwrite())
	if err != nil {
		return errors.Wrapf(err, "transport failed")
	}
	// --- end transfer ---

	// now, we check the result of our transport action in the target
	// repository
	// --- begin verify-a ---
	tcv, err := target.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "transported version not found")
	}
	defer tcv.Close()
	// --- end verify-a ---

	// please be aware that all resources in the target now are localBlobs,
	// if the target is a CTF archive. If it is an OCI registry, all the OCI
	// artifact resources will be uploaded as OCI artifacts into the target
	// repository and the access specifications are adapted to type `ociArtifact`.
	// --- begin verify-b ---
	fmt.Printf("*** target version in transportation target\n")
	err = describeVersion(tcv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}
	// --- end verify-b ---
	return nil
}
