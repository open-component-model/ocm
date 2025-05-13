package git

import (
	"compress/gzip"
	"context"

	gogit "github.com/go-git/go-git/v5"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/optionutils"
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

// BlobAccess clones the repository into a temporary filesystem, packs it into a tar.gz file,
// and returns a BlobAccess object for the tar.gz file.
func BlobAccess(opt ...Option) (_ bpi.BlobAccess, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	options := optionutils.EvalOptions(opt...)
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
	if _, err := c.Repository(context.Background()); err != nil {
		return nil, err
	}

	filteredRepositoryFS := &filteredVFS{
		FileSystem: repositoryFS,
		filter: func(s string) bool {
			return s != gogit.GitDirName
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

	zip := gzip.NewWriter(dw)

	if err := tarutils.PackFsIntoTar(filteredRepositoryFS, "", zip, tarutils.TarFileSystemOptions{}); err != nil {
		return nil, err
	}
	// Close the write to make sure that the digest writer calculates on a closed file
	if err := zip.Close(); err != nil {
		return nil, err
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
