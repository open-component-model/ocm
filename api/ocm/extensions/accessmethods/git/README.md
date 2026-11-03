
# Access Method `git` - Git Commit Access

## Synopsis

```yaml
type: Git/v1
```

Provided blobs use the following media type for: `application/x-tgz`

The artifact content is provided as a gzip-compressed tar archive. Both
`Git/v1` and its alias `Git` produce archives with the checked-out
files and a `.git` directory containing a detached `HEAD` and the complete
history reachable from that commit. The legacy type spellings (`git`,
`git/v1alpha1`, and `Git/v1alpha1`) retain their files-only archive format.

#### Archive layout

The archive is byte-identical to the one OCM v2 builds for the same commit, so a
resource created by either implementation verifies against the digest recorded by
the other. It therefore depends on the commit alone, not on how the source
repository stores its objects nor on the host the archive was built on: tar
headers carry no owner and no modification time, directories are `0755`, files
`0644` or `0755`, and symlinks `0777`.

Next to the checked-out files it holds exactly these entries:

```text
.git/HEAD                                  the commit, detached, without refs
.git/config                                core defaults of an initialized repository
.git/objects/info/                         (empty)
.git/objects/pack/pack-<checksum>.pack     the objects reachable from the commit,
.git/objects/pack/pack-<checksum>.idx      in hash order, submodules excluded
.git/objects/pack/pack-<checksum>.rev
.git/refs/heads/                           (empty)
.git/refs/tags/                            (empty)
```

Everything else a clone holds - the source remote, refs, the index and logs -
stays out of the archive. The same archive is produced by the CLI `git` input
type, so a local blob created with it can be uploaded back into a Git repository
by OCM v2.

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
