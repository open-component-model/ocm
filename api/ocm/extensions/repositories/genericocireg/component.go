package genericocireg

import (
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/refmgmt"
)

const META_SEPARATOR = ".build-"

////////////////////////////////////////////////////////////////////////////////

type componentAccessImpl struct {
	bridge    repocpi.ComponentAccessBridge
	repo      *RepositoryImpl
	name      string
	namespace oci.NamespaceAccess
}

func newComponentAccess(repo *RepositoryImpl, name string, main bool) (*repocpi.ComponentAccessInfo, error) {
	mapped, err := repo.MapComponentNameToNamespace(name)
	if err != nil {
		return nil, err
	}
	namespace, err := repo.ocirepo.LookupNamespace(mapped)
	if err != nil {
		return nil, err
	}
	impl := &componentAccessImpl{
		repo:      repo,
		name:      name,
		namespace: namespace,
	}
	return &repocpi.ComponentAccessInfo{impl, "OCM component[OCI]", main}, nil
}

func (c *componentAccessImpl) Close() error {
	refmgmt.AllocLog.Trace("closing component [OCI]", "name", c.name)
	err := c.namespace.Close()
	refmgmt.AllocLog.Trace("closed component [OCI]", "name", c.name)
	return err
}

func (c *componentAccessImpl) SetBridge(base repocpi.ComponentAccessBridge) {
	c.bridge = base
}

func (c *componentAccessImpl) GetParentBridge() repocpi.RepositoryViewManager {
	return c.repo.bridge
}

func (c *componentAccessImpl) GetContext() cpi.Context {
	return c.repo.GetContext()
}

func (c *componentAccessImpl) GetName() string {
	return c.name
}

////////////////////////////////////////////////////////////////////////////////

func toTag(v string) (string, error) {
	_, err := semver.NewVersion(v)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(v, "+", META_SEPARATOR), nil
}

func toVersion(t string) string {
	next := 0
	for {
		if idx := strings.Index(t[next:], META_SEPARATOR); idx >= 0 {
			next += idx + len(META_SEPARATOR)
		} else {
			break
		}
	}
	if next == 0 {
		return t
	}
	return t[:next-len(META_SEPARATOR)] + "+" + t[next:]
}

func (c *componentAccessImpl) IsReadOnly() bool {
	return c.repo.IsReadOnly()
}

func (c *componentAccessImpl) ListVersions() ([]string, error) {
	tags, err := c.namespace.ListTags()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		// omit reported digests (typically for ctf)
		if ok, _ := artdesc.IsDigest(t); !ok {
			result = append(result, toVersion(t))
		}
	}
	return result, err
}

func (c *componentAccessImpl) HasVersion(vers string) (bool, error) {
	tags, err := c.namespace.ListTags()
	if err != nil {
		return false, err
	}
	for _, t := range tags {
		// omit reported digests (typically for ctf)
		if ok, _ := artdesc.IsDigest(t); !ok {
			if vers == t {
				return true, nil
			}
		}
	}
	return false, err
}

func (c *componentAccessImpl) LookupVersion(version string) (*repocpi.ComponentVersionAccessInfo, error) {
	tag, err := toTag(version)
	if err != nil {
		return nil, err
	}
	acc, err := c.namespace.GetArtifact(tag)
	if err != nil {
		if errors.IsErrNotFound(err) {
			return nil, cpi.ErrComponentVersionNotFoundWrap(err, c.name, version)
		}
		return nil, err
	}
	m := accessobj.ACC_WRITABLE
	if c.IsReadOnly() {
		m = accessobj.ACC_READONLY
	}
	return newComponentVersionAccess(m, c, version, acc, true)
}

func (c *componentAccessImpl) NewVersion(version string, overrides ...bool) (*repocpi.ComponentVersionAccessInfo, error) {
	if c.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	override := general.Optional(overrides...)
	tag, err := toTag(version)
	if err != nil {
		return nil, err
	}
	acc, err := c.namespace.GetArtifact(tag)
	if err == nil {
		if override {
			return newComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc, false)
		}
		return nil, errors.ErrAlreadyExists(cpi.KIND_COMPONENTVERSION, c.name+"/"+version)
	}
	if !errors.IsErrNotFoundKind(err, oci.KIND_OCIARTIFACT) {
		return nil, err
	}
	acc, err = c.namespace.NewArtifact()
	if err != nil {
		return nil, err
	}
	return newComponentVersionAccess(accessobj.ACC_CREATE, c, version, acc, false)
}
