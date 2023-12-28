# Working with Configurations

This tour illustrates the basic configuration management
included in the OCM library. The library provides
an extensible framework to bring together configuration settings
and configurable objects.

It covers five basic scenarios:
- [`basic`](01-basic-config-management.go) Basic configuration management illustrating the configuration of credentials.
- [`generic`](02-handle-arbitrary-config.go) Handling of arbitrary configuration.
- [`ocm`](03-using-ocm-config.go) Central configuration
- [`provide`](04-write-config-type.go) Providing new config object types
- [`consume`](05-write-config-consumer.go) Preparing objects to be configured by the config management

## Running the example

You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
```

Set your favorite OCI registry and don't forget to add the repository prefix for your OCM repository hosted in this registry.

## Walkthrough

### Basic Configuration Management

Similar to the other context areas, Configuration is handled by the configuration contexts.
Therefore, for the example, we just get the default configuration context.

```go
	ctx := config.DefaultContext()
```

The configuration context handles configuration objects.
A configuration object is any object implementing
the `config.Config` interface. The task of a config object
is to apply configuration to some target object.

One such object is the configuration object for
credentials provided by the credentials context.
It finally applies settings to a credential context.

```go
	creds := credcfg.New()
```

Here, we can configure credential settings:
credential repositories and consumer id mappings.
We do this by setting the credentials provided
by our config file for the consumer id used
by our configured OCI registry.

```go
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	creds.AddConsumer(
		id,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)
```

(Credential) Configuration objects are typically serializable and deserializable.

```go
	spec, err := json.MarshalIndent(creds, "  ", "  ")
	if err != nil {
		return errors.Wrapf(err, "marshal credential config")
	}

	fmt.Printf("this a a credential configuration object:\n%s\n", string(spec))
```

Like all the other manifest based descriptions this format always includes
a type field, which can be used to deserialize a specification into
the appropriate object.
This can be done by the config context. It accepts YAML or JSON.

```go
	o, err := ctx.GetConfigForData(spec, nil)
	if err != nil {
		return errors.Wrapf(err, "deserialize config")
	}

	if diff := deep.Equal(o, creds); len(diff) != 0 {
		fmt.Printf("diff:\n%v\n", diff)
		return fmt.Errorf("invalid des/erialization")
	}
```

Regardless what variant is used (direct specification object or descriptor)
the config object can be added to a config context.

```go
	err = ctx.ApplyConfig(creds, "explicit cred setting")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}
```

Every config object implements the
`ApplyTo(ctx config.Context, target interface{}) error` method.
It takes an object, which wants to be configured.
The config object then decides, whether it provides
settings for the given object and calls the appropriate
methods on this object (after a type cast).

Here is the code snippet from the apply method of the credential
config object ([.../pkg/contexts/credentials/config/type.go](../../../../../pkg/contexts/credentials/config/type.go)):

```go

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	list := errors.ErrListf("applying config")
	t, ok := target.(cpi.Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}
	for _, e := range a.Consumers {
		t.SetCredentialsForConsumer(e.Identity, CredentialsChain(e.Credentials...))
	}
        ...
```

This way the config mechanism reverts the configuration
request, it does not actively configure something, instead
an object, which wants to be configured calls the config
context to apply pending configs.
To do this the config context manages a queue of config objects
and applies them to an object to be configured.

If the credential context is asked now for credentials,
it asks the config context for pending config objects
and applies them.
Therefore, we now should get the configured credentials, here.

```go
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
```

### Handling of Arbitrary Configuration

The config management not only manages configuration objects for any
other configurable object, it also provides a configuration object of
its own. The task of the object is to handle other configuration objects
to be applied to a configuration object.

```go
	generic := configcfg.New()
```

The generic config object holds a list of any other config objects,
or their specification formats.
Additionally, it is possible to configure named sets
of configurations, which can later be enabled
on-demand by their name at the config context.

We recycle our credential config from the last example to get
a config object to be added to our generic config object.

```go
	creds, err := credConfig(cfg)
	if err != nil {
		return err
	}
```

Now, we can add this credential config object to
our generic config list.

```go
	err = generic.AddConfig(creds)
	if err != nil {
		return errors.Wrapf(err, "adding config")
	}
```

As we have seen in our previous example config objects are typically
serializable and deserializable. This also holds for the generic config
object of the config context.

```go
	spec, err := json.MarshalIndent(generic, "  ", "  ")
	if err != nil {
		return errors.Wrapf(err, "marshal credential config")
	}

	fmt.Printf("this a a generic configuration object:\n%s\n", string(spec))
