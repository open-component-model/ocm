package git

import (
	"context"
	"fmt"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/revlist"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
)

// BlobAccess clones the repository into a temporary filesystem and returns a
// tar.gz containing the checked-out files and the reachable Git history.
func BlobAccess(opt ...Option) (_ bpi.BlobAccess, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	var options Options
	for _, opt := range opt {
		opt.ApplyTo(&options)
	}
	if options.URL == "" {
		return nil, errors.New("no URL specified")
	}
	log := options.Logger("RepoUrl", options.URL)

	if err := options.ConfigureAuthMethod(); err != nil {
		return nil, err
	}

	c, err := git.NewClient(options.ClientOptions)
	if err != nil {
		return nil, err
	}

	tmpFS, cleanup, err := options.CachingFilesystem()
	if err != nil {
		return nil, err
	} else if cleanup != nil {
		finalize.With(cleanup)
	}

	// store the repo in a temporary filesystem subfolder, so the tgz can go in the root without issues.
	if err := tmpFS.MkdirAll("repository", 0o700); err != nil {
		return nil, err
	}

	repositoryFS, err := projectionfs.New(tmpFS, "repository")
	if err != nil {
		return nil, err
	}
	finalize.With(func() error {
		return tmpFS.RemoveAll("repository")
	})

	// redirect the client to the temporary filesystem for storage of the repo, otherwise it would use memory
	if err := c.Setup(context.Background(), repositoryFS); err != nil {
		return nil, err
	}

	// get the repository, triggering a clone if not present into the filesystem provided by setup
	repo, err := c.Repository(context.Background())
	if err != nil {
		return nil, err
	}
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve git HEAD: %w", err)
	}
	packName, err := writeHistory(repo, repositoryFS, head.Hash())
	if err != nil {
		return nil, err
	}
	filteredRepositoryFS := &filteredVFS{
		FileSystem: repositoryFS,
		filter: func(s string) bool {
			if s == gogit.GitDirName || s == ".git/HEAD" || s == ".git/refs" || s == ".git/objects" || s == ".git/objects/pack" {
				return true
			}
			return s == ".git/objects/pack/"+packName+".pack" || s == ".git/objects/pack/"+packName+".idx" || !strings.HasPrefix(s, gogit.GitDirName+"/")
		},
	}

	// pack all downloaded files into a tar.gz file
	fs := tmpFS
	tgz, err := vfs.TempFile(fs, "", "git-*.tar.gz")
	if err != nil {
		return nil, err
	}

	dw := iotools.NewDigestWriterWith(digest.SHA256, tgz)
	finalize.Close(dw)

	if err := tarutils.TgzFs(filteredRepositoryFS, dw, tarutils.TarFileSystemOptions{ZeroModTime: true, NormalizeHeaders: true}); err != nil {
		return nil, fmt.Errorf("failed to create tgz: %w", err)
	}

	log.Debug("created", "file", tgz.Name())

	return file.BlobAccessForTemporaryFilePath(
		mime.MIME_TGZ,
		tgz.Name(),
		file.WithFileSystem(fs),
		file.WithDigest(dw.Digest()),
		file.WithSize(dw.Size()),
	), nil
}

// writeHistory adds a pack containing precisely the objects reachable from
// HEAD. The archive filter excludes all other clone metadata and object packs.
func writeHistory(repo *gogit.Repository, fs vfs.FileSystem, commit plumbing.Hash) (string, error) {
	hashes, err := revlist.Objects(repo.Storer, []plumbing.Hash{commit}, nil)
	if err != nil {
		return "", fmt.Errorf("cannot collect git history: %w", err)
	}
	for _, hash := range hashes {
		if err := repo.Storer.HasEncodedObject(hash); err != nil {
			return "", fmt.Errorf("incomplete git history at %s: %w", hash, err)
		}
	}
	sort.Sort(plumbing.HashSlice(hashes))
	packer, ok := repo.Storer.(storer.PackfileWriter)
	if !ok {
		return "", fmt.Errorf("git storage cannot write packfiles")
	}
	pack, err := packer.PackfileWriter()
	if err != nil {
		return "", fmt.Errorf("cannot create git packfile: %w", err)
	}
	checksum, err := packfile.NewEncoder(pack, contentStorer{repo.Storer}, false).Encode(hashes, config.DefaultPackWindow)
	if err != nil {
		_ = pack.Close()
		return "", fmt.Errorf("cannot encode git packfile: %w", err)
	}
	if err := pack.Close(); err != nil {
		return "", fmt.Errorf("cannot finish git packfile: %w", err)
	}
	head, err := fs.OpenFile(".git/HEAD", vfs.O_WRONLY|vfs.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("cannot write detached git HEAD: %w", err)
	}
	if _, err := head.Write([]byte(commit.String() + "\n")); err != nil {
		_ = head.Close()
		return "", fmt.Errorf("cannot write detached git HEAD: %w", err)
	}
	if err := head.Close(); err != nil {
		return "", fmt.Errorf("cannot finish detached git HEAD: %w", err)
	}
	return "pack-" + checksum.String(), nil
}

// Hiding DeltaObjectStorer makes pack generation depend on object content,
// rather than on the source repository's existing delta layout.
type contentStorer struct{ storer.EncodedObjectStorer }

type filteredVFS struct {
	vfs.FileSystem
	filter func(string) bool
}

func (f *filteredVFS) Open(name string) (vfs.File, error) {
	if !f.filter(name) {
		return nil, vfs.SkipDir
	}
	return f.FileSystem.Open(name)
}

func (f *filteredVFS) OpenFile(name string, flags int, perm vfs.FileMode) (vfs.File, error) {
	if !f.filter(name) {
		return nil, vfs.SkipDir
	}
	return f.FileSystem.OpenFile(name, flags, perm)
}

func (f *filteredVFS) Stat(name string) (vfs.FileInfo, error) {
	if !f.filter(name) {
		return nil, vfs.SkipDir
	}
	return f.FileSystem.Stat(name)
}

func (f *filteredVFS) Lstat(name string) (vfs.FileInfo, error) {
	if !f.filter(name) {
		return nil, vfs.SkipDir
	}
	return f.FileSystem.Lstat(name)
}
