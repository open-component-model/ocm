package main

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/config/cpi"
	configcfg "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/examples/lib/helper"
	"sigs.k8s.io/yaml"
)

// TYPE is the name of our new configuration object type.
// To be globally unique, it should always end with a
// DNS domain owned by the provider of the new type.
// --- begin type name ---
const TYPE = "example.config.acme.org"

// --- end type name ---

// ExampleConfigSpec is the new Go type for the config specification
// covering our example configuration.
// It just encapsulates our simple configuration structure
// used to configure the examples of our tour.
// --- begin config type ---
type ExampleConfigSpec struct {
	// ObjectVersionedType is the base type providing the type feature
	// for (config) specifications.
	runtime.ObjectVersionedType `json:",inline"`
	// Config is our example config representation.
	helper.Config `json:",inline"`
}

// --- end config type ---

// NewConfig provides a config object for out helper configuration.
// --- begin constructor ---
func NewConfig(cfg *helper.Config) cpi.Config {
	return &ExampleConfigSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(TYPE),
		Config:              *cfg,
	}
}

// --- end constructor ---

// additional setters can be used to configure the configuration object.
// Here, programmatic objects (like an ocm.RepositorySpec) are
// converted to a form storable in the configuration object.
// --- begin setters ---

// SetTargetRepository takes a repository specification
// and adds its serialized form to the config object.
func (c *ExampleConfigSpec) SetTargetRepository(target ocm.RepositorySpec) error {
	data, err := json.Marshal(target)
	if err != nil {
		return err
	}
	c.Target = data
	return nil
}

// SetTargetRepositoryData sets the target repository specification
// from a byte sequence.
func (c *ExampleConfigSpec) SetTargetRepositoryData(data []byte) error {
	err := runtime.CheckSpecification(data)
	if err != nil {
		return err
	}
	c.Target = data
	return nil
}

// --- end setters ---

// --- begin config interface ---

// RepositoryTarget consumes a repository name.
type RepositoryTarget interface {
	SetRepository(r string)
}

// --- end config interface ---

// ApplyTo is used to apply the provided configuration settings
// to a dedicated object, which wants to be configured.
// --- begin method apply ---.
func (c *ExampleConfigSpec) ApplyTo(_ cpi.Context, tgt interface{}) error {
	switch t := tgt.(type) {
	// if the target is a credentials context
	// configure the credentials to be used for the
	// described OCI repository.
	case credentials.Context:
		// determine the consumer id for our target repository-
		id, err := oci.GetConsumerIdForRef(c.Repository)
		if err != nil {
			return errors.Wrapf(err, "invalid consumer")
		}
		// create the credentials.
		creds := c.GetCredentials()

		// configure the targeted credential context with
		// the provided credentials (see previous examples).
		t.SetCredentialsForConsumer(id, creds)

	// if the target consumes an OCI repository, propagate
	// the provided OCI repository ref.
	case RepositoryTarget:
		t.SetRepository(c.Repository)

	// all other targets are ignored, we don't have
	// something to set at these objects.
	default:
		return cpi.ErrNoContext(TYPE)
	}
	return nil
}

// --- end method apply ---

// to enable automatic deserialization of our new config type,
// we have to tell the configuration management about our
// new type. This is done by a registration function,
// which gets called with a dedicated type object for
// the new config type.
// a type object describes the config type, its type name, how
// it is serialized and deserialized and some description.
// we use a standard type object, here, instead of implementing
// an own one. It is parameterized by the Go pointer type for
// our specification object.

// --- begin init ---.
func init() {
	// register the new config type, so that is can be used
	// by the config management to deserialize appropriately
	// typed specifications.
	cpi.RegisterConfigType(cpi.NewConfigType[*ExampleConfigSpec](TYPE, "this ia config object type based on the example config data."))
}

// --- end init ---.

func WriteConfigType(cfg *helper.Config) error {
	// after preparing a new special config type
	// we can feed it into the config management.
	// because of the registration the config management
	// now knows about this new type.

	// A usual, we gain access to our required
	// contexts.
	// --- begin default context ---
	credctx := credentials.DefaultContext()

	// the credential context is based on a config context
	// used to configure it.
	ctx := credctx.ConfigContext()
	// --- end default context ---

	// to setup our environment we create our new config based on the actual
	// settings and apply it to the config context.
	// --- begin apply ---
	examplecfg := NewConfig(cfg)
	ctx.ApplyConfig(examplecfg, "special acme config")
	// --- end apply ---

	// If you omit the above call, no credentials
	// will be found later.

	// now we should be prepared to get the credentials
	// --- begin query credentials ---
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "cannot get consumer id")
	}
	fmt.Printf("usage context: %s\n", id)

	// the returned credentials are provided via an interface, which might change its
	// content, if the underlying credential source changes.
	creds, err := credentials.CredentialsForConsumer(credctx, id, ociidentity.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "credentials")
	}
	fmt.Printf("credentials: %s\n", obfuscate(creds))
	// --- end query credentials ---

	// Because of the new credential type, such a specification can
	// now be added to the ocm config, also.
	// So, we could use our special tour config file content
	// directly as part of the ocm config.

	// --- begin in ocmconfig ---
	ocmcfg := configcfg.New()
	err = ocmcfg.AddConfig(examplecfg)

	spec, err := yaml.Marshal(ocmcfg)
	if err != nil {
		return errors.Wrapf(err, "marshal ocm config")
	}

	// the result is a minimal ocm configuration file
	// just providing our new example configuration.
	fmt.Printf("this a typical ocm config file:\n--- begin ocmconfig ---\n%s--- end ocmconfig ---\n", string(spec))
	// --- end in ocmconfig ---

	// above, we added a new kind of target, the RepositoryTarget interface.
	// Just by providing an implementation for this interface, we can
	// configure such an object using the config management.

	// --- begin apply interface ---
	target := &SimpleRepositoryTarget{}

	_, err = ctx.ApplyTo(0, target)
	if err != nil {
		return errors.Wrapf(err, "applying to new target")
	}
	fmt.Printf("repository for target: %s\n", target.repository)
	// --- end apply interface ---

	// This way any specialized configuration object can be added
	// by a user of the OCM library. It can be used to configure
	// existing objects or even new object types, even in combination.
	//
	// What is still required is a way
	// to implement new config targets, objects, which want
	// to be configured and which autoconfigure themselves when
	// used. Our simple repository target is just an example
	// for some kind of ad-hoc configuration.
	// a complete scenario is shown in the next example.
	return nil
}

// --- begin demo target ---

// SimpleRepositoryTarget is demo target object
// just implementing our new configuration interface.
type SimpleRepositoryTarget struct {
	repository string
}

var _ RepositoryTarget = (*SimpleRepositoryTarget)(nil)

func (t *SimpleRepositoryTarget) SetRepository(repo string) {
	t.repository = repo
}

// --- end demo target ---
