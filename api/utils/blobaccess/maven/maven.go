package maven

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/tech/maven"
)

type (
	BlobMeta    = maven.FileMeta
	Repository  = maven.Repository
	Coordinates = maven.Coordinates
)

func NewFileRepository(path string, fss ...vfs.FileSystem) *Repository {
	return maven.NewFileRepository(path, fss...)
}

func NewUrlRepository(repoUrl string, fss ...vfs.FileSystem) (*Repository, error) {
	return maven.NewUrlRepository(repoUrl, fss...)
}

type optionwrapper struct {
	options *Options
}

func (o *optionwrapper) ApplyTo(opts *Coordinates) {
	maven.WithOptionalExtension(o.options.Extension).ApplyTo(opts)
	maven.WithOptionalClassifier(o.options.Classifier).ApplyTo(opts)
	maven.WithOptionalMediaType(o.options.MediaType).ApplyTo(opts)
}

func NewCoordinates(groupId, artifactId, version string, opts ...Option) *Coordinates {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	return maven.NewCoordinates(groupId, artifactId, version, &optionwrapper{&eff})
}
