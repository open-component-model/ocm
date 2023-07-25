# Creating OCM Content in a CTF File

For creating a filesystem representation a dedicated kind
of OCM repository implementation can be used: a Common Transport
Format based directory or archive. Here, no credentials
are required. The basic content handling is identical
to the [OCI-based OCM example](creds.md).

## Creating a CTF 

A transport file can be created in a virtual filesystem.
The simplest way is just to use the OS filesystem `osfs.New()`,
which is the default.

In this example we just use a memory based filesystem.

```go
  import "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/repositories/ctf"

  octx := ocm.DefaultContext()

  memfs := memoryfs.New()

  repo, err := ctf.Open(octx, accessobj.ACC_CREATE, "test", 0o700, accessio.PathFileSystem(memfs))
  if err != nil {
      return err
  }
```

For the complete example, please have a look at the [complete code](ctf/example.go).
