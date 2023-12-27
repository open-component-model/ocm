{{credentials}}
# Working with Credentials

This tour illustrates the basic handling of credentials
using the OCM library. The library provides
an extensible framework to bring together credential providers
and credential cosunmers in a technology-agnostic way.

It covers four basic scenarios:
- [`basic`](01-using-credentials.go) Writing to a repository with directly specified credentials.
- [`context`](02-basic-credential-management.go) Using credentials via the credential management to publish a component version.
- [`read`](02-basic-credential-management.go) Read the previously created component version using the credential management.
- [`credrepo`](03-credential-repositories.go) Providing credentials via credential repositories.

## Running the example

You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have content similar to:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
```

Set your favorite OCI registry and don't forget to add the repository prefix for your OCM repository hosted in this registry.

## Walkthrough

### Writing to a repository with directly specified credentials.

As usual, we start with getting access to an OCM context
object.

```go
{{include}{../../02-composing-a-component-version/01-basic-componentversion-creation.go}{default context}}
```

So far, we just used memory or filesystem based
OCM repositories to create component versions.
If we want to store something in a remotely accessible
repository typically some credentials are required
for write access.

The OCM library uses a generic abstraction for credentials.
It is just set of properties. To offer various credential sources
There is an interface `credentials.Credentials` provided,
whose implementations provide access to those properties.
A simple property based implementation is `credentials.DirectCredentials.


The most simple use case is to provide the credentials
directly for the repository access creation.
The example config file provides such credentials
for an OCI registry.

```go
{{include}{../../03-working-with-credentials/01-using-credentials.go}{new credentials}}
```

Now, we can use the OCI repository access creation from the [first tour](../01-getting-started/README.md#walkthrough),
but we pass the credentials as additional parameter.
To give you the chance to specify your own registry the URL
is taken from the config file.

```go
{{include}{../../03-working-with-credentials/01-using-credentials.go}{repository access}}
```

If registry name and credentials are fine, we should be able
now to add a new component version to this repository using the coding
from the previous examples, but now we use a public repository, instead
of a memory or filesystem based one. This coding is in function `addVersion`
in `common.go` (It is shared by the other examples, also).

```go
{{include}{../../03-working-with-credentials/common.go}{create version}}
```

In contract to our [first tour](../01-getting-started/README.md#walkthrough)
we cannot list components, here.
OCI registries do not support component listers, therefore we
just lookup the actually added version to verify the result.

```go
{{include}{../../03-working-with-credentials/01-using-credentials.go}{lookup}}
```

The coding for `describeVersion` is similar to the one shown in the [first tour]({{describe-version}}).

### Using the Credential Management

Passing credentials directly at the repository
is fine, as long only the component version
will be accessed. But as soon as described
resource content will be read, the required
credentials and credential types are dependent
on the concrete component version, because
it might contain any kind of access method
referring to any kind of resource repository
type.

To solve this problem of passing any set
of credentials the OCM context object is
used to store credentials. This handled
by a sub context, the *Credentials context*.

As usual, we start with the default OCM context.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{default context}}
```

It is now used to gain access to the appropriate
credential context.


```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{cred context}}
```

The credentials context brings together
providers of credentials, for example a
vault or a local docker/config.json
and credential consumers like GitHub or
OCI registries.
It must be able to distinguish various kinds
of consumers. This is done by identifying
a dedicated consumer with a set of properties
called `credentials.ConsumerId`. It consists
at least of a consumer type property and a
consumer type specific set of properties
describing the concrete instance of such
a consumer, for example an OCI artifact in
an OCI registry is identified by a host and
a repository path.

A credential provider like a vault just provides
named credential set and typically does not
know anything about the use case for these sets.
The task of the credential context is now to
provide credentials for a dedicated consumer.
Therefore, it maintains a configurable
mapping of credential sources (credentials in
a credential repository) and a dedicated consumer.

This mapping defines a usecase, also based on
a property set and dedicated credentials.
If credentials are required for a dedicated
consumer, it matches the defined mappings and
returned the best matching entry.

Matching? Let's take GitHub OCI registry as an
example. There are different owners for
different repository path (the GitHub org/user).
Therefore, different credentials needs to be provided
for different repository paths.
For example credentials for ghcr.io/acme can be used
for a repository ghcr.io/acme/ocm/myimage.

To start with the credentials context we just
provide an explicit mapping for our use case.

First, we create our credentials object as before.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{new credentials}}
```

