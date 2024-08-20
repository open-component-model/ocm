package git

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils/accessobj"
)

var worktreeBranch = plumbing.NewBranchReferenceName("ocm")

type client struct {
	vfs vfs.FileSystem
	*gurl

	auth AuthMethod
}

type Client interface {
	Repository(ctx context.Context) (*git.Repository, error)
	Refresh(ctx context.Context) error
	Update(ctx context.Context, msg string, push bool) error
	accessobj.Setup
	accessobj.Closer
}

type ClientOptions struct {
}

var _ Client = &client{}

func NewClient(url string, _ ClientOptions) (Client, error) {
	gitURL, err := decodeGitURL(url)
	if err != nil {
		return nil, err
	}

	return &client{
		vfs:  memoryfs.New(),
		gurl: gitURL,
	}, nil
}

func (c *client) Repository(ctx context.Context) (*git.Repository, error) {
	billy := VFSBillyFS(c.vfs)

	strg, err := getStorage(billy)
	if err != nil {
		return nil, err
	}

	newRepo := false
	repo, err := git.Open(strg, billy)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.CloneContext(ctx, strg, billy, &git.CloneOptions{
			Auth:          c.auth,
			URL:           c.url.String(),
			RemoteName:    git.DefaultRemoteName,
			ReferenceName: c.ref,
			SingleBranch:  true,
			Depth:         0,
			Tags:          git.AllTags,
		})
		newRepo = true
	}
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return git.Open(strg, billy)
	}

	if err != nil {
		return nil, err
	}
	if newRepo {
		if err := repo.FetchContext(ctx, &git.FetchOptions{
			Auth:       c.auth,
			RemoteName: git.DefaultRemoteName,
			Depth:      0,
			Tags:       git.AllTags,
			Force:      false,
		}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil, err
		}
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: worktreeBranch,
			Create: true,
			Keep:   true,
		}); err != nil {
			return nil, err
		}

		if err := worktree.AddGlob("*"); err != nil {
			return nil, err
		}

		if _, err := worktree.Commit("OCM Repository Setup", &git.CommitOptions{}); err != nil && !errors.Is(err, git.ErrEmptyCommit) {
			return nil, err
		}
	}

	return repo, nil
}

func getStorage(base billy.Filesystem) (storage.Storer, error) {
	dotGit, err := base.Chroot(git.GitDirName)
	if err != nil {
		return nil, err
	}

	return filesystem.NewStorage(
		dotGit,
		cache.NewObjectLRUDefault(),
	), nil
}

func (c *client) TopLevelDirs(ctx context.Context) ([]os.FileInfo, error) {
	repo, err := c.Repository(ctx)
	if err != nil {
		return nil, err
	}

	fs, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	return fs.Filesystem.ReadDir(".")
}

func (c *client) Refresh(ctx context.Context) error {
	repo, err := c.Repository(ctx)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.PullContext(ctx, &git.PullOptions{
		Auth:       c.auth,
		RemoteName: git.DefaultRemoteName,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) && !errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return err
	}

	return nil
}

func (c *client) Update(ctx context.Context, msg string, push bool) error {
	repo, err := c.Repository(ctx)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.AddGlob("*")

	if err != nil {
		return err
	}

	_, err = worktree.Commit(msg, &git.CommitOptions{})

	if err != nil {
		return err
	}

	if !push {
		return nil
	}

	if err := repo.PushContext(ctx, &git.PushOptions{
		RemoteName: git.DefaultRemoteName,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

func (c *client) Setup(system vfs.FileSystem) error {

	c.vfs = system

	if _, err := c.Repository(context.Background()); err != nil {
		return fmt.Errorf("failed to setup repository %q: %w", c.url.String(), err)
	}
	return nil
}

func (c *client) Close(_ *accessobj.AccessObject) error {
	return c.Update(context.Background(), "OCM Repository Update", true)
}
