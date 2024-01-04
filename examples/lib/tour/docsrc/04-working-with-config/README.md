{{config}}
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

You can call the main program with a config file option (`--config <file>`) and the name of the scenario.
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
{{include}{../../04-working-with-config/01-basic-config-management.go}{default context}}
```

The configuration context handles configuration objects.
A configuration object is any object implementing
the `config.Config` interface. The task of a config object
is to apply configuration to some target object.

One such object is the configuration object for
credentials provided by the credentials context.
It finally applies settings to a credential context.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{cred config}}
```

Here, we can configure credential settings:
credential repositories and consumer id mappings.
We do this by setting the credentials provided
by our config file for the consumer id used
by our configured OCI registry.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{configure creds}}
```

(Credential) Configuration objects are typically serializable and deserializable.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{marshal}}
```

Like all the other manifest based descriptions this format always includes
a type field, which can be used to deserialize a specification into
the appropriate object.
This can be done by the config context. It accepts YAML or JSON.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{unmarshal}}
```

Regardless what variant is used (direct specification object or descriptor)
the config object can be added to a config context.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{apply config}}
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
{{include}{../../../../../pkg/contexts/credentials/config/type.go}{apply}}
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
Therefore, we now should be able to get the configured credentials.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{get credentials}}
```

### Handling of Arbitrary Configuration

The config management not only manages configuration objects for any
other configurable object, it also provides a configuration object of
its own. The task of the object is to handle other configuration objects
to be applied to a configuration object.

```go
{{include}{../../04-working-with-config/02-handle-arbitrary-config.go}{config config}}
```

The generic config object holds a list of any other config objects,
or their specification formats.
Additionally, it is possible to configure named sets
of configurations, which can later be enabled
on-demand by their name at the config context.

We recycle our credential config from the last example to get
a config object to be added to our generic config object.

```go
{{include}{../../04-working-with-config/02-handle-arbitrary-config.go}{sub config}}
```

Now, we can add this credential config object to
our generic config list.

```go
{{include}{../../04-working-with-config/02-handle-arbitrary-config.go}{add config}}
```

As we have seen in our previous example, config objects are typically
serializable and deserializable. This also holds for the generic config
object of the config context.

```go
{{include}{../../04-working-with-config/02-handle-arbitrary-config.go}{serialized}}
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
{{include}{../../04-working-with-config/02-handle-arbitrary-config.go}{query}}
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
{{include}{../../04-working-with-config/03-using-ocm-config.go}{central config}}
```

This file typically contains the serialization of such a generic
configuration specification (or any other serialized configuration object),
enriched with specialized config specifications for
credentials, default repositories, signing keys and any
other configuration specification.

{{ocm-config-file}}
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
of credential repository, which reads the specialized storage and
feeds it into the credential context. Those specifications
can be added via the credential configuration object to
the central configuration.

One such repository type is the Docker config type. It
reads a `dockerconfig.json` file and feeds in the credentials.
Because it is used for a dedicated purpose (credentials for 
OCI registries), it not only can feed the credentials, but
also their mapping to consumer ids.

We first create the specification for a new credential repository of
type `dockerconfig` describing the default location
of the standard Docker config file.

```go
{{include}{../../04-working-with-config/03-using-ocm-config.go}{docker config}}
```

By adding the default location for the standard Docker config
file, all credentials provided by the `docker login` command
are available in the OCM toolset, also.

A typical minimal <code>.ocmconfig</code> file can be composed as follows.
We add this config object to an empty generic configuration object
and print the serialized form. The result can be used as
default initial OCM configuration file.

```go
{{include}{../../04-working-with-config/03-using-ocm-config.go}{default config}}
```

