package git

import (
	"context"

	gogit "github.com/go-git/go-git/v5"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
)

func BlobAccess(opt ...Option) (_ bpi.BlobAccess, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	options := optionutils.EvalOptions(opt...)
	if options.URL == "" {
		return nil, errors.New("no URL specified")
	}
	log := options.Logger("RepoUrl", options.URL)

	if options.AuthMethod == nil && options.Credentials.Value != nil {
		authMethod, err := git.AuthFromCredentials(options.Credentials.Value)
		if err != nil && !errors.Is(err, git.ErrNoValidGitCredentials) {
			return nil, err
		} else {
			options.AuthMethod = authMethod
		}
	}

	c, err := git.NewClient(options.ClientOptions)
	if err != nil {
		return nil, err
	}

	tmpfs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}
	finalize.With(func() error {
		return vfs.Cleanup(tmpfs)
	})

	// redirect the client to the temporary filesystem for storage of the repo, otherwise it would use memory
	if err := c.Setup(tmpfs); err != nil {
		return nil, err
	}

	// get the repository, triggering a clone if not present into the filesystem provided by setup
	if _, err := c.Repository(context.Background()); err != nil {
		return nil, err
	}

	// remove the .git directory as it shouldn't be part of the tarball
	if err := tmpfs.RemoveAll(gogit.GitDirName); err != nil {
		return nil, err
	}

	// pack all downloaded files into a tar.gz file
	fs := options.GetCachingFileSystem()
	tgz, err := vfs.TempFile(fs, "", "git-*.tar.gz")
	if err != nil {
		return nil, err
	}

	dw := iotools.NewDigestWriterWith(digest.SHA256, tgz)
	finalize.Close(dw)

	if err := tarutils.TgzFs(tmpfs, dw); err != nil {
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
