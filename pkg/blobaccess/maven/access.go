package maven

import (
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/maven"
)

func DataAccess(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (bpi.DataAccess, error) {
	return BlobAccess(repo, groupId, artifactId, version, opts...)
}

func BlobAccess(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (bpi.BlobAccess, error) {
	eff := optionutils.EvalOptions(opts...)
	s := &spec{
		coords:  maven.NewCoordinates(groupId, artifactId, version, maven.WithOptionalClassifier(eff.Classifier), maven.WithOptionalExtension(eff.Extension)),
		repo:    repo,
		options: eff,
	}
	return s.getBlobAccess()
}

func BlobAccessForCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) (bpi.BlobAccess, error) {
	return BlobAccess(repo, coords.GroupId, coords.ArtifactId, coords.Version, optionutils.WithDefaults(opts, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))...)
}

func Provider(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccess(repo, groupId, artifactId, version, opts...)
		return b, err
	})
}

func ProviderCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) bpi.BlobAccessProvider {
	return Provider(repo, coords.GroupId, coords.ArtifactId, coords.Version, optionutils.WithDefaults(opts, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))...)
}
