package maven

import (
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type BlobMeta = maven.FileMeta

func DataAccessForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (blobaccess.DataAccess, error) {
	return BlobAccessForMaven(repo, groupId, artifactId, version, opts...)
}

func BlobAccessForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (blobaccess.BlobAccess, error) {
	eff := optionutils.EvalOptions(opts...)
	s := &spec{
		coords:  maven.NewCoordinates(groupId, artifactId, version, maven.WithOptionalClassifier(eff.Classifier), maven.WithOptionalExtension(eff.Extension)),
		repo:    repo,
		options: eff,
	}
	return s.getBlobAccess()
}

func BlobAccessForMavenCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) (blobaccess.BlobAccess, error) {
	return BlobAccessForMaven(repo, coords.GroupId, coords.ArtifactId, coords.Version, append([]Option{WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension)}, opts...)...)
}

func BlobAccessProviderForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccessForMaven(repo, groupId, artifactId, version, opts...)
		return b, err
	})
}

func BlobAccessProviderForMavenCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) bpi.BlobAccessProvider {
	return BlobAccessProviderForMaven(repo, coords.GroupId, coords.ArtifactId, coords.Version, append([]Option{WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension)}, opts...)...)
}
