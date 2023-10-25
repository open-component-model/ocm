// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/open-component-model/ocm/examples/lib/helper"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/errors"
)

func UsingCredentialsA(cfg *helper.Config) error {
	// yes, we need an OCM context, again
	ctx := ocm.DefaultContext()

	// So far, we just use memory or filesystem based
	// OCM repositories to create component versions.
	// If we want to store something in a remotely accessible
	// repository typically some credentials are required.
	//
	// The OCM library uses a generic abstraction for credentials-
	// It is just set of properties. To offer various credential sources
	// There is an interface credentials.Credentials provides,
	// whose implementations provide access to those properties.
	// A simple property based implementation is credentials.DirectCredentials.
	//
	// The most simple use case is to provide the credentials
	// directly for the repository access creation.
	// The example config file provides such credentials
	// for an OCI registry.

	creds := ociidentity.SimpleCredentials(cfg.Username, cfg.Password)

	// now we can use the OCI repository access creation from
	// example, but we pass the credentials as additional parameter.
	// To give you the chance to specify your own registry the URL
	// is taken from the config file.
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)

	repo, err := ctx.RepositoryForSpec(spec, creds)
	if err != nil {
		return err
	}
	defer repo.Close()

	// if registry name and credentials are fine, we should be able
	// now to add a new component version using the coding
	// from the previous example.

	// now we create a component version in this repository.
	err = addVersion(repo, "acme.org/example03", "v0.1.0")
	if err != nil {
		return err
	}

	// list the versions as known from example 1
	// OCI registries do not support component listers, therefore we
	// just list the actually added version.
	cv, err := repo.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "added version not found")
	}
	return describeVersion(cv)
}
