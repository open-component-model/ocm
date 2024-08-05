package dirtree

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	ocmlog "ocm.software/ocm/api/utils/logging"
)

const (
	DIR  = "tree"
	FILE = "blob"
)

const (
	ModeSym  Mode = ModeBlob | 1<<(15-2)
	ModeDir  Mode = 1 << (15 - 1)
	ModeBlob Mode = 1 << (15)
)

// LogRealm is the realm used for logging output of this package.
var LogRealm = logging.Package()

// LogContext is the default logging content used by dirtree functions.
// It uses its package location as message context and is based on the ocm
// logging context.
var LogContext = ocmlog.Context().WithContext(LogRealm)

type Mode = uint32

func FileMode(m vfs.FileMode) Mode {
	return Mode((m & 0o111) | (0o644))
}

type Node interface {
	Type() string
	Hash() []byte
	Mode() Mode
	Completed() bool

	Complete()
}

type FileNode struct {
	mode Mode
	hash []byte
}

func (n *FileNode) Type() string {
	return FILE
}

func (n *FileNode) Hash() []byte {
	return n.hash
}

func (n *FileNode) Mode() Mode {
	return n.mode
}

func (n *FileNode) Completed() bool {
	return true
}

func (n *FileNode) Complete() {}

type fileNode = FileNode

type DirNode struct {
	ctx Context
	fileNode
	completed bool
	content   map[string]Node
}

func (d *DirNode) Completed() bool {
	return d.completed
}

func (d *DirNode) AddNode(name string, n Node) error {
	if d.completed {
		return errors.ErrClosed()
	}
	if d.content == nil {
		d.content = map[string]Node{}
	}
	if d.content[name] != nil {
		return errors.ErrAlreadyExists("entry", name)
	}
	LogContext.Logger().Trace("add node", "name", name, "type", n.Type())
	d.content[name] = n
	return nil
}

func (d *DirNode) Context() Context {
	return d.ctx
}

func (d *DirNode) Complete() {
	if d.Completed() {
		return
	}
	d.completed = true
	if len(d.content) == 0 {
		return
	}
	names := []string{}
	for k, n := range d.content {
		n.Complete()
		names = append(names, k)
	}
	sort.Strings(names)
	doc := &bytes.Buffer{}
	log := d.ctx.Logger()
	log.Trace("complete dir hash", "mode", fmt.Sprintf("%o", d.Mode()), "entries", strings.Join(names, ", "))
	for _, name := range names {
		n := d.content[name]
		if n.Hash() == nil {
			continue
		}
		log.Trace("entry", "name", name, "mode", fmt.Sprintf("%o", n.Mode()), "hash", hex.EncodeToString(n.Hash()))
		fmt.Fprintf(doc, "%o %s\000", n.Mode(), name)
		doc.Write(n.Hash())
	}

	hash, err := hashIt(d.ctx, "tree", int64(doc.Len()), bytes.NewReader(doc.Bytes()))
	if err != nil {
		panic(err)
	}
	d.hash = hash
	d.content = nil
}

func hashIt(ctx Context, typ string, size int64, r io.Reader) ([]byte, error) {
	hash := ctx.Hasher()
	var w io.Writer = hash
	var buf *bytes.Buffer

	if size < 300 {
		buf = &bytes.Buffer{}
		w = buf
	}

	err := ctx.WriteHeader(w, typ, size)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return nil, err
	}
	if buf != nil {
		_, err := hash.Write(buf.Bytes())
		if err != nil {
			return nil, err
		}
	}

	h := hash.Sum(nil)
	if buf != nil {
		ctx.Logger().Trace("hash", "type", typ, "size", size, "eff", len(buf.Bytes()), "hash", hex.EncodeToString(h), "content", hex.EncodeToString(buf.Bytes()))
	} else {
		ctx.Logger().Trace("hash", "type", typ, "size", size, "hash", hex.EncodeToString(h))
	}

	return h, nil
}

func NewFileNode(ctx Context, mode vfs.FileMode, size int64, r io.Reader) (*FileNode, error) {
	log := ctx.Logger()
	m := ctx.FileMode(mode)
	log.Trace("file hash", "mode", fmt.Sprintf("%o", m), "size", size)

	hash, err := hashIt(ctx, "blob", size, r)
	if err != nil {
		return nil, err
	}
	return &FileNode{
		mode: m,
		hash: hash,
	}, nil
}

func NewLinkNode(ctx Context, link string) (*FileNode, error) {
	log := ctx.Logger()
	log.Trace("link hash", "mode", fmt.Sprintf("%o", ModeSym), "size", len(link))

	hash, err := hashIt(ctx, "blob", int64(len(link)), bytes.NewReader([]byte(link)))
	if err != nil {
		return nil, err
	}
	return &FileNode{
		mode: ModeSym,
		hash: hash,
	}, nil
}

func NewDirNode(ctx Context) *DirNode {
	return &DirNode{
		fileNode: FileNode{
			mode: ModeDir,
		},
		ctx:       ctx,
		completed: false,
		content:   map[string]Node{},
	}
}
