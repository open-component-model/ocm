package maven

import (
	"fmt"

	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	mavenblob "github.com/open-component-model/ocm/pkg/blobaccess/maven"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/maven/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/runtime"
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
		return mavenblob.BlobAccessForMavenCoords(repo, &a.Coordinates,
			mavenblob.WithCredentialContext(octx),
			mavenblob.WithLoggingContext(octx),
			mavenblob.WithCachingFileSystem(vfsattr.Get(octx)))
	}
	return accspeccpi.AccessMethodForImplementation(accspeccpi.NewDefaultMethodImpl(cv, a, "", a.MimeType(), factory), nil)
}

func (a *AccessSpec) GetInexpensiveContentVersionIdentity(cv accspeccpi.ComponentVersionAccess) string {
	creds, err := identity.GetCredentials(cv.GetContext(), a.RepoUrl, a.GroupId)
	if err != nil {
		return ""
	}
	mvncreds := mavenblob.MapCredentials(creds)
	fs := vfsattr.Get(cv.GetContext())
	repo, err := maven.NewUrlRepository(a.RepoUrl, fs)
	if err != nil {
		return ""
	}
	files, err := repo.GavFiles(&a.Coordinates, mvncreds)
	if err != nil {
		return ""
	}
	files = a.Coordinates.FilterFileMap(files)
	if len(files) != 1 {
		return ""
	}
	if optionutils.AsValue(a.Extension) == "" {
		return ""
	}
	for _, h := range files {
		id, _ := a.Location(repo).GetHash(mvncreds, h)
		return id
	}
	return ""
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
