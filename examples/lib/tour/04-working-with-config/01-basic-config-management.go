package main

import (
	"encoding/json"
	"fmt"

	"github.com/go-test/deep"
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	credcfg "ocm.software/ocm/api/credentials/config"
	"ocm.software/ocm/api/credentials/extensions/repositories/directcreds"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/examples/lib/helper"
)

func BasicConfigurationHandling(cfg *helper.Config) error {
	// configuration is handled by the configuration context.
	// --- begin default context ---
	ctx := config.DefaultContext()
	// --- end default context ---

	// the configuration context handles configuration objects.
	// a configuration object is any object implementing
	// the config.Config interface. The task of a config object
	// is to apply configuration to some target object.

	// one such object is the configuration object for
	// credentials.
	// It finally applies settings to a credential context.

	// --- begin cred config ---
	creds := credcfg.New()
	// --- end cred config ---

	// here, we can configure credential settings:
	// credential repositories and consumer id mappings.
	// We do this by setting the credentials provided
	// by our config file for the consumer id used
	// by our configured OCI registry.
	// --- begin configure creds ---
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	creds.AddConsumer(
		id,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)
	// --- end configure creds ---

	// configuration objects are typically serializable and deserializable.

	// --- begin marshal ---
	spec, err := json.MarshalIndent(creds, "  ", "  ")
	if err != nil {
		return errors.Wrapf(err, "marshal credential config")
	}

	fmt.Printf("this a a credential configuration object:\n%s\n", string(spec))
	// --- end marshal ---

	// like all the other manifest based descriptions this format always includes
	// a type field, which can be used to deserialize a specification into
	// the appropriate object.
	// This can be done by the config context. It accepts YAML or JSON.

	// --- begin unmarshal ---
	o, err := ctx.GetConfigForData(spec, nil)
	if err != nil {
		return errors.Wrapf(err, "deserialize config")
	}

	if diff := deep.Equal(o, creds); len(diff) != 0 {
		fmt.Printf("diff:\n%v\n", diff)
		return fmt.Errorf("invalid des/erialization")
	}
	// --- end unmarshal ---

	// regardless what variant is used (direct object or descriptor)
	// the config object can be added to a config context.
	// --- begin apply config ---
	err = ctx.ApplyConfig(creds, "explicit cred setting")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}
	// --- end apply config ---

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
	// and applies them to an object to be configured.

	// If the credential context is asked now for credentials,
	// it asks the config context for pending config objects
	// and applies them.
	// Therefore, we now should be able to get the configured credentials.

	// --- begin get credentials ---
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
	// --- end get credentials ---
	return nil
}
