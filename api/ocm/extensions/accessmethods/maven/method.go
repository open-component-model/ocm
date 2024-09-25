package maven

import (
	"fmt"

	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	mavenblob "ocm.software/ocm/api/utils/blobaccess/maven"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type of Maven repository.
const (
	Type   = "maven"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a Maven artifact.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// RepoUrl is the base URL of the Maven repository.
	RepoUrl string `json:"repoUrl"`

	maven.Coordinates `json:",inline"`
}

// Option defines the interface function "ApplyTo()".
type Option = maven.CoordinateOption

type WithClassifier = maven.WithClassifier

func WithOptionalClassifier(c *string) Option {
	return maven.WithOptionalClassifier(c)
}

type WithExtension = maven.WithExtension

func WithOptionalExtension(e *string) Option {
	return maven.WithOptionalExtension(e)
}

///////////////////////////////////////////////////////////////////////////////

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new Maven repository access spec version v1.
func New(repository, groupId, artifactId, version string, opts ...Option) *AccessSpec {
	accessSpec := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		RepoUrl:             repository,
		Coordinates:         *maven.NewCoordinates(groupId, artifactId, version, opts...),
	}
	return accessSpec
}

// NewForCoordinates creates a new Maven repository access spec version v1.
func NewForCoordinates(repository string, coords *maven.Coordinates, opts ...Option) *AccessSpec {
	optionutils.ApplyOptions(coords, opts...)
	accessSpec := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		RepoUrl:             repository,
		Coordinates:         *coords,
	}
	return accessSpec
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("Maven package '%s' in repository '%s' path '%s'", a.Coordinates.String(), a.RepoUrl, a.Coordinates.FilePath())
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

// GetReferenceHint returns the reference hint for the Maven (mvn) artifact.
func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	if a.IsPackage() {
		return a.GAV()
	}
	return ""
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(cv accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	octx := cv.GetContext()

	repo, err := maven.NewUrlRepository(a.RepoUrl, vfsattr.Get(cv.GetContext()))
	if err != nil {
		return nil, err
	}

	factory := func() (blobaccess.BlobAccess, error) {
		return mavenblob.BlobAccessForCoords(repo, &a.Coordinates,
			mavenblob.WithCredentialContext(octx),
			mavenblob.WithLoggingContext(octx),
			mavenblob.WithCachingFileSystem(vfsattr.Get(octx)))
	}
	return accspeccpi.AccessMethodForImplementation(accspeccpi.NewDefaultMethodImpl(cv, a, "", a.MimeType(), factory), nil)
}

func (a *AccessSpec) BaseUrl() string {
	return a.RepoUrl + "/" + a.GavPath()
}

func (a *AccessSpec) ArtifactUrl() string {
	repo, err := maven.NewUrlRepository(a.RepoUrl)
	if err != nil {
		return ""
	}
	return a.Location(repo).String()
}

func (a *AccessSpec) GetCoordinates() *maven.Coordinates {
	return a.Coordinates.Copy()
}
