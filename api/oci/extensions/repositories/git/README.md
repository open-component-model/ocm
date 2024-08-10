
# Repository `GitRepository` - git based repository


### Synopsis

```
type: GitRepository/v1
```

### Description

Artifact namespaces/repositories of the API layer will be mapped to git repository paths.

Supported specification version is `v1`.

### Specification Versions

#### Version `v1`

The type specific specification fields are:

- **`url`** *string*
    
  URL of the git repository in the form of <url>@<ref>#<path> ^([^@#]+)(@[^#\n]+)?(#[^@\n]+)?
  - url is the URL of the git repository
  - ref is the git reference to checkout, if not specified, defaults to "HEAD"
  - path is the path to the file or directory to use as the source, if not specified,defaults to the root of the repository.

### Go Bindings

The Go binding can be found [here](type.go)
