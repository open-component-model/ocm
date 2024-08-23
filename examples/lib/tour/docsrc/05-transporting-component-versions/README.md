# Transporting Component Versions

{{transport}}

This [tour](/examples/lib/tour/05-transporting-component-versions/example.go) illustrates the basic support for
transporting content from one environment into another.

## Running the example

You can call the main program with a config file option (`--config <file>`).
The config file should have the following content:

```yaml
repository: ghcr.io/mandelsoft/ocm
targetRepository:
  type: CommonTransportFormat
  filePath: /tmp/example05.target.ctf
  fileFormat: directory
  accessMode: 2
username:
password:
```

Any supported kind of target repository can be specified by using its
specification type. An OCI registry target would look like this:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
targetRepository:
  type: OCIRegistry
  baseUrl: ghcr.io/mandelsoft/targetocm
ocmConfig: <config file>
```

The actual version of the example just works with the file system
target, because it is not possible to specify credentials for the
target repository in this simple config file. But, if you specify an [OCM config file]({{ocm-config-file}}) you can
add more predefined credential settings to make it possible to use
target repositories requiring credentials. The credentials are
automatically taken from the credentials context and don't need to be
specified when creating the repository access object in the code.

## Walkthrough

As usual, we start with getting access to an OCM context

```go
{{include}{../../05-transporting-component-versions/example.go}{default context}}
```

Then we configure this context with optional ocm config defined in our config file.
See [OCM config scenario in tour 04]({{ocm-config-file}}).

```go
{{include}{../../05-transporting-component-versions/example.go}{configure}}
```

This function simply applies the config file using the utility function
provided by the config management:

```go
{{include}{../../05-transporting-component-versions/example.go}{read config}}
```

The context acts as factory for various model types based on
specification descriptor serialization formats in YAML or JSON.
Access method specifications and repository specification are
examples for this feature.

Now, we use the repository specification serialization format to
determine the target repository for a transport from our yaml
configuration file.

```go
{{include}{../../05-transporting-component-versions/example.go}{target}}
```

For our source we just use the component version provided by the last
examples in a remote repository.
Therefore, we set up the credentials context, as
shown in [tour 03]({{using-cred-management}}).

```go
{{include}{../../05-transporting-component-versions/example.go}{set credentials}}
```

For the transport, we first get access to the component version
we want to transport, by getting the source repository and looking up
the desired component version.

```go
{{include}{../../05-transporting-component-versions/example.go}{source}}
```

We could just add this version to the target repository, but this
would not be a real transport, but just a copy of the component descriptor
and the associated local resources. Transport potentially means more, all
the described artifacts should also be copied into the target environment.

Such an action is done by the library function `transfer.Transfer`.
It takes several settings influencing the transport mode,
for example transitive or value transport.
Here, all resources are transported per value, all external
references will be inlined as `localBlob`s and imported into
the target environment, applying blob upload handlers
where possible. For a CTF archive as target, there are no
configured handlers by default. Therefore, all resources will
be migrated to local blobs.

```go
{{include}{../../05-transporting-component-versions/example.go}{transfer}}
```

Now, we check the result of our transport action in the target
repository.

```go
{{include}{../../05-transporting-component-versions/example.go}{verify-a}}
{{include}{../../05-transporting-component-versions/example.go}{verify-b}}
```

Please be aware that all resources in the target now are `localBlob`s,
if the target is a CTF archive. If it is an OCI registry, all the OCI
artifact resources will be uploaded as OCI artifacts into the target
repository and the access specifications are adapted to type `ociArtifact`,
but now referring to OCI artifacts in the target repository.
