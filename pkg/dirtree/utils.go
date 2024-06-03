package dirtree

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

func createFile(d *DirNode, p string, mode vfs.FileMode, size int64, r io.Reader) (*FileNode, error) {
	dir := path.Dir(p)
	name := path.Base(path.Clean(p))
	d, err := lookupDir(d, dir, true)
	if err != nil {
		return nil, err
	}
	n, err := NewFileNode(d.ctx, mode, size, r)
	if err == nil {
		d.AddNode(name, n)
	}
	return n, err
}

func createLink(d *DirNode, p string, target string) (*FileNode, error) {
	dir := path.Dir(p)
	name := path.Base(path.Clean(p))
	d, err := lookupDir(d, dir, true)
	if err != nil {
		return nil, err
	}
	n, err := NewLinkNode(d.ctx, target)
	if err == nil {
		d.AddNode(name, n)
	}
	return n, err
}

func lookupFile(d *DirNode, p string) Node {
	p = path.Clean(p)
	if path.IsAbs(p) {
		p = p[1:]
	}
	comps := strings.Split(p, "/")
	for i, c := range comps {
		n := d.content[c]
		if n == nil || i == len(comps)-1 {
			return n
		}
		if dd, ok := n.(*DirNode); ok {
			d = dd
		}
	}
	return nil
}

func lookupDir(d *DirNode, p string, create bool) (*DirNode, error) {
	p = path.Clean(p)
	if path.IsAbs(p) {
		p = p[1:]
	}
	if p == "." {
		return d, nil
	}
	for _, c := range strings.Split(p, "/") {
		n := d.content[c]
		if n == nil {
			n = NewDirNode(d.ctx)
			d.content[c] = n
		}
		if dd, ok := n.(*DirNode); ok {
			d = dd
		} else {
			return nil, fmt.Errorf("component %q in %q is no dir", c, p)
		}
	}
	return d, nil
}
