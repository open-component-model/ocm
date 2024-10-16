package git

import (
	"ocm.software/ocm/api/oci/extensions/repositories/git"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/utils/accessobj"
)

const Type = git.Type

type Options = git.Options

type Author = git.Author

func NewRepositorySpec(acc accessobj.AccessMode, url string, opts Options) (*genericocireg.RepositorySpec, error) {
	spec, err := git.NewRepositorySpecFromOptions(acc, url, opts)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
}
