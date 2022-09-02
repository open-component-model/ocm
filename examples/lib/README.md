# How to work with the library


The complete library is arranged around area specific `Context` objects,
which bundle all the settings and supported [extension points](../../docs/ocm/interoperability.md#support-library)
of the Open Component Model.
Extension points are implemented by handlers that can registered at dedicated
context objects or at the default runtime environment.

To just use the library the standard settings will be available by accessing the default
contexts for the area of question.

For using [OCM repositories](../../docs/ocm/model.md#repositories) this can be
done with:

```go
import "github.com/open-component-model/ocm/pkg/contexts/ocm"


func MyFirstOCMApplication() {
   octx := ocm.DefaultContext()
}
```


If a decoupled environment with completely local settings is required, the 
builder methods of the ocm package (`With...`) can be used to compose
a context according to dedicated requirements.

With `ocm.New()` a fresh `ocm` context is created using the default settings.
It is possible to create any number of such contexts.

The context can then be used to gain access to (OCM) repositories, which
provide access to hosted components and component versions.


To access a repository, a [repository specification](../../docs/formats/repositories/README.md)
is required. Every repository type extension supported by this library 
uses an own package under `github.com/open-component-model/ocm/pkg/contexts/ocm/repositories`.
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

The component version now provides access to the described content.
Resource, sources and component references.

Additionally, a standardized Go representation of the descriptor
is available.

```go
  cd := compvers.GetDescriptor()
  if err != nil {
          return err
  }

  data, err := compdesc.Encode(cd)
  if err != nil {
          return err
  }

```

Any resource (or source) can be accessed by getting the appropriate
resource object by its resource identity in the context of the
component version.

```go
  res, err := compvers.GetResource(metav1.NewIdentity(resourceName))
  if err != nil {
          return err
  }

  fmt.Printf("resource %s:\n  type: %s\n", resourceName, res.Meta().Type)
```

The content of a described resource can be accessed using the appropriate
[access method](../../docs/ocm/model.md#artefact-access)
(another extension point of the model) of the resource.
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
