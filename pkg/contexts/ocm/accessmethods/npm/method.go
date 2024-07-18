package npm

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/blobaccess/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/npm"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type of NPM registry.
const (
	Type   = "npm"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a NPM registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Registry is the base URL of the NPM registry
	Registry string `json:"registry"`
	// Package is the name of NPM package
	Package string `json:"package"`
	// Version of the NPM package.
	Version string `json:"version"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new NPM registry access spec version v1.
func New(registry, pkg, version string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Registry:            registry,
		Package:             pkg,
		Version:             version,
	}
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("NPM package %s:%s in registry %s", a.Package, a.Version, a.Registry)
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.Package + ":" + a.Version
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newMethod(c, a))
}

// PackageUrl returns the URL of the NPM package (Registry/Package).
func (a *AccessSpec) PackageUrl() string {
	return strings.TrimSuffix(a.Registry, "/") + path.Join("/", a.Package)
}

// PackageVersionUrl returns the URL of the NPM package-version (Registry/Package/Version).
func (a *AccessSpec) PackageVersionUrl() string {
	return strings.TrimSuffix(a.Registry, "/") + path.Join("/", a.Package, a.Version)
}

func (a *AccessSpec) GetPackageVersion(ctx accspeccpi.Context) (*npm.Version, error) {
	r, err := reader(a, vfsattr.Get(ctx), ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version metadata for %s", a.PackageVersionUrl())
	}
	var version npm.Version
	err = json.Unmarshal(buf, &version)
	if err != nil || version.Dist.Tarball == "" {
		// ugly fallback as workaround for https://github.com/sonatype/nexus-public/issues/224
		var project npm.Project
		err = json.Unmarshal(buf, &project) // parse the complete project
		if err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal version metadata for %s", a.PackageVersionUrl())
		}
		v, ok := project.Version[a.Version] // and pick only the specified version
		if !ok {
			return nil, errors.Newf("version '%s' doesn't exist", a.Version)
		}
		version = v
	}
	return &version, nil
}

////////////////////////////////////////////////////////////////////////////////

func newMethod(c accspeccpi.ComponentVersionAccess, a *AccessSpec) (accspeccpi.AccessMethodImpl, error) {
	factory := func() (blobaccess.BlobAccess, error) {
		meta, err := a.GetPackageVersion(c.GetContext())
		if err != nil {
			return nil, err
		}

		f := func() (io.ReadCloser, error) {
			return reader(a, vfsattr.Get(c.GetContext()), c.GetContext(), meta.Dist.Tarball)
		}
		if meta.Dist.Integrity != "" {
			tf := f
			f = func() (io.ReadCloser, error) {
				r, err := tf()
				if err != nil {
					return nil, err
				}
				digest, err := iotools.DecodeBase64ToHex(meta.Dist.Integrity)
				if err != nil {
					return nil, err
				}
				return iotools.VerifyingReaderWithHash(r, crypto.SHA512, digest), nil
			}
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
		return accessobj.CachedBlobAccessForWriter(c.GetContext(), mime.MIME_TGZ, accessio.NewDataAccessWriter(acc)), nil
	}
	return accspeccpi.NewDefaultMethodImpl(c, a, "", mime.MIME_TGZ, factory), nil
}

func reader(a *AccessSpec, fs vfs.FileSystem, ctx cpi.ContextProvider, tar ...string) (io.ReadCloser, error) {
	url := a.PackageVersionUrl()
	if len(tar) > 0 {
		url = tar[0]
	}
	if strings.HasPrefix(url, "file://") {
		path := url[7:]
		return fs.OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	err = npm.BasicAuth(req, ctx, a.Registry, a.Package)
	if err != nil {
		return nil, err
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		// maybe it's stupid Nexus - https://github.com/sonatype/nexus-public/issues/224?
		url = a.PackageUrl()
		req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		err = npm.BasicAuth(req, ctx, a.Registry, a.Package)
		if err != nil {
			return nil, err
		}

		// close body before overwriting to close any pending connections
		resp.Body.Close()
		resp, err = c.Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
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
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewBuffer(content)), nil
}
