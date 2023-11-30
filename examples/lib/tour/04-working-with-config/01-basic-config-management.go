// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/go-test/deep"
	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/errors"
)

func BasicConfigurationHandling(cfg *helper.Config) error {
	// configuration is handled by the configuration context.
	ctx := config.DefaultContext()

	// the configuration context handles configuration objects.
	// a configuration object is any object implementing
	// the config.Config interface.

	// one such object is the configuration object for
	// credentials.

	creds := credcfg.New()

	// here we can configure credential settings:
	// credential repositories and consumer is mappings.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	creds.AddConsumer(
		id,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)

	// credential objects are typically serializable and deserializable.

	spec, err := json.MarshalIndent(creds, "  ", "  ")
	if err != nil {
		return errors.Wrapf(err, "marshal credential config")
	}

	fmt.Printf("this a a credential configuration object:\n%s\n", string(spec))

	// like all the other maifest based description this format always includes
	// a type field, which can be used to deserialize a specification into
	// the appropriate object.
	// This can ebe done by the config context. It accepts YAML or JSON.

	o, err := ctx.GetConfigForData(spec, nil)
	if err != nil {
		return errors.Wrapf(err, "deserialize config")
	}

	if diff := deep.Equal(o, creds); len(diff) != 0 {
		fmt.Printf("diff:\n%v\n", diff)
		return fmt.Errorf("invalid des/erialization")
	}

	// regardless what variant is used (direct object or descriptor)
	// the config object can be added to a config context.
	err = ctx.ApplyConfig(creds, "explicit cred setting")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}

	// Every config object implements the
	// ApplyTo(ctx config.Context, target interface{}) error method.
	// It takes an object, which wants to be configured.
	// The config object then decides, whether it provides
	// settings for the given object and calls the appropriate
	// methods on this object (after a type cast).
	//
	// This way the config mechanism reverts the configuration
	// request, it does not actively configure something, instead
	// an object, which wants to be configured calls the config
	// context to apply pending configs.
	// The config context manages a queue of config objects
	// and applys them to an object to be configured.

	// If ask he credential context now for credentials,
	// it asks the config context for pending config objects
	// and apply them.
	// Theregore, we now should the  configured creentials, here.

	credctx := credentials.DefaultContext()

	found, err := credentials.CredentialsForConsumer(credctx, id)
	if err != nil {
		return errors.Wrapf(err, "cannot get credentials")
	}
	// an error is only provided if something went wrong while determining
	// the credentials. Delivering NO credentials is a valid result.
	if found == nil {
		return fmt.Errorf("no credentials found")
	}
	fmt.Printf("consumer id: %s\n", id)
	fmt.Printf("credentials: %s\n", obfuscate(found))

	if found.GetProperty(credentials.ATTR_USERNAME) != cfg.Username {
		return fmt.Errorf("password mismatch")
	}
	if found.GetProperty(credentials.ATTR_PASSWORD) != cfg.Password {
		return fmt.Errorf("password mismatch")
	}
	return nil
}