The result should look similar to (but with reordered fields):
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
{{execute}{go}{run}{../../04-working-with-config}{--config}{settings.yaml}{ocm}{<extract>}{ocmconfig}}
```

Besides from a file, such a config can be provided as data, also,
taken from any other source, for example from a Kubernetes secret.

```go
{{include}{../../04-working-with-config/03-using-ocm-config.go}{by data}}
```

If you have provided your OCI credentials with
`docker login`, they should now be available.

```go
{{include}{../../04-working-with-config/03-using-ocm-config.go}{query}}
```

#### Templating

The configuration library function does not only read the
ocm config file, it also applies [*spiff*](github.com/mandelsoft/spiff)
processing to the provided YAML/JSON content. *Spiff* is an
in-domain yaml-based templating engine. Therefore, you can use
any spiff dynaml expression to define values or even complete
sub structures.

```go
{{include}{../../04-working-with-config/03-using-ocm-config.go}{spiff}}
```

This config object is not directly usable, because the cert value is not
a valid certificate. We use it here just to generate the serialized form.

```yaml
{{execute}{go}{run}{../../04-working-with-config}{--config}{settings.yaml}{ocm}{<extract>}{spiffocmconfig}}
```

If this is used with the above library functions, the finally generated
config object will contain the read file content, which is hopefully a
valid certificate.

{{tour04-arbitrary}}
### Providing new config object types

So far, we just used existing config types to configure existing objects.
But the configuration management is highly extensible, and it is quite
simple to provide new config types, which can be used to configure
any new or existing object, which is prepared to consume configuration.

The next [chapter]({{consume-config}}) will show how to prepare an
object to be automatically configurable by
the configuration management. Here, we focus on the implementation of
new config object types. Therefore, we want to configure the
credential context by a new configuration object.

#### The Configuration Object Type

Typically, every kind of configuration object lives in its own package,
which always have the same layout.

A configuration object has a *type*, the configuration type. Therefore,
the package declares a constant `TYPE`.

It is the name of our new configuration object type.
To be globally unique, it should always end with a
DNS domain owned by the provider of the new type.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{type name}}
```

Next, we need a Go type. `ExampleConfigSpec` is the new Go type for the
config specification covering our example configuration.
It just encapsulates our simple configuration structure
used to configure the examples of our tour.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{config type}}
```

Every config type structure must contain a field (and the appropriate methods)
for storing the config type name. This is done by embedding the
type `runtime.ObjectVersionedType` from the `runtime` package. This package
contains everything to work with specification objects and
serialization/deserialization.

As second field we just embed the config structure used to read the tour
config. This way any kind of configuration information can be mapped
to the configuration management.

A config type typically provide a constructor for a config object of
this type:

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{constructor}}
```

Additional setters can be used to configure the configuration object.
Here, programmatic objects (like an `ocm.RepositorySpec`) are
converted to a form storable in the configuration object.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{setters}}
```

The utility function `runtime.CheckSpecification` can be used to 
check a byte sequence to be a valid specification.
It just checks for a valid YAML document featuring a non-empty
`type` field:

```go
{{include}{../../../../../pkg/runtime/utils.go}{check}}
```

The most important method to implement is `ApplyTo(_ cpi.Context, tgt interface{}) error`,
which must be implemented by all configuration objects.
Its task is to apply the described configuration settings to a dedicated
object.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{method apply}}
```

Therefore, it decides, whether it is able to handle a dedicated type of target
object and how to configure it. This way a configuration object
may apply is settings or even parts of its setting to any kind of target object.

Our configuration object supports two kinds of target objects:
if the target is a credentials context
it configures the credentials to be used for the
described OCI repository similar to our [credential management example]({{using-cred-management}}).

But we want to accept more types of target objects. Therefore, we 
introduce an own interface declaring the methods required for applying
some configuration settings.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{config interface}}
```

By checking the target object against this interface, we are able 
to configure any kind of object, as long as it provides the necessary
configuration methods.

Now, we are nearly prepared to use our new configuration, there is just one step
missing. To enable the automatic recognition of our new type (for example
in the ocm config file), we have to tell the configuration management
about the new type. This is done by an `init()` function in our config package.

Here, we call a registration function,
which gets called with a dedicated type object for the new config type.
A *type object* describes the config type, its type name, how 
it is serialized and deserialized and some description.
We use a standard type object, here, instead of implementing
an own one. It is parameterized by the Go pointer type (`*ExampleConfigSpec`) for
our specification object.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{init}}
```

