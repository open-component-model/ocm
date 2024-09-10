package main

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/examples/lib/helper"
)

func UsingCredentialsA(cfg *helper.Config) error {
	// yes, we need an OCM context, again
	// --- begin default context ---
	ctx := ocm.DefaultContext()
	// --- end default context ---

	// So far, we just used memory or file system based
	// OCM repositories to create component versions.
	// If we want to store something in a remotely accessible
	// repository typically some credentials are required
	// for write access.
	//
	// The OCM library uses a generic abstraction for credentials.
	// It is just set of properties. To offer various credential sources
	// there is an interface credentials.Credentials provided,
	// whose implementations provide access to those properties.
	// A simple property based implementation is credentials.DirectCredentials.
	//
	// The most simple use case is to provide the credentials
	// directly for the repository access creation.
	// The example config file provides such credentials
	// for an OCI registry.

	// --- begin new credentials ---
	creds := ociidentity.SimpleCredentials(cfg.Username, cfg.Password)
	// --- end new credentials ---

	// now we can use the OCI repository access creation from the first tour,
	// but we pass the credentials as additional parameter.
	// To give you the chance to specify your own registry, the URL
	// is taken from the config file.
	// --- begin repository access ---
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)

	repo, err := ctx.RepositoryForSpec(spec, creds)
	if err != nil {
		return err
	}
	defer repo.Close()
	// --- end repository access ---

	// if registry name and credentials are fine, we should be able
	// now to add a new component version using the coding
	// from the previous example, but now we use a public repository, instead
	// of a memory or file system based one.

	// now we create a component version in this repository.
	err = addVersion(repo, "acme.org/example03", "v0.1.0")
	if err != nil {
		return err
	}

	// In contrast to our first tour we cannot list components, here.
	// OCI registries do not support component listers, therefore we
	// just look up the actually added version to verify the result.
	// --- begin lookup ---
	cv, err := repo.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "added version not found")
	}
	defer cv.Close()
	return errors.Wrapf(describeVersion(cv), "describe failed")
	// --- end lookup ---
}
