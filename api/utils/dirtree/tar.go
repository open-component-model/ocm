package dirtree

import (
	"archive/tar"
	"fmt"
	"io"
	"path"

	"github.com/mandelsoft/goutils/errors"
)

func NewTarDirNode(ctx Context, tr *tar.Reader) (*DirNode, error) {
	d := NewDirNode(ctx)
	links := map[string]string{}
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				for s, t := range links {
					f := lookupFile(d, t)
					if f == nil {
						return nil, fmt.Errorf("cannot resolve link %q->%q", s, t)
					}
					s = path.Clean(s)
					dd, err := lookupDir(d, path.Dir(s), true)
					if err != nil {
						return nil, err
					}
					dd.AddNode(path.Base(s), f)
				}
				d.Complete()
				return d, nil
			}
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			_, err := lookupDir(d, header.Name, true)
			if err != nil {
				return nil, fmt.Errorf("file %s: %w", header.Name, err)
			}
		case tar.TypeReg:
			_, err := createFile(d, header, tr)
			if err != nil {
				return nil, fmt.Errorf("file %s: %w", header.Name, err)
			}
		case tar.TypeLink:
			links[header.Name] = header.Linkname
		case tar.TypeSymlink:
			_, err := createLink(d, header.Name, header.Linkname)
			if err != nil {
				return nil, fmt.Errorf("symlink %s: %w", header.Name, err)
			}
		default:
			return nil, fmt.Errorf("unsupported file type %c", header.Typeflag)
		}
	}
}
