package maven

import (
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type BlobMeta = maven.FileMeta

func DataAccessForMaven(repoUrl, groupId, artifactId, version string, opts ...Option) (blobaccess.DataAccess, error) {
	return BlobAccessForMaven(repoUrl, groupId, artifactId, version, opts...)
}

func BlobAccessForMaven(repoUrl, groupId, artifactId, version string, opts ...Option) (blobaccess.BlobAccess, error) {
	eff := optionutils.EvalOptions(opts...)
	s := &spec{
		coords:  maven.NewCoordinates(groupId, artifactId, version, maven.WithOptionalClassifier(eff.Classifier), maven.WithOptionalExtension(eff.Extension)),
		repoUrl: repoUrl,
		options: eff,
	}
	return s.getBlobAccess()
}

func BlobAccessForMavenCoords(repoUrl string, coords *maven.Coordinates, opts ...Option) (blobaccess.BlobAccess, error) {
	return BlobAccessForMaven(repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, append([]Option{WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension)}, opts...)...)
}

func BlobAccessProviderForMaven(repoUrl, groupId, artifactId, version string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccessForMaven(repoUrl, groupId, artifactId, version, opts...)
		return b, err
	})
}

func BlobAccessProviderForMavenCoords(repoUrl string, coords *maven.Coordinates, opts ...Option) bpi.BlobAccessProvider {
	return BlobAccessProviderForMaven(repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, append([]Option{WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension)}, opts...)...)
}
