package s3

import (
	"fmt"

	. "github.com/mandelsoft/goutils/exception"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/s3/identity"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessio/downloader"
	"ocm.software/ocm/api/utils/accessio/downloader/s3"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type of S3 registry.
const (
	Type = "s3"

	LegacyType   = "S3"
	LegacyTypeV1 = LegacyType + runtime.VersionSeparator + "v1"
)

var versions = accspeccpi.NewAccessTypeVersionScheme(Type).WithKindAliases(LegacyType)

var formats = accspeccpi.NewAccessSpecFormatVersionRegistry()

func init() {
	formats.Register(Type, runtime.NewConvertedVersion[accspeccpi.AccessSpec, *AccessSpec, *AccessSpecV1](&converterV1{}))
	formats.Register(LegacyType, runtime.NewConvertedVersion[accspeccpi.AccessSpec, *AccessSpec, *AccessSpecV1](&converterV1{}))

	initV1()
	initV2()

	anon := accspeccpi.MustNewAccessSpecMultiFormatVersion(Type, formats)
	Must(versions.Register(accspeccpi.NewAccessSpecTypeByFormatVersion(Type, anon, accspeccpi.WithDescription(usage), accspeccpi.WithConfigHandler(ConfigHandler()))))
	Must(versions.Register(accspeccpi.NewAccessSpecTypeByFormatVersion(LegacyType, anon)))
	accspeccpi.RegisterAccessTypeVersions(versions)
}

// AccessSpec describes the access for a S3 registry.
type AccessSpec struct {
	runtime.InternalVersionedTypedObject[accspeccpi.AccessSpec]

	// Region needs to be set even though buckets are global.
	// We can't assume that there is a default region setting sitting somewhere.
	// +optional
	Region string
	// Bucket where the s3 object is located.
	Bucket string
	// Key of the object to look for. This value will be used together with Bucket and Version to form an identity.
	Key string
	// Version of the object.
	// +optional
	Version string
	// MediaType defines the mime type of the object to download.
	// +optional
	MediaType  string
	downloader downloader.Downloader
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new GitHub registry access spec version v1.
func New(region, bucket, key, version, mediaType string, downloader ...downloader.Downloader) *AccessSpec {
	return &AccessSpec{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[accspeccpi.AccessSpec](versions, Type),
		Region:                       region,
		Bucket:                       bucket,
		Key:                          key,
		Version:                      version,
		MediaType:                    mediaType,
		downloader:                   utils.Optional(downloader...),
	}
}

func (a AccessSpec) MarshalJSON() ([]byte, error) {
	return runtime.MarshalVersionedTypedObject(&a)
}

func (a *AccessSpec) Describe(ctx accspeccpi.Context) string {
	return fmt.Sprintf("S3 key %s in bucket %s", a.Key, a.Bucket)
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(ctx accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (a *AccessSpec) AccessMethod(c accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newMethod(c, a))
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	blobaccess.BlobAccess

	comp accspeccpi.ComponentVersionAccess
	spec *AccessSpec
}

var _ accspeccpi.AccessMethodImpl = (*accessMethod)(nil)

func newMethod(c accspeccpi.ComponentVersionAccess, a *AccessSpec) (*accessMethod, error) {
	creds, err := getCreds(a, c.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get creds: %w", err)
	}

	var (
		accessKeyID  string
		accessSecret string
	)
	if creds != nil {
		accessKeyID = creds.GetProperty(identity.ATTR_AWS_ACCESS_KEY_ID)
		accessSecret = creds.GetProperty(identity.ATTR_AWS_SECRET_ACCESS_KEY)
	}
	var awsCreds *s3.AWSCreds
	if accessKeyID != "" {
		awsCreds = &s3.AWSCreds{
			AccessKeyID:  accessKeyID,
			AccessSecret: accessSecret,
		}
	}
	d := a.downloader
	if d == nil {
		d = s3.NewDownloader(a.Region, a.Bucket, a.Key, a.Version, awsCreds)
	}
	w := accessio.NewWriteAtWriter(d.Download)
	// don't change the spec, leave it empty.
	mediaType := a.MediaType
	if mediaType == "" {
		mediaType = mime.MIME_OCTET
	}
	cacheBlobAccess := accessobj.CachedBlobAccessForWriter(c.GetContext(), mediaType, w)
	return &accessMethod{
		spec:       a,
		comp:       c,
		BlobAccess: cacheBlobAccess,
	}, nil
}

func getCreds(a *AccessSpec, cctx credentials.Context) (credentials.Credentials, error) {
	return identity.GetCredentials(cctx, "", a.Bucket, a.Key, a.Version)
}

func (_ *accessMethod) IsLocal() bool {
	return false
}

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *accessMethod) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	return identity.GetConsumerId("", m.spec.Bucket, m.spec.Key, m.spec.Version)
}

func (m *accessMethod) GetIdentityMatcher() string {
	return hostpath.IDENTITY_TYPE
}
