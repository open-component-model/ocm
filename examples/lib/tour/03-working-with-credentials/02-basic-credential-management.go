package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/examples/lib/helper"
)

func UsingCredentialsB(cfg *helper.Config, create bool) error {
	// --- begin default context ---
	ctx := ocm.DefaultContext()
	// --- end default context ---

	// Passing credentials directly at the repository
	// is fine, as long only the component version
	// will be accessed. But as soon as described
	// resource content will be read, the required
	// credentials and credential types are dependent
	// on the concrete component version, because
	// it might contain any kind of access method
	// referring to any kind of resource repository
	// type.
	//
	// To solve this problem of passing any set
	// of credentials the OCM context object is
	// used to store credentials. This is handled
	// by a sub context, the Credentials context.

	// --- begin cred context ---
	credctx := ctx.CredentialsContext()
	// --- end cred context ---

	// The credentials context brings together
	// providers of credentials, for example a
	// Vault or a local Docker config.json
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
	// an OCI registry is identified by a host and
	// a repository path.
	//
	// A credential provider like a vault just provides
	// named credential sets and typically does not
	// know anything about the use case for these sets.
	// The task of the credential context is to
	// provide credentials for a dedicated consumer.
	// Therefore, it maintains a configurable
	// mapping of credential sources (credentials in
	// a credential repository) and a dedicated consumer.
	//
	// This mapping defines a use case, also based on
	// a property set and dedicated credentials.
	// If credentials are required for a dedicated
	// consumer, it matches the defined mappings and
	// returned the best matching entry.
	//
	// Matching? Let's take the GitHub OCI registry as an
	// example. There are different owners for
	// different repository paths (the GitHub org/user).
	// Therefore, different credentials need to be provided
	// for different repository paths.
	// For example, credentials for ghcr.io/acme can be used
	// for a repository ghcr.io/acme/ocm/myimage.

	// To start with the credentials context we just
	// provide an explicit mapping for our use case.

	// first, we create our credentials object as before.
	// --- begin new credentials ---
	creds := identity.SimpleCredentials(cfg.Username, cfg.Password)
	// --- end new credentials ---

	// Then we determine the consumer id for our use case.
	// The repository implementation provides a function
	// for this task. It provides the most general property
	// set for an OCI based OCM repository.
	// --- begin consumer id ---
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	// --- end consumer id ---

	// the used functions above are just convenience wrappers
	// around the core type ConsumerId, which might be provided
	// for dedicated repository/consumer technologies.
	// everything can be done directly with the core interface.

	// --- begin set credentials ---
	credctx.SetCredentialsForConsumer(id, creds)
	// --- end set credentials ---

	// now the context is prepared to provide credentials
	// for any usage of our OCI registry, regardless
	// of its type.

	// let's test, whether it could provide credentials
	// for storing our component version.

	// first we get the repository object for our OCM repository.
	// --- begin get repository ---
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec, creds)
	if err != nil {
		return err
	}
	defer repo.Close()
	// --- end get repository ---

	// second, we determine the consumer id for our intended repository access.
	// a credential consumer may provide consumer id information
	// for a dedicated sub user context.
	// This is supported by the OCM repo implementation for OCI registries.
	// The usage context is here the component name.

	// --- begin get access id ---
	id = credentials.GetProvidedConsumerId(repo, credentials.StringUsageContext("acme.org/example03"))
	if id == nil {
		return fmt.Errorf("repository does not support consumer id queries")
	}
	fmt.Printf("usage context: %s\n", id)
	// --- end get access id ---

	// third, we ask the credential context for appropriate credentials.
	// the basic context method `credctx.GetCredentialsForConsumer` returns
	// a credentials source interface able to provide credentials
	// for a changing credentials source. Here, we use a convenience
	// function, which directly provides a credentials interface for the
	// actually valid credentials.
	// an error is only provided if something went wrong while determining
	// the credentials. Delivering NO credentials is a valid result.
	// the returned interface then offers access to the credential properties.
	// via various methods.

	// --- begin get credentials ---
	creds, err = credentials.CredentialsForConsumer(credctx, id, identity.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "no credentials")
	}
	if creds == nil {
		return fmt.Errorf("no credentials found")
	}
	fmt.Printf("credentials: %s\n", obfuscate(creds.Properties()))
	// --- end get credentials ---

	// Now we can continue with our basic component version composition
	// from the last example, or we just display the content.

	// --- begin add version ---
	if create {
		// now we create a component version in this repository.
		err = addVersion(repo, "acme.org/example03", "v0.1.0")
		if err != nil {
			return err
		}
	}
	// --- end add version ---

	// list the versions as known from example 1
	// OCI registries do not support component listers, therefore we
	// just get and describe the actually added version.
	// --- begin show version ---
	cv, err := repo.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "added version not found")
	}
	defer cv.Close()

	err = describeVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}
	// --- end show version ---

	// as you have seen in the resource list, our image artifact has been
	// uploaded to the OCI registry and the access method has be changed
	// to ociArtifact.
	// It is not longer a local blob.

	// --- begin examine cli ---
	res, err := cv.SelectResources(selectors.Name("ocmcli"))
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
	// --- end examine cli ---

	// this resource access effectively points to the ame OCI registry,
	// but a completely different repository.
	// If you are using ghcr.io, this freshly created repo is private,
	// therefore, you need credentials for accessing the content.
	// An access method also acts as credential consumer, which
	// tries to get required credentials from the credential context.
	// Optionally, an access method can act as provider for a consumer id, so that
	// it is possible to query the used consumer id from the method object.

	// --- begin image credentials ---
	id = credentials.GetProvidedConsumerId(meth, credentials.StringUsageContext("acme.org/example3"))
	if id == nil {
		fmt.Printf("no consumer id info for access method\n")
	} else {
		fmt.Printf("usage context: %s\n", id)
	}
	// --- end image credentials ---

	// Because the credentials context now knows the required credentials,
	// the access method as credential consumer can access the blob.

	// --- begin image access ---
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
	// --- end image access ---
	return nil
}
