package signing

import (
	"io"
	"os"
	"sync"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

// VerifiedStore is an interface for some kind of
// memory providing information about verified
// component versions and the digests of the verified
// artifacts. It is used to verify downloaded resource content,
// without requiring to verify the complete component version, again.
// If the component version has already been marked as being verified
// only the digest of the downloaded content has be compared with the
// digest already marked as verified in the context of its component version.
//
// A typical implementation is a file based store, which stored the serialized
// component versions (see NewVerifiedStore).
type VerifiedStore interface {
	Add(cd *compdesc.ComponentDescriptor, signatures ...string)
	Remove(n common.VersionedElement)
	Get(n common.VersionedElement) *compdesc.ComponentDescriptor
	GetEntry(n common.VersionedElement) *StorageEntry

	GetResourceDigest(n common.VersionedElement, id metav1.Identity) *metav1.DigestSpec
	GetResourceDigestByIndex(n common.VersionedElement, idx int) *metav1.DigestSpec

	Entries() []common.NameVersion

	Load() error
	Save() error
}

type verifiedStore struct {
	lock    sync.Mutex
	storage *StorageDescriptor
	fs      vfs.FileSystem
	file    string
}

var _ VerifiedStore = (*verifiedStore)(nil)

// NewLocalVerifiedStore creates a memory based VerifiedStore.
func NewLocalVerifiedStore() VerifiedStore {
	return &verifiedStore{storage: &StorageDescriptor{}}
}

// NewVerifiedStore loads or creates a new filesystem based VerifiedStore.
func NewVerifiedStore(path string, fss ...vfs.FileSystem) (VerifiedStore, error) {
	eff, err := utils.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	fs := utils.FileSystem(fss...)

	s := &verifiedStore{
		fs:   fs,
		file: eff,
	}

	err = s.Load()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (v *verifiedStore) Load() error {
	v.lock.Lock()
	defer v.lock.Unlock()
	if v.fs == nil {
		return nil
	}

	dir := filepath.Dir(v.file)

	if ok, err := vfs.DirExists(v.fs, dir); !ok || err != nil {
		if err != nil {
			return err
		}
		return errors.ErrNotFound("directory", dir)
	}

	var storage StorageDescriptor
	f, err := v.fs.Open(v.file)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		err = runtime.DefaultYAMLEncoding.Unmarshal(data, &storage)
		if err != nil {
			return err
		}
	}
	v.storage = &storage
	return nil
}

func (v *verifiedStore) Save() error {
	v.lock.Lock()
	defer v.lock.Unlock()
	if v.fs == nil {
		return nil
	}

	data, err := runtime.DefaultYAMLEncoding.Marshal(v.storage)
	if err != nil {
		return err
	}

	err = vfs.WriteFile(v.fs, v.file, data, 0o600)
	if err != nil {
		return err
	}
	return nil
}

func (v *verifiedStore) Entries() []common.NameVersion {
	v.lock.Lock()
	defer v.lock.Unlock()

	entries := make([]common.NameVersion, 0, len(v.storage.ComponentVersions))
	for _, entry := range v.storage.ComponentVersions {
		entries = append(entries, common.VersionedElementKey(entry.Descriptor))
	}
	return entries
}

func (v *verifiedStore) Add(cd *compdesc.ComponentDescriptor, signatures ...string) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.storage == nil {
		v.storage = &StorageDescriptor{}
	}
	if v.storage.ComponentVersions == nil {
		v.storage.ComponentVersions = map[string]*StorageEntry{}
	}
	key := common.VersionedElementKey(cd).String()
	old := v.storage.ComponentVersions[key]
	if old == nil || !old.Descriptor.Descriptor().Equal(cd) {
		old = &StorageEntry{
			Descriptor: (*compdesc.GenericComponentDescriptor)(cd),
		}
	}
	for _, e := range signatures {
		old.Signatures = sliceutils.AppendUnique(old.Signatures, e)
	}
	v.storage.ComponentVersions[key] = old
}

func (v *verifiedStore) Remove(n common.VersionedElement) {
	v.lock.Lock()
	defer v.lock.Unlock()

	delete(v.storage.ComponentVersions, common.VersionedElementKey(n).String())
}

func (v *verifiedStore) GetEntry(n common.VersionedElement) *StorageEntry {
	v.lock.Lock()
	defer v.lock.Unlock()

	return v.storage.ComponentVersions[common.VersionedElementKey(n).String()]
}

func (v *verifiedStore) Get(n common.VersionedElement) *compdesc.ComponentDescriptor {
	v.lock.Lock()
	defer v.lock.Unlock()

	entry := v.storage.ComponentVersions[common.VersionedElementKey(n).String()]
	if entry == nil {
		return nil
	}
	return entry.Descriptor.Descriptor()
}

func (v *verifiedStore) GetResourceDigest(n common.VersionedElement, id metav1.Identity) *metav1.DigestSpec {
	cd := v.Get(n)
	if cd == nil {
		return nil
	}
	r, err := cd.GetResourceByIdentity(id)
	if err != nil {
		return nil
	}
	return r.Digest
}

func (v *verifiedStore) GetResourceDigestByIndex(n common.VersionedElement, idx int) *metav1.DigestSpec {
	cd := v.Get(n)
	if cd == nil {
		return nil
	}
	if idx < 0 || idx >= len(cd.Resources) {
		return nil
	}
	return cd.Resources[idx].Digest
}

type StorageEntry struct {
	Signatures []string                             `json:"signatures,omitempty"`
	Descriptor *compdesc.GenericComponentDescriptor `json:"descriptor"`
}

type StorageDescriptor struct {
	ComponentVersions map[string]*StorageEntry `json:"componentVersions,omitempty"`
}
