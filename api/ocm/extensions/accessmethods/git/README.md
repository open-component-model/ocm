
# Access Method `git` - Git Commit Access

## Synopsis

```yaml
type: git/v1
```

Provided blobs use the following media type for: `application/x-tgz`

The artifact content is provided as gnu-zipped tar archive

### Description

This method implements the access of the content of a git commit stored in a
git repository.

Supported specification version is `v1`

### Specification Versions

#### Version `v1`

The type specific specification fields are:

- **`repoUrl`**  *string*

  Repository URL with or without scheme.

- **`ref`** (optional) *string*

  Original ref used to get the commit from

- **`commit`** *string*

  The sha/id of the git commit

### Go Bindings

The go binding can be found [here](method.go)


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

  file, err := r.Next()
  if err != nil {
    panic(err)
  }
  
  if file.Name != "README.md" {
    panic("Expected README.md")
  }

  data, err := io.ReadAll(r)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(data))
}
```