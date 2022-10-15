# Dealing with Configuration

It is possible to explicitly configure all kinds of context by using
the various configuration methods provided by a context.

Nevertheless, it might be useful to externalize the configuration
and delegate it to some *configuration objects*.
For example, the credential settings should be taken from a configuration file.

This is the task of another kind of context, the *config context*.
It manages a sequence of applied configuration objects.
A *configuration object* takes the burden to keep some configuration data
and apply it to an appropriate configuration target.

Such a *configuration target* is typically again a context. For example
a configuration object dealing with credential settings configures
a credential context. But in general, a configuration target may be any
kind of object. To apply a configuration, a configuration object
is called for a desired target. The object may decide to apply itself to this
target or to bypass the target object.

The configuration objects are typed and should be serializable. This
enables the configuration context to use scheme objects, like the other 
context types to reconstruct configurations from a byte stream/textual
representation.

Besides dedicated configuration object types provided by the various context
types, the config context provides a generic configuration type, also.
It is basically a list of other configuration objects, that can be reconstructed
by their deserialization schemes. If applied to a target object, it just
applies the contained configuration in the order of their appearance.

## Using the Configuration Context with explicit Configuration Objects

A first example how to use configuration objects can be found 
[here](config1/example.go).

It just configures a configuration object provided by the credential
context able to configure credential settings.

```go
	cid := credentials.ConsumerIdentity{
		ociid.ID_TYPE:       ociid.CONSUMER_TYPE,
		ociid.ID_HOSTNAME:   "ghcr.io",
		ociid.ID_PATHPREFIX: "mandelsoft",
	}

	// create a credential configuration object
	// and configure it to provide some direct consumer credentials.
	creds := ccfg.New()
	creds.AddConsumer(
		cid,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)
```

It just declares direct credentials for a dedicated consumer id (see the 
[credentials example](creds.md)).

The ocm context can be used to get access to the appropriate configuration
context, which is used to apply the configuration object.

```go
	octx := ocm.DefaultContext()
	cctx := octx.ConfigContext()

	err = cctx.ApplyConfig(creds, "explicit")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}
```

After the object has been applied, the result can be observed on the
intended target object. For a credential configuration this is the
credential context. Here, the credentials for the configured consumer id can
be queried. In the example this is a crednetials object valid for an
OCI registry.

```go
	credctx := octx.CredentialsContext()

	found, err := credctx.GetCredentialsForConsumer(cid, ociid.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "cannot extract credentials")
	}
	got, err := found.Credentials(credctx)
	if err != nil {
		return errors.Wrapf(err, "cannot evaluate credentials")
	}

	fmt.Printf("found: %s\n", got)
```

## Configurations in Configuration Files

The previous example just demonstrates the basic flow, it might not be 
very useful, because the consumer could directly be configured at the 
credential context.

The complete mechanism becomes valuable, if some kind of generic
configuration handling is required. This could be, for example, to
read configurations from external configuration sources, e.g. a central
configuration file.

Instead of providing dedicated mechanism to configure various target environments
the configuration context provides a uniform generic mechanism to
handle arbitrary coniguration settings for any target environment.

The configuration context provides a runtime.Scheme object to
register known configuration types, which offer a deserialization.
This allows storing configuration settings in files. The configuration
context itself provides an aggregative configuration object, which can be used
to host any other configuration object.

As already known by all the scheme based contexts (for example the repository
specifications), the serialized form always features a type field.
A  configuration object for the configuration context could look as follows:

```yaml
type: generic.config.ocm.software/v1
configurations:
  - type: credentials.config.ocm.software
    consumers:
      - identity:
          type: OCIRegistry
          hostname: ghcr.io
          pathprefix: mandelsoft
        credentials:
          - type: Credentials
            properties:
              username: mandelsoft
              password: some-token
    repositories:
      - repository:
          type: DockerConfig/v1
          dockerConfigFile: "~/.docker/config.json"
          propagateConsumerIdentity: true
```

If supports a single data field `configurations`, which is a list
of serialized configuration objects. In the example, here are two entries,
the configuration object from the example above, and a credential repository
specification referring to a docker config file (as used in the credentials
example).

It can be applied as whole as shown in the following code snippet:

```go
	data, err := ioutil.ReadFile(CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot read configuration file %s", CFGFILE)
	}

	octx := ocm.DefaultContext()
	cctx := octx.ConfigContext()

	_, err = cctx.ApplyData(data, runtime.DefaultYAMLEncoding, CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot apply config data")
	}
```

It uses the exactly same configuration mechanism shown in the previous
example, so the query code looks all the same:

```go
	cid := credentials.ConsumerIdentity{
		ociid.ID_TYPE:       ociid.CONSUMER_TYPE,
		ociid.ID_HOSTNAME:   "ghcr.io",
		ociid.ID_PATHPREFIX: "mandelsoft",
	}

	// as before
	credctx := octx.CredentialsContext()

	found, err := credctx.GetCredentialsForConsumer(cid, ociid.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "cannot extract credentials")
	}
	got, err := found.Credentials(credctx)
	if err != nil {
		return errors.Wrapf(err, "cannot evaluate credentials")
	}

	fmt.Printf("found: %s\n", got)
```

