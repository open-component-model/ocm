# Basic Usage of OCM Repositories

{{getting-started}}

This [tour](/examples/lib/tour/01-getting-started/example.go) illustrates the basic usage of the API to
access component versions in an OCM repository.

## Running the example

You can call the main program with a config file argument
(`--config <file>`), where the config file has the following content:

```yaml
component: github.com/mandelsoft/examples/cred1
repository: ghcr.io/mandelsoft/ocm
version: 0.1.0
```

{{getting-started-walkthrough}}

## Walkthrough

The basic entry point for using the OCM library is always
an [OCM Context object](/examples/lib/contexts.md). It bundles all
configuration settings and type registrations, like
access methods, repository types, etc, and
configuration settings, like credentials,
which should be used when working with the OCM
ecosystem.

Therefore, the first step is always to get access to such
a context object. Our example uses the default context
provided by the library, which covers the complete
type registration contained in the executable.

It can be accessed by a function of the `api/ocm` package.

```go
{{include}{../../01-getting-started/example.go}{default context}}
```

The context acts as the central entry
point to get access to OCM elements.
First, we get a repository, to look for
component versions. We use the OCM
repository hosted on `ghcr.io`, which is providing the standard OCM
components.

For every storage technology used to store
OCM components, there is a serializable
descriptor object, the *repository specification*.
It describes the information required to access
the repository and can be used to store the serialized
form as part of other resources, for example
Kubernetes resources or configuration settings.
The available repository implementations can be found
under `.../api/ocm/extensions/repositories`.

```go
{{include}{../../01-getting-started/example.go}{repository spec}}
```

The context can now be used to map the descriptor
into a repository object, which then provides access
to the OCM elements stored in this repository.

```go
{{include}{../../01-getting-started/example.go}{repository}}
```

To release potentially allocated temporary resources, many objects
must be closed, if they are not used anymore.
This is typically done by a `defer` statement placed after a
successful object retrieval.

```go
{{include}{../../01-getting-started/example.go}{close}}
```

All kinds of repositories, regardless of their type
feature the same interface to work with OCM content.
It can be used to access stored elements.
First of all, a repository hosts component versions.
They are stored for components. Components are not
necessarily explicit objects stored in an OCM repository.
But they have features like a name and versions. Therefore, the
repository abstraction provided by the library offers
a component object, which can be retrieved from a
repository object. A component has a name and acts as
namespace for versions.

```go
{{include}{../../01-getting-started/example.go}{lookup component}}
```

Now we look for the versions of the component
available in this repository.

```go
{{include}{../../01-getting-started/example.go}{versions}}
```

OCM version names must follow the *SemVer* rules.
Therefore, we can simply order the versions and print them.

```go
{{include}{../../01-getting-started/example.go}{semver}}
```

Now, we have a look at the latest version. It is
the last one in the list.

```go
{{include}{../../01-getting-started/example.go}{lookup version}}
```

{{describe-version}}

The component version object provides access
to the component descriptor

```go
{{include}{../../01-getting-started/example.go}{component descriptor}}
```

and the resources described by the component version.

```go
{{include}{../../01-getting-started/example.go}{resources}}
```

This results in the following output (the shown version might
differ, because the code always describes the latest version):

```text
{{execute}{go}{run}{../../01-getting-started}{<extract>}{version}}
```

Resources have some metadata, like their identity and a resource type.
And, most importantly, they describe how the content of the resource
(as blob) can be accessed.
This is done by an *access specification*, again a serializable descriptor,
like the repository specification.

The component version used here contains the executables for the OCM CLI
for various platforms. The next step is to
get the executable for the actual environment.
The identity of a resource described by a component version
consists of a set of properties. The property `name` is mandatory. But there may be more identity attributes
finally stored as ``extraIdentity` in the component descriptor.

A convention is to use dedicated identity properties to indicate the
operating system and the architecture for executables.

```go
{{include}{../../01-getting-started/example.go}{find executable}}
```

Now we want to retrieve the executable. The library provides two
basic ways to do this.

First, there is the direct way to gain access to the blob by using
the basic model operations to get a reader for the resource blob.
Therefore, in a first step we get the access method for the resource

```go
{{include}{../../01-getting-started/example.go}{getting access}}
{{include}{../../01-getting-started/example.go}{closing access}}
```

The method needs to be closed, because the method
object may cache the technical blob representation
generated by accessing the underlying access technology.
(for example, accessing an OCI image requires a sequence of
backend requests for the manifest, the layers, etc, which will
then be packaged into a tar archive returned as blob).
This caching may not be required, if the backend directly
returns a blob.

Now, we get access to the reader providing the blob content.
The blob features a mime type, which can be used to understand
the format of the blob. Here, we have a plain octet stream.

```go
{{include}{../../01-getting-started/example.go}{getting reader}}
```

Because this code sequence is a common operation, there is a
utility function handling this sequence. A shorter way to get
a resource reader is as follows:

```go
{{include}{../../01-getting-started/example.go}{utility function}}
```

Before we download the content we check the error and prepare
closing the reader, again

```go
{{include}{../../01-getting-started/example.go}{closing reader}}
```

Now, we just read the content and copy it to the intended
output file (`/tmp/ocmcli`).

```go
{{include}{../../01-getting-started/example.go}{copy}}
```

Another way to download a resource is to use registered *downloaders*.
`download.DownloadResource` is used to download resources with specific handlers for
selected resource and mime type combinations.
The executable downloader is registered by default and automatically
sets the `X` flag for the written file.

```go
{{include}{../../01-getting-started/example.go}{download}}
```
