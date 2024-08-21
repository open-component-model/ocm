package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

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
	opts ClientOptions

	// vfs tracks the current filesystem where the repo will be stored (at the root)
	vfs vfs.FileSystem

	// url is the git URL of the repository
	*gurl

	// auth is the authentication method to use when accessing the repository
	auth AuthMethod

	// repo is a reference to the git repository if it is already open
	repo   *git.Repository
	repoMu sync.Mutex
}

type Client interface {
	Repository(ctx context.Context) (*git.Repository, error)
	Refresh(ctx context.Context) error
	Update(ctx context.Context, msg string, push bool) error
	accessobj.Setup
	accessobj.Closer
}

type ClientOptions struct {
	URL string
	Author
}

type Author struct {
	Name  string
	Email string
}

var _ Client = &client{}

func NewClient(opts ClientOptions) (Client, error) {
	gitURL, err := decodeGitURL(opts.URL)
	if err != nil {
		return nil, err
	}

	return &client{
		vfs:  memoryfs.New(),
		gurl: gitURL,
		opts: opts,
	}, nil
}

func (c *client) Repository(ctx context.Context) (*git.Repository, error) {
	c.repoMu.Lock()
	defer c.repoMu.Unlock()
	if c.repo != nil {
		return c.repo, nil
	}

	billyFS, err := VFSBillyFS(c.vfs)
	if err != nil {
		return nil, err
	}

	strg, err := GetStorage(billyFS)
	if err != nil {
		return nil, err
	}

	newRepo := false
	repo, err := git.Open(strg, billyFS)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.CloneContext(ctx, strg, billyFS, &git.CloneOptions{
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
		return git.Open(strg, billyFS)
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

	if err := c.opts.applyToRepo(repo); err != nil {
		return nil, err
	}

	c.repo = repo

	return repo, nil
}

func GetStorage(base billy.Filesystem) (storage.Storer, error) {
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

	if err = worktree.AddGlob("*"); err != nil {
		return err
	}

	_, err = worktree.Commit(msg, &git.CommitOptions{})

	if errors.Is(err, git.ErrEmptyCommit) {
		return nil
	}

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
	if err := c.Update(context.Background(), "OCM Repository Update", true); err != nil {
		return fmt.Errorf("failed to close repository %q: %w", c.url.String(), err)
	}
	return nil
}

func (o ClientOptions) applyToRepo(repo *git.Repository) error {
	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	if o.Author.Name != "" {
		cfg.User.Name = o.Author.Name
	}

	if o.Author.Email != "" {
		cfg.User.Email = o.Author.Email
	}

	return repo.SetConfig(cfg)
}
