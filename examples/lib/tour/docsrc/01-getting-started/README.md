# Basic Usage of OCM Repositories
{{getting-started}}

This [tour](example.go) illustrates the basic usage of the API to
access component versions in an OCM repository.

## Running the example

You can just call the main program with some config file argument
with the following content:

```yaml
component: github.com/mandelsoft/examples/cred1
repository: ghcr.io/mandelsoft/ocm
version: 0.1.0
```

## Walkthrough

The basic entry point for using the OCM library is always
an [OCM Context object](../../contexts.md). It bundles all
configuration settings and type registrations, like
access methods, repository types, etc, and
configuration settings, like credentials,
which should be used when working with the OCM
ecosystem.

Therefore, the first step is always to get access to such
a context object. Our example uses the default context
provided by the library, which covers the complete
type registration contained in the executable.

It can be accessed by a function of the ocm package.

```go
{{include}{../../01-getting-started/example.go}{default context}}
```

The context acts as the central entry
point to get access to OCM elements.
First, we get a repository, to look for
component versions. We use the OCM
repository providing the standard OCM
components hosted on `ghcr.io`.

For every storage technology used to store
OCM components, there is a serializable
descriptor object, the *repository specification*.
It describes the information required to access
the repository and can be used to store the serialized
form as part of other resources, for example
Kubernetes resources or configuration settings.
The available repository implementations can be found
under `.../pkg/contexts/ocm/repositories`.

```go
{{include}{../../01-getting-started/example.go}{repository spec}}
```

The context can now be used to map the descriptor
into a repository object, which then provides access
to the OCM elements stored in this repository.

```go
{{include}{../../01-getting-started/example.go}{repository}}
```

Many objects must be closed, if they should not be used
anymore, to release potentially allocated temporary resources.
This is typically done by a `defer` statement placed after a
successful object retrieval.

```go
{{include}{../../01-getting-started/example.go}{close}}
```

Now we look for the versions of the component
available in this repository.

```go
{{include}{../../01-getting-started/example.go}{versions}}
```

OCM version names must follow the semver rules.
Therefore, we can simply order the versions and print them.

```go
{{include}{../../01-getting-started/example.go}{semver}}
```

Now, we have a look at the latest version, it is
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

```
{{execute}{go}{run}{../../01-getting-started}{<extract>}{version}}
```

Resources have some metadata, like the resource identity and a resource type.
And they describe how the content of the resource (as blob) can be accessed.
This is done by an *access specification*, again a serializable descriptor,
like the respository specification.

The component version contains the executables for the OCM CLI
for various platforms. The next step is to
get the executable for the actual environment.
The identity of a resource described by a component version
consists of a set of properties. The property name must
always be given.

A convention is to use dedicated labels to indicate the operating system
and the architecture for executables.

```go
{{include}{../../01-getting-started/example.go}{find executable}}
```

Now we want to retrieve the executable. There are two basic ways
to do this.

First, there is the direct way to gain access to the blob by using
the basic model operations to get a reader for the resource blob.
Therefore, in a first step we get the access method for the resource

```go
{{include}{../../01-getting-started/example.go}{getting access}}
{{include}{../../01-getting-started/example.go}{closing access}}
```

The method needs to be closed, because the method
object may cache the technical blob representation
generated accessing the underlying access technology.
(for example, accessing an OCI image requires a sequence of
backend accesses for the manifest, the layers, etc which will
then be packaged into a tar archive returned as blob).
This caching may not be required, if the backend directly
returns a blob.

Now we get access to the reader providing the blob content.
The blob features a mime type, which can be used to understand
the format of the blob. Here, we have a plain octet stream.

```go
{{include}{../../01-getting-started/example.go}{getting reader}}
```

Because this sequence is a common operation, there is a
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
output file.

```go
{{include}{../../01-getting-started/example.go}{copy}}
```

Another way to download a resource is to use registered downloaders.
`download.DownloadResource` is used to download resources with specific handlers for the
selected resource and mime type combinations.
The executable downloader is registered by default and automatically
sets the X flag.

```go
{{include}{../../01-getting-started/example.go}{download}}
```
