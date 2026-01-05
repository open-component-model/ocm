package download

import (
	"github.com/spf13/pflag"

	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{}
}

type Option struct {
	Layers    []int
	DirTree   bool
	OCILayout bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.IntSliceVarP(&o.Layers, "layers", "", nil, "extract dedicated layers")
	fs.BoolVarP(&o.DirTree, "dirtree", "", false, "extract as effective filesystem content")
	fs.BoolVarP(&o.OCILayout, "oci-layout", "", false, "download as OCI Image Layout (blobs in blobs/<algorithm>/<encoded>)")
}

func (o *Option) Usage() string {
	return `
With option <code>--layers</code> it is possible to request the download of
dedicated layers, only. Option <code>--dirtree</code> expects the artifact to
be a layered filesystem (for example OCI Image) and provided the effective
filesystem content.

Option <code>--oci-layout</code> changes the blob storage structure in the downloaded
artifact. Without this option, blobs are stored in a flat directory at
<code>blobs/&lt;algorithm&gt;.&lt;encoded&gt;</code> (e.g., <code>blobs/sha256.abc123...</code>).
With this option, blobs are stored in a nested directory structure at
<code>blobs/&lt;algorithm&gt;/&lt;encoded&gt;</code> (e.g., <code>blobs/sha256/abc123...</code>)
as specified by the OCI Image Layout Specification
(see <a href="https://github.com/opencontainers/image-spec/blob/main/image-layout.md">
https://github.com/opencontainers/image-spec/blob/main/image-layout.md</a>).
`
}
