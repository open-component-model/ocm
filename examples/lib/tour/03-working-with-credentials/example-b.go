// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/errors"
)

func obfuscate(props common.Properties) string {
	if pw, ok := props[credentials.ATTR_PASSWORD]; ok {
		if len(pw) > 5 {
			pw = pw[:5] + "***"
		} else {
			pw = "***"
		}
		props = props.Copy()
		props[credentials.ATTR_PASSWORD] = pw
	}
	return props.String()
}

func UsingCredentialsB(cfg *helper.Config, create bool) error {
	ctx := ocm.DefaultContext()

	// Passing credentials directly at the respository
	// is fine, as long only the component version
	// will be accessed. But as soon as described
	// resource content will be read, the required
	// credentials and credential types are dependent
	// on the concrete conmponent version, because
	// it might contain any kind of access method
	// referring to any kind of resource repository
	// type.
	//
	// To solve this problem of passing any set
	// of credentials the OCM context object is
	// used to store credentials. This handled
	// by a sub context, the Credentials context.

	credctx := ctx.CredentialsContext()

	// The credentials context brings together
	// provider of credentials, for example a
	// vault or a local docker/config.json
	// and credential consumers like GitHub or
	// OCI registries.
	// It must be able to distinguish various kinds
	// of consumers. This is done by identifying
	// a dedicated consumer with a set of properties
	// called credentials.ConsumerId. It consists
	// at least of a consumer type property and a
	// consumer type specific set of properties
	// describing the concrete instance of such
	// a consumer, for example an OCI artifact in
	// an OCI registry s identified by a host and
	// a repository path.
	//
	// A credential provider like a vault just provides
	// named credential set and typically does not
	// know anything about the use case for these sets.
	// The task of the credential context is now to
	// provide credentials for a dedicated consumer.
	// Therefore, it maintains a configurable
	// mapping of credential sources (credentials in
	// a credential repository) and a dedicated consumer.
	//
	// This mapping defines a usecase, also based on
	// a property set and dedicated credentials.
	// If credentials are required for a dedicated
	// consumer, it matches the defined mappings and
	// returned the best matching entry.
	//
	// Matching? Lets take GitHub OCI registry as an
	// example. There are different owners for
	// different repository path (the GitHub org/user).
	// Therfore, different credentials needs to be provided
	// for different repository paths.
	// For example credentials for ghcr.io/acme can be used
	// for a repository ghcr.io/acme/ocm/myimage.

	// To start with the credentials context we just
	// provide an explicit mapping for our use case.

	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	creds := ociidentity.SimpleCredentials(cfg.Username, cfg.Password)
	// the used functions above are just convenience wrappers
	// arround the core type ConsumerId, which might be provided
	// which might be for dedicated repository technologies.
	// everything can be done directly with the core interface.

	credctx.SetCredentialsForConsumer(id, creds)

	// now the context is prepared to provide credentials
	// for any usage of our OCI registry, regardless
	// of its type.

	// lets test, whether it could provide credentials
	// for storing our component version.

	// first we get the repository object for our OCM repository.
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec, creds)
	if err != nil {
		return err
	}
	defer repo.Close()

	// a credential consumer may provide might provide consumer id information
	// for a dedicated sub user context.
	// This is supported by the OCM repo implementation for OCI registres.
	// The usage context is here the component name.
	id = credentials.GetProvidedConsumerId(repo, credentials.StringUsageContext("acme.org/example3"))
	if id == nil {
		return fmt.Errorf("repository does not support consumer id queries")
	}
	fmt.Printf("usage context: %s\n", id)

	// the returned credentials are provided via an interface, which might change its
	// content, if the underlying credential source changes.
	creds, err = credentials.CredentialsForConsumer(credctx, id, ociidentity.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "no credentials")
	}
	fmt.Printf("credentials: %s\n", obfuscate(creds.Properties()))

	// Now we can continue with our basic component version composition
	// from the last example, or we just display the content.

	if create {
		// now we create a component version in this repository.
		err = addVersion(repo, "acme.org/example03", "v0.1.0")
		if err != nil {
			return err
		}
	}

	// list the versions as known from example 1
	// OCI registries do not support component listers, therefore we
	// just list the actually added version.
	cv, err := repo.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "added version not found")
	}
	defer cv.Close()

	err = describeVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}

	// as you have seen in the resource list, out image artifact has been
	// uploaded to the OCI registry and the access method has be changed
	// to ociArtifact.
	// It is not longer a local blob.

	res, err := cv.GetResourcesByName("ocmcli")
	if err != nil {
		return errors.Wrapf(err, "accessing ocmcli resource")
	}
	if len(res) != 1 {
		return fmt.Errorf("oops, there are %d entries for ocmcli", len(res))
	}
	meth, err := res[0].AccessMethod()
	if err != nil {
		return errors.Wrapf(err, "cannot get access method")
	}
	defer meth.Close()

	fmt.Printf("accessing oci image now with %s\n", meth.AccessSpec().Describe(ctx))

	// this resource access points effectively to the same repository.
	// If you are using ghcr.io, this freshly created repo is private,
	// therefore, you need credentials for accessing the content.
	// Because the credentials context now knows the required credentials,
	// the access method as credential consumer can access the blob.

	id = credentials.GetProvidedConsumerId(meth, credentials.StringUsageContext("acme.org/example3"))
	if id == nil {
		fmt.Printf("no consumer id info for access method\n")
	} else {
		fmt.Printf("usage context: %s\n", id)
	}

	writer, err := os.OpenFile("/tmp/example3", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Wrapf(err, "cannot write output file")
	}
	defer writer.Name()

	reader, err := meth.Reader()
	if err != nil {
		return errors.Wrapf(err, "cannot get reader")
	}
	defer reader.Close()
	n, err := io.Copy(writer, reader)
	if err != nil {
		return errors.Wrapf(err, "cannot copy content")
	}
	fmt.Printf("blob has %d bytes\n", n)
	return nil
}