```

The result is a config object hosting a list (with 1 entry)
of other config object specifications.

The generic config object can be added to a config context, again, like
any other config object. If it is asked to configure a configuration
context it uses the methods of the configuration context to apply the
contained list of config objects (and the named set of config lists).
Therefore, all config objects applied to a configuration context are
asked to configure the configuration context itself when queued to the
list of applied configuration objects.

If we now ask the default credential context (which uses the default
configuration context to configure itself) for credentials for our OCI registry,
the credential mapping provided by the config object added to the generic one,
will be found.

```go
	ctx := config.DefaultContext()
	err = ctx.ApplyConfig(creds, "generic setting")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}
	credctx := credentials.DefaultContext()

	// query now works, also.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	found, err := credentials.CredentialsForConsumer(credctx, id)
	if err != nil {
		return errors.Wrapf(err, "cannot get credentials")
	}
	fmt.Printf("consumer id: %s\n", id)
	fmt.Printf("credentials: %s\n", obfuscate(found))
```

The very same mechanism is used to provide central configuration in a
configuration file for the OCM ecosystem, as will be shown in the next example.

### Central Configuration

Although the configuration of an OCM context can
be done by a sequence of explicit calls according to the mechanisms
shown in the examples before, a simple convenience 
library function is provided, which can be used to configure an OCM
context and all related other contexts with a single call
based on a central configuration file (`~/.ocmconfig`)

```go
	ctx := ocm.DefaultContext()
	_, err := utils.Configure(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "configuration")
	}
```

This file typically contains the serialization of such a generic
configuration specification (or any other serialized configuration object),
enriched with specialized config specifications for
credentials, default repositories, signing keys and any
other configuration specification.

#### Standard Configuration File

Most important are here the credentials.
Because OCM embraces lots of storage technologies for artifact
storage as well as storing OCM component version metadata,
there are typically multiple technology specific ways
to configure credentials for command line tools.
Using the credentials settings shown in the previous tour,
it is possible to specify credentials for all
required purposes, and the configuration management provides
an extensible way to embed native technology specific ways
to provide credentials just by adding an appropriate type
of credential repository, which reads the specialized stoarge and
feeds it into the credential context. Those specifications
can be added via the credengtial configuration object to
the central configuration.

One such repository type is the docker config type. It
reads a `dockerconfig.json` file and feeds in the credentials.
Because it is used for a dedicated purpose (credentials for 
OCI registries), it not only can feed the credentials, but
also their mapping to consumer ids.

We first create the specification for a new credential repository of
type `dockerconfig` describing the default location
of the standard docker config file.

```go
	credspec := dockerconfig.NewRepositorySpec("~/.docker/config.json", true)

	// add this repository specification to a credential configuration.
	ccfg := credcfg.New()
	err = ccfg.AddRepository(credspec)
	if err != nil {
		return errors.Wrapf(err, "invalid credential config")
	}
```

By adding the default location for the standard docker config
file, all credentials provided by the <code>docker login</code>
are available in the OCM toolset, also.

A typical minimal <code>.ocmconfig</code> file can be composed as follows.
We add this config object to an empty generic configuration object
and print the serialized form. The result can be used as
default initial OCM configuration file.

```go
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
```

The result should look similar to (but with reorderd fields):
```yaml
type: generic.config.ocm.software
configurations:
  - type: credentials.config.ocm.software
    repositories:
      - repository:
          type: DockerConfig
          dockerConfigFile: ~/.docker/config.json
          propagateConsumerIdentity: true
```

Because of the ordered map keys the actual output looks a little bit confusing:

```yaml
configurations:
- repositories:
  - repository:
      dockerConfigFile: ~/.docker/config.json
      propagateConsumerIdentity: true
      type: DockerConfig
  type: credentials.config.ocm.software
type: generic.config.ocm.software
```

Besides from a file, such a config can be provided as data, also,
taken from any other source, for example from a Kubernetes secret.

```go
	err = utils.ConfigureByData(ctx, spec, "from data")
	if err != nil {
		return errors.Wrapf(err, "configuration")
	}
```

If you have provided your OCI credentials with
docker login, they should now be available.

```go
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
```

#### Templating

The configuration library function not only reads the
ocm config file, it applies [*spiff*](github.com/mandelsoft/spiff)
processing to the provided YAML/JSON content. *Spiff* is an
in-domain yaml-based templating engine. Therefore, you can use
any spiff dynaml expression to define values or even complete
sub structures.

```go
	ocmcfg = configcfg.New()
	ccfg = credcfg.New()
	cspec := credentials.CredentialsSpecFromList("clientCert", `(( read("~/ocm/keys/myClientCert.pem") ))`)
	id = credentials.NewConsumerIdentity("ApplicationServer.acme.org", "hostname", "app.acme.org")
	ccfg.AddConsumer(id, cspec)
	ocmcfg.AddConfig(ccfg)
```

This config object is not directly usable, because the cert value is not
a valid certificate. We use it here just to generate the serialized form.

```yaml
configurations:
- consumers:
  - credentials:
    - credentialsName: Credentials
      properties:
        clientCert: (( read("~/ocm/keys/myClientCert.pem") ))
      type: Credentials
    identity:
      hostname: app.acme.org
      type: ApplicationServer.acme.org
  type: credentials.config.ocm.software
type: generic.config.ocm.software
```

If this is used with the above library functions, the finally generated
config object will contain the read file content, which is hopefully a
valid certificate.
