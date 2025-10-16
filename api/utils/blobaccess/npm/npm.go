package npm

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/tech/npm"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
)

type PackageSpec struct {
	// registry is the base URL of the NPM registry
	registry string
	// pkg is the name of NPM package
	pkg string
	// version of the NPM package.
	version string

	options *Options
}

// NewPackageSpec creates a new NPM registry access PackageSpec version v1.
func NewPackageSpec(registry, pkg, version string, opts ...Option) (*PackageSpec, error) {
	if registry == "" {
		return nil, errors.ErrRequired("registry")
	}
	if pkg == "" {
		return nil, errors.ErrRequired("package")
	}
	if version == "" {
		return nil, errors.ErrRequired("version")
	}
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	return &PackageSpec{
		registry: registry,
		pkg:      pkg,
		version:  version,
		options:  &eff,
	}, nil
}

// PackageUrl returns the URL of the NPM package (registry/Package).
func (a *PackageSpec) PackageUrl() string {
	return strings.TrimSuffix(a.registry, "/") + path.Join("/", a.pkg)
}

// PackageVersionUrl returns the URL of the NPM package-version (registry/Package/Version).
func (a *PackageSpec) PackageVersionUrl() string {
	return strings.TrimSuffix(a.registry, "/") + path.Join("/", a.pkg, a.version)
}

func (a *PackageSpec) GetPackageVersion() (*npm.Version, error) {
	log := a.options.Logger("registry", a.registry)
	log.Debug("query index for NPM registry")
	r, err := a.reader()
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
		v, ok := project.Version[a.version] // and pick only the specified version
		if !ok {
			return nil, errors.Newf("version '%s' doesn't exist", a.version)
		}
		version = v
	}
	log.Debug("found NPM package", "name", version.Name, "version", version.Version, "dist", version.Dist)
	return &version, nil
}

func (a *PackageSpec) GetBlobAccess() (blobaccess.BlobAccess, error) {
	meta, err := a.GetPackageVersion()
	if err != nil {
		return nil, err
	}

	f := func() (io.ReadCloser, error) {
		return a.reader(meta.Dist.Tarball)
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
	return accessobj.CachedBlobAccessForWriterWithCache(a.options.Cache(), mime.MIME_TGZ, accessio.NewDataAccessWriter(acc)), nil
}

func (a *PackageSpec) reader(tar ...string) (io.ReadCloser, error) {
	url := a.PackageVersionUrl()
	if len(tar) > 0 {
		url = tar[0]
	}
	if strings.HasPrefix(url, "file://") {
		path := url[7:]
		return a.options.FileSystem().OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	creds, err := a.options.GetCredentials(a.registry, a.pkg)
	if err != nil {
		return nil, err
	}
	log := a.options.Logger("url", url)

	log.Debug("query NPM registry")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	err = npm.BasicAuthForCreds(req, creds)
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
		err = npm.BasicAuthForCreds(req, creds)
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
	log.Debug("found content in NPM registry", "size", len(content))
	return io.NopCloser(bytes.NewBuffer(content)), nil
}
