# How to work with the library

The complete library is arranged around area specific `Context` objects,
which bundle all the settings and supported [extension points](../../docs/ocm/interoperability.md#support-library)
of the Open Component Model.
Extension points are implemented by handlers that can be registered at dedicated
context objects or at the default runtime environment.
The context then provides all the methods required to access elements
managed in the dedicated area.

## Contexts

A context object is the entry point for using a dedicated functional areas. It bundles all
settings and extensions point implementations for this area.

Therefore, it provides a root object of the type `Context`. This context
object then provides methods to 
- set configurations and to
- get access to elements belonging to this area.

There might be any number of such context objects at the same time and with
different settings. Context objects are typically intended to have a short
lifetime (for example to execute a dedicated request) and can be 
garbage collected afterwards.

The basic elements of most context types are *Specifications* and *Repositories*.
A specification object provides serializable attributes used to describe
dedicated elements in the functional area. Specifications are typed. There might
be different types (with different attribute sets) used to describe instances provided
by different implementations. The root element below the context object is typically
a *Repository* object, which provides access to elements hosted by this repository.

The context itself manages all the specification and element types and provides
an entry point to deserialize specifications and to gain access to described
effective root elements.

For example, the OCI context manages repository specifications and types
used to describe instances of various types of repositories hosting OCI Artifacts
(one such specification/repository type is an *OCI Registry*, another one the docker daemon and a third one a filesystem representation for storing OCI artifacts). 
The repository object provided for a repository specification then provides
access to namespaces (in OCI speak *OCI repositories*), which again provide
access to OCI artifacts (versions): manifests and indices.

More complex contexts (especially the OCM context) may offer access to a more
complex object ecosystem, for more kinds of specifications and object types.

All functional areas supported by contexts can be found as sub packages of
`github.com/open-component-model/ocm/pkg/contexts`. The following context
types are provided:

- `config`: configuration management of all parts of the OCM library.
- `credentials`: credential management
- `oci`: working with OCI registries
- `ocm`: working withOCM repositories
- `clictx`: command line interface 
- `datacontext`: base context functionality used for all kinds of contexts.

To just use the library without special configuration the standard settings will
be available by accessing the default context for the area of question.
Alternatively, context instances with specialized configurations can be 
orchestrated by context builders.

A context package contains the typical user API for the functional area.
Elements required to provide own extension point implementations
(for example new specification and repository types) can be found in the
`cpi`(Context Progamming Interface) sub-package. Internal implementation
utilities are located in the `internal` sub package. It is not intended to
be used outside the context package.

## Working with OCM repositories

For working with [OCM repositories](../../docs/ocm/model.md#repositories) an appropriate
context, which can be used to retrieve OCM repositories, can be accessed with:

```go
import "github.com/open-component-model/ocm/pkg/contexts/ocm"


func MyFirstOCMApplication() {
   octx := ocm.DefaultContext()
   
   ...
}
```

If a decoupled environment with dedicated special settings is required, the 
builder methods of the ocm package (`With...`) can be used to compose
a context.

With `ocm.New()` a fresh `ocm` context is created using the default settings.
It is possible to create any number of such contexts.

The context can then be used to gain access to (OCM) repositories, which
provide access to hosted components and component versions.


To access a repository, a [repository specification](../../docs/formats/repositories/README.md)
is required. Every repository type extension supported by this library 
uses its own package under `github.com/open-component-model/ocm/pkg/contexts/ocm/repositories`.
To access an OCM repository based on an OCI registry the package
`github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg`
contains the appropriate language binding for the OCI registry mappings.

Those packages typically have the method `NewRepositorySpec` to create
an appropriate specification object for a concrete instance.

```go
  repoSpec := ocireg.NewRepositorySpec("ghcr.io/mandelsoft/ocm", nil)

  repo, err := octx.RepositoryForSpec(repoSpec)
  if err != nil {
          return err
  }
  defer repo.Close()
```

Once a repository object is available it can be used to access component versions.

```go
  compvers, err := repo.LookupComponentVersion(componentName, componentVersion)
  if err != nil {
          return err
  }
  defer compvers.Close()
```

The component version now provides access to the described content, the
[component descriptor](../../docs/ocm/model.md#component-descriptor),
[resources](../../docs/ocm/model.md#resources),
[sources](../../docs/ocm/model.md#sources),
[component references](../../docs/ocm/model.md#references), and
[signatures](../../docs/ocm/model.md#signatures).

The component descriptor is accessible by a standardized Go representation,
which is independent of the actually used serialization format.
If can be encoded again in the original or any other supported scheme versions.

```go
  cd := compvers.GetDescriptor()
  data, err := compdesc.Encode(cd)
  if err != nil {
          return err
  }

```

Any resource (or source) can be accessed by getting the appropriate
resource object by its resource [identity](../../docs/ocm/model.md#identity) in
the context of the  component version.

```go
  res, err := compvers.GetResource(metav1.NewIdentity(resourceName))
  if err != nil {
          return err
  }

  fmt.Printf("resource %s:\n  type: %s\n", resourceName, res.Meta().Type)
```

The content of a described resource can be accessed using the appropriate
[access method](../../docs/ocm/model.md#artifact-access) described as part of
the resource specification (another extension point of the model).
It is described by an access specification. Supported methods can be 
directly be requested using the resource object.

```go
  meth, err := res.AccessMethod()
  if err != nil {
          return err
  }
  defer meth.Close()

  fmt.Printf("  mime: %s\n", meth.MimeType())
```

The access method then provides access to the technical blob content.
Here a stream access or a byte array access is possible.

```go
  data, err = meth.Get()
  if err != nil {
          return err
  }

  fmt.Printf("  content:\n%s\n", utils.IndentLines(string(data), "    ",))
```

Besides this simple example, there are more usage scenarios, which
typicaly require more configuration:
- [creating content and credential handling](creds.md)
- [dealing with configuration](config.md)