#### Using our new Config Object

After preparing a new special config type
we can feed it into the config management.
Because of the registration the config management
now knows about this new type.

A usual, we gain access to our required contexts.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{default context}}
```

To setup our environment we create our new config based on the actual settings 
and apply it to the config context.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{apply}}
```

Now, we should be prepared to get the credentials
the usual way.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{query credentials}}
```

#### Using in the OCM Configuration

Because of the new credential type, such a specification can
now be added to the ocm config, also.
So, we could use our special tour config file content
directly as part of the ocm config.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{in ocmconfig}}
```

The resulting config file looks as follows:

```yaml
{{execute}{go}{run}{../../04-working-with-config}{--config}{settings.yaml}{provide}{<extract>}{ocmconfig}}
```

#### Applying to our Configuration Interface

Above, we added a new kind of target, the `RepositoryTarget` interface.
By providing an implementation for this interface, we can
configure such an object using the config management.
We just provide a simple implementation for this interface, just storing the configured
repository specification.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{demo target}}
```

The context management now is able to apply our config to such an object.

```go
{{include}{../../04-working-with-config/04-write-config-type.go}{apply interface}}
```

This way any specialized configuration object can be added
by a user of the OCM library. It can be used to configure
existing objects or even new object types, even in combination.

What is still required is a way
to implement new config targets, objects, which wants
to be configured and which autoconfigure themselves when
used. Our simple repository target is just an example
for some kind of ad-hoc configuration.
A complete scenario is shown in the next example.

{{consume-config}}
### Preparing Objects to be Configured by the Config Management

We already have our new acme.org config object type,
and a target interface which must be implemented by a target
object to be configurable. The last example showed how
such an object can be configured in an ad-hoc manner
by directly requesting it to be configured by the config
management.

Now, we want to provide an object, which configures
itself when used.
Therefore, we introduce a Go type `RepositoryProvider`,
which should be an object, which is
able to provide an OCI repository reference.
It has a setter and a getter (the setter is
provided by our ad-hoc `SimpleRepositoryTarget`).

To be able to configure itself, the object must know about
the config context it should use to configure itself.

Therefore, our type contains an additional field `updater`.
Its type `cpi.Updater` is a utility provided by the configuration
management, which holds a reference to a configuration context 
and is able to
configure an object based on a managed configuration
watermark. It remembers which config objects from the
config queue are already applied, and replays
the config objects applied to the config context
after the last update.

Finally, a mutex field is contained, which is used to
synchronize updates later.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{type}}
```

For this type a constructor is provided, which initializes
the `updater` field with the desired configuration context.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{constructor}}
```

The magic now happens in the methods provided
by our configurable object.
The first step for methods of configurable objects
dependent on potential configuration is always
to update itself using the embedded updater.

Please note, the config management reverses the
request direction. Applying a config object to
the config context does not configure dependent objects,
it just manages a config queue, which is used by potential
configuration targets to configure themselves.
The actual configuration action is always initiated
by the object, which want to be configured.
The reason for this is to avoid references from the
management to managed objects. This would prohibit
the garbage collection of all configurable objects
as long as the configuration context exists.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{method}}
```

After defining our repository provider type we can now start to use it
together with the configuration management and out configuration object.

As usual, we first determine out context to use.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{default context}}
```

New, we create our provide configurable object by binding it
to the config context.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{object}}
```

If we ask now for a repository we will get the empty 
answer, because nothing is configured, yet.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{initial query}}
```

Now, we apply our config from the last example. Therefore, we create and initialize
the config object with our program settings and apply it to the config
context.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{apply config}}
```

Without any further action, asking for a repository now will return the
configured ref. The configurable object automatically catches the
new configuration from the config context.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{query}}
```

Now, we should also be prepared to get the credentials,
our config object configures the provider as well as
the credential context.

```go
{{include}{../../04-working-with-config/05-write-config-consumer.go}{credentials}}
```