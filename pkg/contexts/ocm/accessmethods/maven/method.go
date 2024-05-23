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

// Type is the access type of Maven (mvn) repository.
const (
	Type   = "maven"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a Maven (mvn) artifact.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Repository is the base URL of the Maven (mvn) repository.
	Repository string `json:"repository"`

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

// New creates a new Maven (mvn) repository access spec version v1.
func New(repository, groupId, artifactId, version string, opts ...Option) *AccessSpec {
	accessSpec := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Repository:          repository,
		Coordinates:         *maven.NewCoordinates(groupId, artifactId, version, opts...),
	}
	return accessSpec
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("Maven (mvn) package '%s' in repository '%s' path '%s'", a.Coordinates.String(), a.Repository, a.Coordinates.FilePath())
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

// GetReferenceHint returns the reference hint for the Maven (mvn) artifact.
func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.String()
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(cv accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	octx := cv.GetContext()
	factory := func() (blobaccess.BlobAccess, error) {
		return mavenblob.BlobAccessForMavenCoords(a.Repository, &a.Coordinates,
			mavenblob.WithCredentialContext(octx),
			mavenblob.WithLoggingContext(octx),
			mavenblob.WithFileSystem(vfsattr.Get(octx)))
	}
	return accspeccpi.AccessMethodForImplementation(accspeccpi.NewDefaultMethodImpl(cv, a, "", a.MimeType(), factory), nil)
}

func (a *AccessSpec) GetInexpensiveContentVersionIdentity(cv accspeccpi.ComponentVersionAccess) string {
	creds, err := identity.GetCredentials(cv.GetContext(), a.Repository, a.GroupId)
	if err != nil {
		return ""
	}
	mvncreds := mavenblob.MapCredentials(creds)
	fs := vfsattr.Get(cv.GetContext())
	files, err := maven.GavFiles(a.Repository, &a.Coordinates, mvncreds, fs)
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
		id, _ := maven.GetHash(a.Url(a.Repository), mvncreds, h, fs)
		return id
	}
	return ""
}

func (a *AccessSpec) BaseUrl() string {
	return a.Repository + "/" + a.GavPath()
}

func (a *AccessSpec) ArtifactUrl() string {
	return a.Url(a.Repository)
}

func (a *AccessSpec) GetCoordinates() *maven.Coordinates {
	return a.Coordinates.Copy()
}
