# Signing Component Versions

{{signing}}

This tour illustrates the basic functionality to
sign and verify signatures.

It covers two basic scenarios:

- [`sign`](/examples/lib/tour/06-signing-component-versions/01-basic-signing.go) Create, Sign, Transport and Verify a component version.
- [`context`](/examples/lib/tour/06-signing-component-versions/02-using-context-settings.go) Using context settings to configure signing and verification in target repo.

## Running the examples

You can call the main program with a config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
targetRepository:
  type: CommonTransportFormat
  filePath: /tmp/example06.target.ctf
  fileFormat: directory
  accessMode: 2
ocmConfig: <your ocm config file>
```

The actual version of the example just works with the file system
target, because it is not possible to specify credentials for the
target repository in this simple config file. But, if you specific an [OCM config file](../04-working-with-config/README.md) you can
add more credential settings to make target repositories possible
requiring credentials.

## Walkthrough

### Create, Sign, Transport and Verify a component version

As usual, we start with getting access to an OCM context

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{default context}}
```

Then, we configure this context with optional ocm config defined in our config file.
See [OCM config scenario in tour 04]({{ocm-config-file}}).

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{configure}}
```

To sign a component version we need a private key.
For this example, we just create a local keypair.
To be able to verify later, we should save the public key,
but here we do all this in a single program.

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{create keypair}}
```

{{tour06-compose}}
And we need a component version to sign.
We again compose a component version without a repository
(see [tour02 example 2]({{composition-environment}})).

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{compose}}
```

Now, let's sign the component version.
There might be multiple signatures, therefore every signature
has a name (here `acme.org`). Keys are always specified for
a dedicated signature name. The signing process can be influenced by
several options. Here, we just provide the private key to be used in an ad-hoc manner.
[Later]({{signing-context}}), we will see how everything can be preconfigured in a *signing context*.

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{sign}}
```

Now, we add the signed component version to a target repository.
Here, we just reuse the code from [tour02]({{composition-environment}})

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{add version}}
```

Let's check the target for the new component version.

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{lookup}}
```

Please note, that the version now contains a signature.

Finally, we check whether the signature is still valid for the
target version.

```go
{{include}{../../06-signing-component-versions/01-basic-signing.go}{verify}}
```

{{signing-context}}

### Using Context Settings to Configure Signing

Instead of providing all signing relevant information directly with
the signing or verification calls, it is possible to preconfigure
various information at the OCM context.

As usual, we start with getting access to an OCM context

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{default context}}
```

Then, we configure this context with optional ocm config defined in our config file.
See [OCM config scenario in tour 04]({{ocm-config-file}}).

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{configure}}
```

To sign a component version we need a private key.
For this example, we again just create a local keypair.
To be able to verify later, we should save the public key,
but here we do all this in a single program.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{create keypair}}
```

Finally, we create a component version in our target repository. The called
function

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{setup}}
```

executes the same coding already shown in the [previous]({{tour06-compose}}) example.

#### Signing Using Manual Context Settings

After this preparation we now configure the signing part of the OCM context.
Every OCM context features a signing registry, which provides available
signers and hashers, but also keys and certificates for various purposes.
It is always asked if a key is required, which is
not explicitly given to a signing/verification call.

This context part is implemented as additional attribute stored along
with the context. Attributes are always implemented as a separate package
containing the attribute structure, its deserialization and
a `Get(Context)` function to retrieve the attribute for the context.
This way new arbitrary attributes for various use cases can be added
without the need to change the context interface.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{signing attribute}}
```

Now, we manually add the keys to our context.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{configure keys}}
```

We are prepared now and can sign any component version without specifying further options
in any repository for the signature name `acme.org`.

Therefore, we just get the component version from the prepared repository

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{lookup component version}}
```

and finally sign it. We don't need to present the key, here. It is taken from the
context.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{sign}}
```

The same way we can just call `VerifyComponentVersion` to
verify the signature.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{verify}}
```

#### Configuring Keys with OCM Configuration File

Manually adding keys to the signing attribute
might simplify the call to possibly multiple signing/verification
calls, but it does not help to provide keys via an external
configuration (for example for using the OCM CLI).
In [tour04]({{tour04-arbitrary}})
we have seen how arbitrary configuration
possibilities can be added. The signing attribute uses
this mechanism to configure itself by providing an own
configuration object, which can be used to feed keys (and certificates)
into the signing attribute of an OCM context.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{create signing config}}
```

It provides methods to add elements
like keys and certificates, which convert
these elements into a (de-)serializable form.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{add signing config}}
```

By adding this config to a generic configuration object you get
an OCM config usable to predefine keys for your CLI.

```go
{{include}{../../06-signing-component-versions/02-using-context-settings.go}{print signing config}}
```

And here is a sample output containing the public and private key.

```yaml
{{execute}{go}{run}{../../06-signing-component-versions}{--config}{settings.yaml}{config}{<extract>}{ocmconfig}}
```
