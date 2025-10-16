package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	configcfg "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/credentials"
	credcfg "ocm.software/ocm/api/credentials/config"
	"ocm.software/ocm/api/credentials/extensions/repositories/dockerconfig"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/examples/lib/helper"
	"sigs.k8s.io/yaml"
)

func HandleOCMConfig(cfg *helper.Config) error {
	// Although the configuration of an OCM context can
	// be done by a sequence of explicit calls the mechanism
	// shown in the example before, is used to provide a simple
	// library function, which can be used to configure an OCM
	// context and all related other contexts with a single call
	// based on a central configuration file (~/.ocmconfig)

	// --- begin central config ---
	ctx := ocm.DefaultContext()
	_, err := utils.Configure(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "configuration")
	}
	// --- end central config ---

	// This file typically contains the serialization of such a generic
	// configuration specification (or any other serialized configuration object),
	// enriched with specialized config specifications for
	// credentials, default repositories, signing keys and any
	// other configuration specification.
	//
	// Most important are here the credentials.
	// Because OCM embraces lots of storage technologies for artifact
	// storage as well as storing OCM component version metadata,
	// there are typically multiple technology specific ways
	// to configure credentials for command line tools.
	// Using the credentials settings shown in the previous tour,
	// it is possible to specify credentials for all
	// required purposes, and the configuration management provides
	// an extensible way to embed native technology specific ways
	// to provide credentials just by adding an appropriate type
	// of credential repository, which reads the specialized storage and
	// feeds it into the credential context. Those specifications
	// can be added via the credential configuration object to
	// the central configuration.
	//
	// One such repository type is the Docker config type. It
	// reads a `dockerconfig.json` file and feeds in the credentials.
	// Because it is used for a dedicated purpose (credentials for
	// OCI registries), it not only can feed the credentials, but
	// also their mapping to consumer ids.

	// We first create the specification for a new credential repository of
	// type `dockerconfig` describing the default location
	// of the standard Docker config file.

	// --- begin docker config ---
	credspec := dockerconfig.NewRepositorySpec("~/.docker/config.json", true)

	// add this repository specification to a credential configuration.
	ccfg := credcfg.New()
	err = ccfg.AddRepository(credspec)
	if err != nil {
		return errors.Wrapf(err, "invalid credential config")
	}
	// --- end docker config ---

	// By adding the default location for the standard Docker config
	// file, all credentials provided by the `docker login` command
	// are available in the OCM toolset, also.

	// A typical minimal <code>.ocmconfig</code> file can be composed as follows.
	// We add this config object to an empty generic configuration object
	// and print the serialized form. The result can be used as
	// default initial OCM configuration file.

	// --- begin default config ---
	ocmcfg := configcfg.New()
	err = ocmcfg.AddConfig(ccfg)

	spec, err := yaml.Marshal(ocmcfg)
	if err != nil {
		return errors.Wrapf(err, "marshal ocm config")
	}

	// the result is a typical minimal ocm configuration file
	// just providing the credentials configured with
	// <code>doicker login</code>.
	fmt.Printf("this a typical ocm config file:\n--- begin ocmconfig ---\n%s--- end ocmconfig ---\n", string(spec))
	// --- end default config ---

	// Besides from a file, such a config can be provided as data, also,
	// taken from any other source, for example from a Kubernetes secret

	// --- begin by data ---
	err = utils.ConfigureByData(ctx, spec, "from data")
	if err != nil {
		return errors.Wrapf(err, "configuration")
	}
	// --- end by data ---

	// If you have provided your OCI credentials with
	// `docker login`, they should now be available.

	// --- begin query ---
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	found, err := credentials.CredentialsForConsumer(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "cannot get credentials")
	}
	fmt.Printf("consumer id: %s\n", id)
	fmt.Printf("credentials: %s\n", obfuscate(found))
	// --- end query ---

	// the configuration library function does not only read the
	// ocm config file, it also applies [*spiff*](github.com/mandelsoft/spiff)
	// processing to the provided YAML/JSON content. *Spiff* is an
	// in-domain yaml-based templating engine. Therefore, you can use
	// any spiff dynaml expression to define values or even complete
	// sub structures.

	// --- begin spiff ---
	ocmcfg = configcfg.New()
	ccfg = credcfg.New()
	cspec := credentials.CredentialsSpecFromList("clientCert", `(( read("~/ocm/keys/myClientCert.pem") ))`)
	id = credentials.NewConsumerIdentity("ApplicationServer.acme.org", "hostname", "app.acme.org")
	ccfg.AddConsumer(id, cspec)
	ocmcfg.AddConfig(ccfg)
	// --- end spiff ---

	spec, err = yaml.Marshal(ocmcfg)
	if err != nil {
		return errors.Wrapf(err, "marshal ocm config")
	}
	fmt.Printf("this a typical ocm config file using spiff file operations:\n--- begin spiffocmconfig ---\n%s--- end spiffocmconfig ---\n", string(spec))

	// this config object is not directly usable, because the cert value is not
	// a valid certificate. We use it here just to generate the serialized form.
	// if this is used with the above library functions, the finally generated
	// config object will contain the read file content, which is hopefully a
	// valid certificate.

	return nil
}
