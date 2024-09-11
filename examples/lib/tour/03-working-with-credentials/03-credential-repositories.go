package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/dockerconfig"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/examples/lib/helper"
)

func UsingCredentialsRepositories(cfg *helper.Config) error {
	// --- begin context ---
	ctx := ocm.DefaultContext()
	credctx := ctx.CredentialsContext()
	// --- end context ---

	// The OCM toolset embraces multiple storage
	// backend technologies, for OCM meta data as well
	// as for artifacts described by a component version.
	// All those technologies typically have their own
	// way to configure credentials for command line
	// tools or servers.
	//
	// The credential management provides so-called
	// credential repositories. Such a repository
	// is able to provide any number of named
	// credential sets. This way any special
	// credential store can be connected to the
	// OCM credential management just by providing
	// an own implementation for the repository interface.

	// One such case is the docker config json, a config
	// file used by <code>docker login</code> to store
	// credentials for dedicated OCI registries.

	// --- begin docker config ---
	dspec := dockerconfig.NewRepositorySpec("~/.docker/config.json")
	// --- end docker config ---

	// There are general credential stores, like a HashiCorp Vault
	// or type-specific ones, like the docker config json
	// used to configure credentials for the docker client.
	// (working with OCI registries).
	// Those specialized repository implementation are not only able to
	// provide credential sets, they also know about the usage context
	// of the provided credentials
	// Therefore, such repository implementations are also able to provide
	// credential mappings for consumer ids. This is supported by the credential
	// repository API provided by this library.

	// The docker config is such a case, so we can instruct the
	// repository to automatically propagate appropriate the consumer id
	// mappings. This feature is typically enabled by a dedicated specification
	// option.

	// --- begin propagation ---
	dspec = dspec.WithConsumerPropagation(true)
	// --- end propagation ---

	// now we can just add the repository for this specification to
	// the credential context by getting the repository object for our
	// specification.
	// --- begin add repo ---
	_, err := credctx.RepositoryForSpec(dspec)
	if err != nil {
		return errors.Wrapf(err, "invalid credential repository")
	}
	// --- end add repo ---

	// we are not interested in the repository object, so we just ignore
	// the result.

	// so, if you have done the appropriate docker login for your
	// OCI registry, it should be possible now to get the credentials
	// for the configured repository.

	// We first query the consumer id for the repository, again.
	// --- begin get consumer id ---
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	// --- end get consumer id ---

	// and then get the credentials from the credentials context
	// like in the previous example.
	// --- begin get credentials ---
	creds, err := credentials.CredentialsForConsumer(credctx, id, ociidentity.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "no credentials")
	}
	// an error is only provided if something went wrong while determining
	// the credentials. Delivering NO credentials is a valid result.
	if creds == nil {
		return fmt.Errorf("no credentials found")
	}
	fmt.Printf("credentials: %s\n", obfuscate(creds.Properties()))
	// --- end get credentials ---
	return nil
}
