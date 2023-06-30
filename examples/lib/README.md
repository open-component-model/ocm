# How to work with the library

The complete library is arranged around area specific [`Context` objects](contexts.md),
which bundle all the settings and supported [extension points](../../docs/ocm/interoperability.md#support-library)
of the Open Component Model.
Extension points are implemented by handlers that can be registered at dedicated
context objects or at the default runtime environment.
The context then provides all the methods required to access elements
managed in the dedicated area.


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
- [creating OCM content in temporary CTFs and publishing it](transfer.md)