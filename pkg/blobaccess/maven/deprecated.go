package maven

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/maven"
)

// DataAccessForMaven returns a DataAccess for the Maven artifact with the given coordinates.
// Deprecated: use DataAccess.
func DataAccessForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (bpi.DataAccess, error) {
	return DataAccess(repo, groupId, artifactId, version, opts...)
}

// BlobAccessForMaven returns a BlobAccess for the Maven artifact with the given coordinates.
// Deprecated: use BlobAccess.
func BlobAccessForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) (bpi.BlobAccess, error) {
	return BlobAccess(repo, groupId, artifactId, version, opts...)
}

// BlobAccessForMavenCoords returns a BlobAccessProvider for the Maven artifact with the given coordinates.
// Deprecated: use BlobAccessForCoords.
func BlobAccessForMavenCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) (bpi.BlobAccess, error) {
	return BlobAccessForCoords(repo, coords, opts...)
}

// BlobAccessProviderForMaven returns a BlobAccessProvider for the Maven artifact with the given coordinates.
// Deprecated: use Provider.
func BlobAccessProviderForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...Option) bpi.BlobAccessProvider {
	return Provider(repo, groupId, artifactId, version, opts...)
}

// BlobAccessProviderForMavenCoords returns a BlobAccessProvider for the Maven artifact with the given coordinates.
// Deprecated: use ProviderCoords.
func BlobAccessProviderForMavenCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...Option) bpi.BlobAccessProvider {
	return ProviderCoords(repo, coords, opts...)
}
