// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
)

func UsingCredentialsRepositories(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()
	credctx := ctx.CredentialsContext()

	// The OCM toolset embraces multiple storage
	// backend technologies, for OCM meta data as well
	// as for artifacts described by a component version.
	// All those technologies typically have their own
	// way to configure credentials for command line
	// tools or servers.
	//
	// The credential management provides so-called
	// credential repositories. Such a repository
	// is able to provide any number of names
	// credential sets. This way any special
	// credential store can be connected to the
	// OCM credential management jsu by providing
	// an own implementation for the repository interface.

	// One such case is the docker config json, a config
	// file used by <code>docker login</code> to store
	// credentials for dedicatd OCI regsitries.
	dspec := dockerconfig.NewRepositorySpec("~/.docker/config.json")

	// There are general credential stores, like a HashiCorp Vault
	// or type-specific ones, like the docker config json
	// used to configure credentials for the docker client.
	// (working with OCI registries).
	// Those specialized repository implementation are not only able to
	// provide credential sets, they also know about the usage context.
	// Such repository implementations are able to provide credential
	// mappings for consumer ids, also.

	// The docker config is such a case, so we can instruct the
	// repository to automatically propagate appropriate the consumer id
	// mappings.
	dspec = dspec.WithConsumerPropagation(true)

	// now we can jsut add the repository for this specification to
	// the credential context.
	_, err := credctx.RepositoryForSpec(dspec)
	if err != nil {
		return errors.Wrapf(err, "invalid credential repository")
	}
	// we are not interested in the repository object, so we just ignore
	// the result.

	// so, if you have done the appropriate docker login for your
	// OCI registry, it should be possible now to get the credentials
	// for the configured repository.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}

	// the returned credentials are provided via an interface, which might change its
	// content, if the underlying credential source changes.
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

	return nil
}
