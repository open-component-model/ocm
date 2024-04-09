package mvn

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"encoding/xml
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type of Maven (mvn) repository.
const (
	Type   = "mvn"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a Maven (mvn) repository.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Repository is the base URL of the Maven (mvn) repository
	Repository string `json:"repository"`
	// ArtifactId is the name of Maven (mvn) package
	GroupId string `json:"groupId"`
	// ArtifactId is the name of Maven (mvn) package
	ArtifactId string `json:"artifactId"`
	// Version of the Maven (mvn) package.
	Version string `json:"version"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new Maven (mvn) repository access spec version v1.
func New(repository, groupId, artifactId, version string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Repository:          repository,
		GroupId:             groupId,
		ArtifactId:          artifactId,
		Version:             version,
	}
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("Maven (mvn) package %s:%s:%s in repository %s", a.GroupId, a.ArtifactId, a.Version, a.Repository)
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.GroupId + ":" + a.ArtifactId + ":" + a.Version
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newMethod(c, a))
}

func (a *AccessSpec) GetInexpensiveContentVersionIdentity(access accspeccpi.ComponentVersionAccess) string {
	meta, _ := a.getPackageMeta(access.GetContext())
	if meta != nil {
		return meta.Dist.Shasum
	}
	return ""
}

func (a *AccessSpec) getPackageMeta(ctx accspeccpi.Context) (*meta, error) {
	//    <groupId>cn.afternode.commons</groupId>
	//    <artifactId>commons</artifactId>
	//    <version>1.6</version>
	// repository: https://repo1.maven.org/maven2/
	// groupId: cn/afternode/commons
	// version: 1.6
	// artifactId: commons
	url := a.Repository + path.Join("/", strings.Replace(a.GroupId, ".", "/", -1), a.ArtifactId, a.Version, a.ArtifactId+"-"+a.Version+".pom")
	// https://repo1.maven.org/maven2/cn/afternode/commons/commons/1.6/commons-1.6.pom
	r, err := reader(url, vfsattr.Get(ctx))
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, io.LimitReader(r, 200000))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version metadata for %s", url)
	}

	var metadata meta

	// read xml from buf and fill metadata
	xml.Unmarshal(buf.Bytes(), &metadata)

	err = json.Unmarshal(buf.Bytes(), &metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal version metadata for %s", url)
	}
	return &metadata, nil
}

////////////////////////////////////////////////////////////////////////////////

func newMethod(c accspeccpi.ComponentVersionAccess, a *AccessSpec) (accspeccpi.AccessMethodImpl, error) {
	factory := func() (blobaccess.BlobAccess, error) {
		meta, err := a.getPackageMeta(c.GetContext())
		if err != nil {
			return nil, err
		}

		f := func() (io.ReadCloser, error) {
			return reader(meta.Dist.Tarball, vfsattr.Get(c.GetContext()))
		}
		if meta.Dist.Shasum != "" {
			tf := f
			f = func() (io.ReadCloser, error) {
				r, err := tf()
				if err != nil {
					return nil, err
				}
				return iotools.VerifyingReaderWithHash(r, crypto.SHA1, meta.Dist.Shasum), nil
			}
		}
		acc := blobaccess.DataAccessForReaderFunction(f, meta.Dist.Tarball)
		return accessobj.CachedBlobAccessForWriter(c.GetContext(), mime.MIME_JAR, accessio.NewDataAccessWriter(acc)), nil
	}
	return accspeccpi.NewDefaultMethodImpl(c, a, "", mime.MIME_JAR, factory), nil
}

type meta struct {
	Type   string `json:"type"` // pom, jar, sources, javadoc, module, ...
	MD5    string `json:"md5"`
	Sha1   string `json:"sha1"`
	Sha256 string `json:"sha256"`
	Sha512 string `json:"sha512"`
	Asc    string `json:"asc"`
}

func reader(url string, fs vfs.FileSystem) (io.ReadCloser, error) {
	c := &http.Client{}

	if strings.HasPrefix(url, "file://") {
		path := url[7:]
		return fs.OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, io.LimitReader(resp.Body, 2000))
		if err != nil {
			return nil, errors.Newf("version meta data request %s provides %s", url, resp.Status)
		}
		return nil, errors.Newf("version meta data request %s provides %s: %s", url, resp.Status, buf.String())
	}
	return resp.Body, nil
}
