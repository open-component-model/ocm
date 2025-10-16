package git

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitclient "github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	mlog "github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/utils/logging"
)

const LogAttrProtocol = "protocol"

func init() {
	// override the logging realm for http based git clients
	gitclient.InstallProtocol("http", githttp.NewClient(&http.Client{
		Transport: logging.NewRoundTripper(http.DefaultTransport, logging.DynamicLogger(REALM, mlog.NewAttribute(LogAttrProtocol, "http"))),
	}))
	gitclient.InstallProtocol("https", githttp.NewClient(&http.Client{
		Transport: logging.NewRoundTripper(http.DefaultTransport, logging.DynamicLogger(REALM, mlog.NewAttribute(LogAttrProtocol, "https"))),
	}))
	// TODO Determine how we ideally log for ssh+git protocol
}

var DefaultWorktreeBranch = plumbing.NewBranchReferenceName("ocm")

type client struct {
	opts ClientOptions

	// vfs tracks the current filesystem where the repo will be stored (at the root)
	vfs vfs.FileSystem

	// repo is a reference to the git repository if it is already open
	repo   *git.Repository
	repoMu sync.Mutex
}

// Client is a heavy abstraction over the go git Client that opinionates the remote as git.DefaultRemoteName
// as well as access to it via high level functions that are usually required for operation within OCM CTFs that are stored
// within Git. It is not general-purpose.
type Client interface {
	// Repository returns the git repository for the client initialized in the Filesystem given to Setup.
	// If Setup is not called before Repository, it will an in-memory filesystem.
	// Repository will attempt to initially clone the repository if it does not exist.
	// If the repository is already open or cloned in the filesystem, it will attempt to open & return the existing repository.
	// If the remote repository does not exist, a new repository will be created with a dummy commit and the remote
	// configured to the given URL. At that point it is up to the remote to accept an initial push to the repository or not with the
	// given AuthMethod.
	Repository(ctx context.Context) (*git.Repository, error)

	// Setup will override the current filesystem with the given filesystem. This will be the filesystem where the repository will be stored.
	// There can be only one filesystem per client.
	// If the filesystem contains a repository already, it can be consumed by a subsequent call to Repository.
	Setup(context.Context, vfs.FileSystem) error
}

type ClientOptions struct {
	// URL is the URL of the git repository to clone or open.
	URL string
	// Ref is the reference to the repository to clone or open.
	// If empty, it will default to plumbing.HEAD of the remote repository.
	// If the remote does not exist, it will attempt to push to the remote with DefaultWorktreeBranch on Client.Update.
	// To point to a remote branch, use refs/heads/<branch>.
	// To point to a tag, use refs/tags/<tag>.
	Ref string
	// Commit is the commit hash to checkout after cloning the repository.
	// If empty, it will default to the plumbing.HEAD of the Ref.
	Commit string
	// AuthMethod is the authentication method to use for the repository.
	AuthMethod AuthMethod
}

var _ Client = &client{}

func NewClient(opts ClientOptions) (Client, error) {
	pref := plumbing.HEAD
	if opts.Ref != "" {
		pref = plumbing.ReferenceName(opts.Ref)
		if err := pref.Validate(); err != nil {
			return nil, fmt.Errorf("invalid reference %q: %w", opts.Ref, err)
		}
	}

	return &client{
		vfs:  memoryfs.New(),
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

	depth := 0
	if c.opts.Commit == "" {
		depth = 1 // if we have no dedicated commit we can checkout HEAD, and thus a shallow clone is ok
	}

	newRepo := false
	repo, err := git.Open(strg, billyFS)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.CloneContext(ctx, strg, billyFS, &git.CloneOptions{
			Auth:              c.opts.AuthMethod,
			URL:               c.opts.URL,
			RemoteName:        git.DefaultRemoteName,
			ReferenceName:     plumbing.ReferenceName(c.opts.Ref),
			SingleBranch:      true,
			Depth:             depth,
			ShallowSubmodules: depth == 1,
			Tags:              git.AllTags,
		})
		newRepo = true
	}
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		repo, err = git.Open(strg, billyFS)
		if err != nil {
			return nil, fmt.Errorf("failed to open repository based on URL %q after it was determined to be an empty clone: %w", c.opts.URL, err)
		}
	}

	if err != nil {
		return nil, err
	}
	if newRepo {
		if err := c.newRepository(ctx, repo); err != nil {
			return nil, err
		}
	}

	c.repo = repo

	return repo, nil
}

func (c *client) newRepository(ctx context.Context, repo *git.Repository) error {
	if err := repo.FetchContext(ctx, &git.FetchOptions{
		Auth:       c.opts.AuthMethod,
		RemoteName: git.DefaultRemoteName,
		Depth:      0,
		Tags:       git.AllTags,
		Force:      false,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	var hash plumbing.Hash
	if c.opts.Commit != "" {
		hash = plumbing.NewHash(c.opts.Commit)
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Hash:   hash,
		Branch: DefaultWorktreeBranch,
		Create: true,
		Keep:   true,
	}); err != nil {
		return err
	}

	return nil
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

func (c *client) Setup(ctx context.Context, system vfs.FileSystem) error {
	c.vfs = system
	if _, err := c.Repository(ctx); err != nil {
		return fmt.Errorf("failed to setup repository %q: %w", c.opts.URL, err)
	}
	return nil
}
