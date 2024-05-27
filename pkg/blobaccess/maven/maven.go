package maven

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/optionutils"
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
}

func NewCoordinates(groupId, artifactId, version string, opts ...Option) *Coordinates {
	eff := optionutils.EvalOptions(opts...)
	return maven.NewCoordinates(groupId, artifactId, version, &optionwrapper{eff})
}
