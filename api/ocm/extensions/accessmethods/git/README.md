
# Access Method `git` - Git Commit Access

## Synopsis

```yaml
type: Git/v1
```

Provided blobs use the following media type for: `application/x-tgz`

The artifact content is provided as a gzip-compressed tar archive. `Git/v1`
archives contain the checked-out files and a `.git` directory with a detached
`HEAD` and the complete history reachable from that commit. The legacy type
spellings (`git`, `git/v1alpha1`, `Git`, and `Git/v1alpha1`) are still accepted
and produce the same history-bearing archive.

### Description

This method implements the access of the content of a git commit stored in a
git repository.

The current access type is `Git/v1`. The legacy specification version is
`v1alpha1`.

### Specification Versions

#### Version `v1`

The type specific specification fields are:

- **`repository`**  *string*

  Repository URL with or without scheme.

- **`ref`** (optional) *string*

  Original ref used to get the commit from

- **`commit`** *string*

  The sha/id of the git commit

### Go Bindings

The Go binding is defined in [method.go](method.go).

#### Example

```go
package main

import (
  "archive/tar"
  "bytes"
  "compress/gzip"
  "fmt"
  "io"

  "ocm.software/ocm/api/ocm"
  "ocm.software/ocm/api/ocm/cpi"
  me "ocm.software/ocm/api/ocm/extensions/accessmethods/git"
)

func main() {
  ctx := ocm.New()
  accessSpec := me.New(
    "https://github.com/octocat/Hello-World.git",
    me.WithRef("refs/heads/master"),
  )
  method, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
  if err != nil {
    panic(err)
  }
  content, err := method.GetContent()
  if err != nil {
    panic(err)
  }
  unzippedContent, err := gzip.NewReader(bytes.NewReader(content))

  r := tar.NewReader(unzippedContent)

  var file *tar.Header
  for {
    file, err = r.Next()
    if err != nil {
      panic(err)
    }
    if file.Name == "README.md" {
      break
    }
  }

  data, err := io.ReadAll(r)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(data))
}
```
