package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"

	"ocm.software/ocm/api/utils/accessio/downloader"
)

const localRemoteName = "origin"

type CloseableDownloader interface {
	downloader.Downloader
	Close() error
}

// Downloader simply uses the default HTTP client to download the contents of a URL.
type Downloader struct {
	cloneOpts *git.CloneOptions
	grepOpts  *git.GrepOptions

	matching *regexp.Regexp

	mu      sync.Mutex
	buf     *bytes.Buffer
	storage storage.Storer
}

var _ downloader.Downloader = (*Downloader)(nil)

func NewDownloader(url string, ref string, path string) CloseableDownloader {
	refName := plumbing.ReferenceName(ref)
	return &Downloader{
		cloneOpts: &git.CloneOptions{
			URL:           url,
			RemoteName:    localRemoteName,
			ReferenceName: refName,
			SingleBranch:  true,
			Tags:          git.NoTags,
			Depth:         0,
		},
		matching: regexp.MustCompile(fmt.Sprintf(`%s`, path)),
		buf:      bytes.NewBuffer(make([]byte, 0, 4096)),
		storage:  memory.NewStorage(),
	}
}

func (d *Downloader) Download(w io.WriterAt) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	ctx := context.Background()

	// no support for git archive yet, so we need to clone the repository in bare mode
	repo, err := git.CloneContext(ctx, d.storage, nil, d.cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", d.cloneOpts.URL, err)
	}

	trees, err := repo.TreeObjects()
	if err != nil {
		return fmt.Errorf("failed to get tree objects: %w", err)
	}

	if err := trees.ForEach(func(t *object.Tree) error {
		return t.Files().ForEach(d.copyFileToBuffer)
	}); err != nil {
		return fmt.Errorf("failed to iterate over trees: %w", err)
	}

	defer d.buf.Reset()
	if _, err := w.WriteAt(d.buf.Bytes(), 0); err != nil {
		return fmt.Errorf("failed to write blobs: %w", err)
	}

	return nil
}

func (d *Downloader) copyFileToBuffer(file *object.File) error {
	if !d.matching.MatchString(file.Name) {
		return nil
	}

	reader, err := file.Reader()
	if err != nil {
		return err
	}
	_, err = io.Copy(d.buf, reader)
	return errors.Join(err, reader.Close())
}

func (d *Downloader) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.buf = nil
	d.storage = nil

	return nil
}
