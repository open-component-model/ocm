# How to work with the library

The complete library is arranged around area specific [`Context` objects](contexts.md),
which bundle all the settings and supported
[extensions](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/07-extensions.md#extending-the-open-component-model)
of the *Open Component Model*.
Extension points are implemented by handlers that can be registered at dedicated
context objects or at the default runtime environment.
The context then provides all the methods required to access elements
managed in the dedicated area.

The examples shown here provide an overview of the library.
A more detailed annotated tour through various aspects of the library
with ready-to work examples can be for [here](tour).

In the [comparison scenario](comparison-scenario/README.md) there is
an example for an end-to-end scenario, from providing a component version
by a software provider, over its publishing up to the consumption in an
air-gapped environment, and the final deployment in this environment.
Especially the deployment part just wants to illustrate the basic
workflow using a Helm chart based example. It is not intended to be used
as productive environment.

## Working with OCM repositories

For working with [OCM repositories](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/01-model.md#component-repositories) an appropriate
context, which can be used to retrieve OCM repositories, can be accessed with:

```go
import "ocm.software/ocm/api/ocm"


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

To access a repository, a [repository specification](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/01-model.md#component-repositories)
is required. Every repository type extension supported by this library
uses its own package under `ocm.software/ocm/api/ocm/extensions/repositories`.
To access an OCM repository based on an OCI registry the package
`ocm.software/ocm/api/ocm/extensions/repositories/ocireg`
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
[component descriptor](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/01-model.md#components-and-component-versions),
[resources](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/02-elements-toplevel.md#resources),
[sources](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/02-elements-toplevel.md#sources),
[component references](https://github.com/open-component-model/ocm-spec/blob/main/doc/05-guidelines/03-references.md), and
[signatures](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/03-elements-sub.md#signatures).

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
resource object by its resource [identity](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/02-elements-toplevel.md#component-identity) in
the context of the  component version.

```go
  res, err := compvers.GetResource(metav1.NewIdentity(resourceName))
  if err != nil {
          return err
  }

  fmt.Printf("resource %s:\n  type: %s\n", resourceName, res.Meta().Type)
```

The content of a described resource can be accessed using the appropriate
[access method](https://github.com/open-component-model/ocm-spec/blob/main/doc/04-extensions/02-access-types) described as part of
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
typically require more configuration:

- [creating content and credential handling](creds.md)
- [dealing with configuration](config.md)
- [creating OCM content in temporary CTFs and publishing it](transfer.md)

## End-to-end Scenario

In folder [`comparison-scenario`](comparison-scenario/README.md) there is
an example for an end-to-end scenario,
from building a component version to publishing, consuming and deploying
it in a separate environment. It shows the usage of the OCM library to
implement all the required process steps.

It builds a component version for the [podinfo helm chart](https://artifacthub.io/packages/helm/podinfo/podinfo).
There are two scenarios:

- provisioning
  - building a component version with a helm based deployment description
  - signing it and
  - publishing it
- consumption in a separate repository environment
  - transferring the component version into a separate repository environment.
  - using the deployment description to localize the helm chart value - preparing
    values to refer to the podinfo OCI image available in the local environment.
  - finally deploying it with helm.
