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
		coords:  maven.NewCoordinates(groupId, artifactId, version, eff.Classifier, eff.Extension),
		repoUrl: repoUrl,
		options: eff,
	}
	return s.getBlobAccess()
}

func BlobAccessProviderForMaven(repoUrl, groupId, artifactId, version string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccessForMaven(repoUrl, groupId, artifactId, version, opts...)
		return b, err
	})
}
