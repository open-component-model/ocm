
# Repository `GitRepository` - git based repository

## Synopsis

```yaml
type: GitRepository/v1
```

### Description

Artifact namespaces/repositories of the API layer will be mapped to git repository paths.

Supported specification version is `v1`.

### Specification Versions

#### Version `v1`

The type specific specification fields are:

- **`url`** *string*

  URL of the git repository in any standard git URL format.
  The schemes `http`, `https`, `git`, `ssh` and `file` are supported.

- **`ref`** *string*

  The git reference to use. This can be a branch, tag, or commit hash. The default is `HEAD`, pointing to the default branch of a repository in most implementations.

### Go Bindings

The Go binding can be found [here](type.go)