The complete example can be found [here](config2/example.go).

## Standard Configuration

The OCM client tool supports reading configuration from a file `~/.ocmconfig`
to configure the used OCM context.
This functionality is offered by a library function, also. The function

    pkg.contexts.ocm.utils.Configure(ctx ocm.Context, path string, fss ...vfs.FileSystem) (ocm.Context, error)

searched for a configuration file and applies it. If not found it looks for
a docker config file and applies an appropriate setting (see example above).

If the config data is already provided by some other means, it can be directly be
applied with the function 

    pkg.contexts.ocm.utils.ConfigureByData(ctx ocm.Context, data []byte, info string) error

Both functions process the YAML content with [spiff](https://github.com/mandelsoft/spiff),
an in-domain templating engine, which allows generating parts of the configuration.

## Configuration Objects

It is very simple to provide own configuration types. It is just
a GO struct implementing the interface `credentials.cpi.Config`. The main method
here is `ApplyTo(configctx Context, target interface{}) error`, which is used to apply
the content to a dedicated target object. The method has to decide on its own
whether it applies to the passed object (type) at all, or what part of its content
is applied. 

This way it is possible to provide configuration objects that configure multiple
types of targets based on the same configuration information.

## Targets

Any object may be used as target. If it is not accepted by any of the specified
configuration objects, they are just ignored.

The typical use-case is to configure contexts. To be able to get up-to-date
with configuration settings applied after an object has been created (with
some initial configuration), the methods of a target object depending on potential
configuration, have to update the target configuration prior to their 
execution. This is supported by an `Update` object, which related to a 
configuration context.

A complete example covering the following two sections can be found
[here](config3/example.go).

### A simple Configuration/Target Pair

A target type provides a connection to a configuration context using
an instance of the `Updater` type. Additionally,  it provides some
data, which is matter to some configuration.

```go
type Target struct {
	updater cpi.Updater
	value   string
}

func NewTarget(ctx cpi.Context) *Target {
	return &Target{
		updater: cpi.NewUpdate(ctx),
	}
}

func (t *Target) SetValue(v string) {
	t.value = v
}

func (t *Target) GetValue() string {
	t.updater.Update(t)
	return t.value
}
```

Whenever a method is called, which depends on potentially configurable
information the `Update`method must be called on the updater instance.
The configuration context keeps track of a sequence of applied configuration 
objects. The updater objects stored the sequence number of the latest executed
update. Calling the `Update` method just replays the configuration objects
applied since the last update.

A configuration object then may look as follows:

```go
const TYPE = "mytype.config.mandelsoft.org"

type Config struct {
	runtime.ObjectVersionedType `json:",inline""`
	Value                       string `json:"value"`
}

func (c *Config) ApplyTo(context cpi.Context, i interface{}) error {
	if i == nil {
		return nil
	}
	t, ok := i.(*Target)
	if !ok {
		return cpi.ErrNoContext(TYPE)
	}
	t.SetValue(c.Value)
	return nil
}

var _ cpi.Config = (*Config)(nil)

func NewConfig(v string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(TYPE),
		Value:               v,
	}
}
```

The type must be uniquely chosen to support the aggregation of configuration
objects in their serialized form.

With this preparing work, an application using the configuration context
to configure this new target object could look like this:

```go
	ctx := config.DefaultContext()

	target := NewTarget(ctx)

	err := ctx.ApplyConfig(NewConfig("hello world"), "explicit1")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config 1")
	}

	fmt.Printf("value is %q\n", target.GetValue())

	err = ctx.ApplyConfig(NewConfig("hello universe"), "explicit2")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config 2")
	}

	fmt.Printf("value is %q\n", target.GetValue())

	newtarget := NewTarget(ctx)
	fmt.Printf("value is %q\n", newtarget.GetValue())
```

Once a connection of the target object to a configuration context is
established, it does not matter, whether the object is created before or
after applying a configuration to the configuration context.
Therefore, a configuration can be applied long before real targets are created.

The configuration context does never refer to potential targets, therefore 
the garbage collection of target objects is not blocked by the existence
of a configuration context for those objects.

### Using External Configuration for the new Configuration Object

If the new configuration object type is registered at the configuration
context, it can even be used together with other configurations provided
by a configuration file as shown in some example above.

For the default scheme used by the default context this can be done
by an `init` function:

```go
func init() {
	cpi.RegisterConfigType(TYPE, cpi.NewConfigType(TYPE, &Config{}, "just provide a value for Target objects"))
}
```

It just creates a type object based on a prototype object and adds some
documentation, which will automatically added to a command line
documentation provided by *cobra*.

Now it is possible use a configuration file

```yaml
type: generic.config.ocm.software/v1
configurations:
  - type: mytype.config.mandelsoft.org
    value: external configuration
```

to configure a dedicated context as shown in the second example:

```go
	data, err := ioutil.ReadFile(CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot read configuration file %s", CFGFILE)
	}
	
	_, err = ctx.ApplyData(data, runtime.DefaultYAMLEncoding, CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot apply config data")
	}

	fmt.Printf("value is %q\n", newtarget.GetValue())
```

When composing a new configuration context with a context builder, it is
possible to use another scheme instance than the default one,
configured by `init` functions.