Then we determine the consumer id for our use case.
The repository implementation provides a function
for this task. It provides the most common property
set (no repository path) for an OCI based OCM repository.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{consumer id}}
```

The used functions above are just convenience wrappers
around the core type ConsumerId, which might be provided
for dedicated repository/consumer technologies.
Everything can be done directly with the core interface
and property name constants provided by the dedicted technologies. 

Once we have the id we can finnaly set the credentials for this
id.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{set credentials}}
```

Now, the context is prepared to provide credentials 
for any usage of our OCI registry
Lets test, whether it could provide credentials 
for storing our component version.

First, we get the repository object for our OCM repository.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{get repository}}
```

Second, we determine the consumer id for our intended repository acccess.
A credential consumer may provide might provide consumer id information
for a dedicated sub user context.
This is supported by the OCM repo implementation for OCI registries.
The usage context is here the component name.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{get access id}}
```

Third, we ask the credential context for appropriate credentials.
The basic context method `credctx.GetCredentialsForConsumer` returns
a credentials source interface able to provide credentials
for a changing credentials source. Here, we use a convenience
function directly providing a credentials interface for the
actually valid credentials.
An error is only provided if something went wrong while determining
the credentials. Delivering NO credentials is a valid result.
the returned interface then offers access to the credential properties.
via various methods.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{get credentials}}
```

Now, we can continue with our basic component version composition
from the last example, or we just display the content.

The following code snipped show the code for the `context` variant
creating a new version, the `read` variant just omits the version creation.
The rest of the example is then the same.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{add version}}
```

Let's verify the created content and list the versions as known from tour 1.
OCI registries do not support component listers, therefore we
just get and describe the actually added version.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{show version}}
```

As we can see in the resource list, our image artifact has been
uploaded to the OCI registry as OCI artifact and the access method has be changed
to `ociArtifact`. It is not longer a local blob.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{examine cli}}
```

This resource access points effectively to the same OCI registry,
but a completely different repository.
If you are using *ghcr.io*, this freshly created repo is private,
therefore, we need credentials for accessing the content.
An access method also acts as credential consumer, which
tries to get required credentials from the credential context.
Optionally, an access method can act as provider for a consumer id, so that
it is possible to query the used consumer id from the method object.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{image credentials}}
```

Because the credentials context now knows the required credentials,
the access method as credential consumer can access the blob.

```go
{{include}{../../03-working-with-credentials/02-basic-credential-management.go}{image access}}
```

### Providing credentials via credential repositories

The OCM toolset embraces multiple storage
backend technologies, for OCM metadata as well
as for artifacts described by a component version. 
All those technologies typically have their own
way to configure credentials for command line
tools or servers.

The credential management provides so-called
credential repositories. Such a repository
is able to provide any number of named
credential sets. This way any special
credential store can be connected to the
OCM credential management judt by providing
an own implementation for the repository interface.

One such case is the docker config json, a config
file used by <code>docker login</code> to store
credentials for dedicated OCI registries.

We start again by providing access to the
OCM context and the connected credential context.


```go
{{include}{../../03-working-with-credentials/03-credential-repositories.go}{context}}
```

In package `.../contexts/credentials/repositories` you can find
packages for predefined implementations for some standard credential repositories,
for example `dockerconfig`.

```go
{{include}{../../03-working-with-credentials/03-credential-repositories.go}{docker config}}
```

There are general credential stores, like a HashiCorp Vault
or type-specific ones, like the docker config json
used to configure credentials for the docker client.
(working with OCI registries).
Those specialized repository implementations are not only able to
provide credential sets, they also know about the usage context
of the provided credentials.
Therefore, such repository implementations are able to provide credential
mappings for consumer ids, also. This is supported by the credential repository
API provided by this library.

The docker config is such a case, so we can instruct the
repository to automatically propagate appropriate the consumer id
mappings. This feature is typically enabled by a dedicated specfication
option.

```go
{{include}{../../03-working-with-credentials/03-credential-repositories.go}{propagation}}
```

Implementations for more generic credential repositories can also use this
feature, if the repository allows adding arbitrary metadata. This is for
example used by the `vault` implementation. It uses dedicated attributes
to allow the user to configure intended consumer id properties.

now we can just add the repository for this specification to
the credential context by getting the repository object for our
specification.

```go
{{include}{../../03-working-with-credentials/03-credential-repositories.go}{add repo}}
```

We are not interested in the repository object, so we just ignore
the result.

So, if you have done the appropriate docker login for your 
OCI registry, it should be possible now to get the credentials
for the configured repository.

We first query the consumer id for the repository, again.

```go
{{include}{../../03-working-with-credentials/03-credential-repositories.go}{get consumer id}}
```

and then get the credentials from the credentials context like in the previous example.

```go
{{include}{../../03-working-with-credentials/03-credential-repositories.go}{get credentials}}
```
