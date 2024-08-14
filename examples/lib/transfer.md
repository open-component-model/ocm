# Creating OCM content in temporary CTFs and publishing it

OCM content can be directly crated in public OCM repositories as shown in other examples.
But it is also possible to compose component versions in local or temporary filesystem
structured according the [Common Transport Format](https://github.com/open-component-model/ocm-spec/blob/main/doc/01-model/01-model.md#component-repositories).
Afterward, they can be transferred/published to public OCM repositories.

This example shows some cases how this can be achieved directly using the OCM library.
A complete example can be found in [transfer1](transfer1/example.go).

## Using standard OCM configuration files to configure a Context

If you do not want to configure your credentials or other settings
directly using the API, you could apply the standard CLI configuration
files directly via API calls.

```go
  // configure default context by evaluating standard config sources
  err = configutils.Configure("")
  if rerr != nil {
    return err
  }
```

## Handling of Close calls and error propagation

Many API calls provide closeable objects. They must be closed to release
potential external or internal resources held for this objects.

To be able to catch potential errors provided by those methods, simple
defer statements like `defer x.Close()` should be avoided. To catch errors
appearing during the cleanup of a function body, a finalizer can be used:

```go
  func TransferApplication() (rerr error) {
    // setup error propagation for deferred cleanup/close methods.
    var finalize finalizer.Finalizer
    defer finalize.FinalizeWithErrorPropagation(&rerr)
    // ...
  }
```

This code snippet creates a finalizer object and requests the handling of
all finalization code at the end of the function. Potential errors occurring
during the cleanup are incorporated into the error return of the function
call. `Close` calls can then simply be added by `finalize.Close(x)`.

## Creating a temporary orchestration environment

First, a temporary filesystem is created, which is then used to
create a CTF directory structure.

```go
    // import "github.com/mandelsoft/vfs/pkg/memoryfs"

  // create a temporary orchestration environment for a set of
  // component versions. We use a CTF here stored either
  // in a temporary filesystem folder or in memory.
  tmpfs, err := osfs.NewTempFileSystem()
  if err != nil {
      return err
  }
  finalize.With(func() error { return vfs.Cleanup(tmpfs) })

  // if you have not much direct blob content, you could use
  // a memory filesystem instead
  // tmpfs:=memoryfs.New()

  repo, err := ctf.Open(octx, accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(tmpfs), accessio.FormatDirectory)
  if err != nil {
      return errors.Wrapf(err, "cannot create CTF")
  }
  finalize.Close(repo)
```

Instead of using a temporary filesystem it is also possible to create the CTF file
directly in some path in your filesystem to keep the content for later usage.
Or, you can directly use a memory filesystem, if the size of the intended blobs is
very limited.

The resulting repository can be used like any other OCM repository implementation
to orchestrate new component versions using the standard repository API of the
OCM context.

This can be done as shown in the [credential examples](creds.md).

The next step then is to transfer the orchestrated content to another OCM repository.
The basic functionality provided by the library is the transport of a dedicated
component version (optionally by traversing all the component references).
CTF objects support a Lister interface, which can be used to discover contained
components.

This can be easily combined to transfer the complete content. First, we have
get access to the intended target repository:

```go
  uni, err := ocm.ParseRepo(cfg.Repository)
  if err != nil {
      return errors.Wrapf(err, "invalid repo spec")
  }
  repoSpec, err := octx.MapUniformRepositorySpec(&uni)
  if err != nil {
      return errors.Wrapf(err, "invalid repo spec")
  }

  // if you know you have an OCI registry based OCM repository
  // repoSpec := ocireg.NewRepositorySpec(cfg.Repository)

  // if you want to provide specific credentials....
  // target, err := octx.RepositoryForSpec(repoSpec, cfg.GetCredentials())

  // use credentials from config context (for example initialized by Configure above)
  target, err := octx.RepositoryForSpec(repoSpec)
  if err != nil {
      return err
  }
  finalize.Close(target)
```

To parse an OCM repository reference you can use the `ParseRepo` function.
It provides a uniform representation of a parsed string representation.
This one can then be mapped to a regular `RespositorySpec` object, which is mapped by the OCM context to a repository implementation.

Instead of this string parsing, an appropriate repository specification object
can directly be created as shown in the other examples.

Once we have access to the target repository we just list the
components and subsequent versions contained in the CTF, which are then
transferred to the target repository.

As preparation step we create a standard transfer handler,
which controls the transfer process. The standard handler
just offers some commonly used options, like transfer-by-value
for the found resources. In this example we just want to keep
the location of the resources as they are provided by the CTF,
but to potentially overwrite existing component versions in the
target repository (`transferHandler, err := standard.New(standard.Overwrite())`.

```go
  lister := repo.ComponentLister()
  if lister == nil {
      return fmt.Errorf("repo does not support lister")
  }
  comps, err := lister.GetComponents("", true)
  if rerr != nil {
      return errors.Wrapf(err, "cannot list components")
  }

  printer := common.NewPrinter(os.Stdout)
  closure := transfer.TransportClosure{}
  transferHandler, err := standard.New(standard.Overwrite())
  if rerr != nil {
      return err
  }
  for _, cname := range comps {
      loop := finalize.Nested()

      c, err := repo.LookupComponent(cname)
      if err != nil {
          return errors.Wrapf(err, "cannot get component %s", cname)
      }
      loop.Close(c)

      vnames, err := c.ListVersions()
      if err != nil {
          return errors.Wrapf(err, " cannot list versions for component %s", cname)
      }

      for _, vname := range vnames {
          loop := loop.Nested()

          cv, err := c.LookupVersion(vname)
          if err != nil {
              return errors.Wrapf(err, "cannot get version %s for component %s", vname, cname)
          }
          loop.Close(cv)

          err = transfer.TransferVersion(printer, closure, cv, target, transferHandler)
          if err != nil {
              return errors.Wrapf(err, "cannot transfer version %s for component %s", vname, cname)
          }

          if err := loop.Finalize(); err != nil {
              return err
          }
      }
      if err := loop.Finalize(); err != nil {
          return err
      }
  }
```

With `closure := transfer.TransportClosure{}` a shared transport store is
created, which remembers already transported component versions. It is
used for all calls of `TransferVersion` to avpid duplicate transfers.
THis is especially relevant, if the transitive transfer option is set.
In this example this all content of the CTF is transferred without the
transitive option, so it is not necessarily required.

But if your setup creates component versions with references to component
versions not contained in the CTF, the transitive option might be useful
to assure the completeness of your component versions in the target repository.
To resolve those external references a resolver must be specified for the
transfer handler.

## Defer in loops

One common problem in Go is the finalization of elements in loops.
We have here two nested loops and want to close elements allocated in a
loop step as soon as possible directly when the loop step is finished.

The problem is, that the `Close` methods should also be called, if the
complete function is left in case of an error occurring in a loop step.
Therefore, we cannot just put the call to the `Close` methods at the end of the loop.
But, also the regular `defer` mechanism cannot be used, because it would delay
the execution of all `Close` methods to the final end of the function.
Using both a `defer` and an explicit `Close` at the end of the loop step is also
not possible, because then `Close` would potentially be called twice.

Here, the used finalizer object can help. It is possible to
create a nested finalizer specifically used inside a loop step:

```go
  for ... {
      loop := finalize.Nested()
      ...
      loop.Close(c)
      ...
      if err := loop.Finalize(); err != nil {
          return err
      }
  }
```

The nested finalizer can be finalized at the end of the loop,
but it is also executed once the function is left for the `defer` of
the root finalizer at the beginning of the function, if it is not yet
executed. Once it has been explicitly called it is automatically removed
from the next outer scope.
