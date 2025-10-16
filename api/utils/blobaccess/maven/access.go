package maven

import (
	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

func DataAccess(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (bpi.DataAccess, error) {
	return BlobAccess(repo, groupId, artifactId, version, opts...)
}

func BlobAccess(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (bpi.BlobAccess, error) {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	s := &spec{
		coords:  maven.NewCoordinates(groupId, artifactId, version, maven.WithOptionalClassifier(eff.Classifier), maven.WithOptionalExtension(eff.Extension), maven.WithOptionalMediaType(eff.MediaType)),
		repo:    repo,
		options: &eff,
	}
	return s.getBlobAccess()
}

func BlobAccessForCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) (bpi.BlobAccess, error) {
	return BlobAccess(repo, coords.GroupId, coords.ArtifactId, coords.Version, optionutils.WithDefaults(opts, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension), WithOptionalMediaType(coords.MediaType))...)
}

func Provider(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccess(repo, groupId, artifactId, version, opts...)
		return b, err
	})
}

func ProviderCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) bpi.BlobAccessProvider {
	return Provider(repo, coords.GroupId, coords.ArtifactId, coords.Version, optionutils.WithDefaults(opts, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension), WithOptionalMediaType(coords.MediaType))...)
}
