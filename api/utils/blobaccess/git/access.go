package git

import (
	"context"
	"fmt"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
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
// tar.gz containing the checked-out files. WithFullHistory also includes Git history.
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
	filter := func(s string) bool { return s != gogit.GitDirName }
	if options.FullHistory {
		head, err := repo.Head()
		if err != nil {
			return nil, fmt.Errorf("cannot resolve git HEAD: %w", err)
		}
		storage, err := writeGitStorage(repo, repositoryFS, head.Hash())
		if err != nil {
			return nil, err
		}
		filter = func(s string) bool {
			if s != gogit.GitDirName && !strings.HasPrefix(s, gogit.GitDirName+"/") {
				return true
			}
			return storage[s]
		}
	}
	filteredRepositoryFS := &filteredVFS{FileSystem: repositoryFS, filter: filter}

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

// repositoryConfig is the repository configuration OCM v2 archives: the defaults
// of a freshly initialized go-git repository. It is pinned instead of taken from
// the clone, whose configuration names the source remote, and must stay
// byte-identical with v2 (see writeGitStorage).
const repositoryConfig = "[core]\n\tbare = false\n\trepositoryformatversion = 1\n\tfilemode = true\n"

// writeGitStorage prepares the Git storage of the archive and returns the entries
// it consists of, which is what the archive carries of the clone's .git directory:
// a detached HEAD, the repository configuration, and exactly one packfile with its
// index and reverse index, holding precisely the objects reachable from commit.
// Everything else the clone keeps - the source remote, refs, the index, logs and
// the fetched packs - stays out.
//
// The entries, their contents and their order are those of an archive built by OCM
// v2, so that both implementations digest the same bytes for the same commit.
func writeGitStorage(repo *gogit.Repository, fs vfs.FileSystem, commit plumbing.Hash) (map[string]bool, error) {
	pack, err := writePackfile(repo, commit)
	if err != nil {
		return nil, err
	}
	if err := writeRevIndexFor(fs, pack); err != nil {
		return nil, err
	}
	if err := vfs.WriteFile(fs, ".git/HEAD", []byte(commit.String()+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("cannot write detached git HEAD: %w", err)
	}
	if err := vfs.WriteFile(fs, ".git/config", []byte(repositoryConfig), 0o644); err != nil {
		return nil, fmt.Errorf("cannot write git configuration: %w", err)
	}

	// Directories the layout contains empty; a clone need not have them.
	directories := []string{".git/objects/info", ".git/refs/heads", ".git/refs/tags"}
	for _, dir := range directories {
		if err := fs.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("cannot create git storage directory %q: %w", dir, err)
		}
	}

	entries := map[string]bool{}
	for _, entry := range append(directories,
		".git", ".git/HEAD", ".git/config", ".git/objects", ".git/objects/pack", ".git/refs",
		pack+".pack", pack+".idx", pack+".rev",
	) {
		entries[entry] = true
	}
	return entries, nil
}

// writePackfile stores the objects reachable from commit as a single packfile in
// the repository and returns its path without extension. Objects are packed in
// hash order, as OCM v2 and git itself do.
func writePackfile(repo *gogit.Repository, commit plumbing.Hash) (string, error) {
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
		return "", errors.Join(fmt.Errorf("cannot encode git packfile: %w", err), pack.Close())
	}
	if err := pack.Close(); err != nil {
		return "", fmt.Errorf("cannot finish git packfile: %w", err)
	}
	return ".git/objects/pack/pack-" + checksum.String(), nil
}

// writeRevIndexFor writes the reverse index of an already stored packfile, which
// go-git v5 does not maintain, from the index go-git wrote for it.
func writeRevIndexFor(fs vfs.FileSystem, pack string) error {
	index, err := fs.Open(pack + ".idx")
	if err != nil {
		return fmt.Errorf("cannot read git packfile index: %w", err)
	}
	defer func() { _ = index.Close() }()

	decoded := idxfile.NewMemoryIndex()
	if err := idxfile.NewDecoder(index).Decode(decoded); err != nil {
		return fmt.Errorf("cannot decode git packfile index: %w", err)
	}

	rev, err := fs.OpenFile(pack+".rev", vfs.O_WRONLY|vfs.O_CREATE|vfs.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("cannot create git reverse index: %w", err)
	}
	if err := writeRevIndex(rev, decoded); err != nil {
		return errors.Join(err, rev.Close())
	}
	if err := rev.Close(); err != nil {
		return fmt.Errorf("cannot finish git reverse index: %w", err)
	}
	return nil
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
