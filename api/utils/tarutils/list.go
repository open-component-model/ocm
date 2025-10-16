package tarutils

import (
	"archive/tar"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/compression"
)

func ListArchiveContent(path string, fss ...vfs.FileSystem) ([]string, error) {
	sfs := utils.OptionalDefaulted(osfs.New(), fss...)

	f, err := sfs.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", path)
	}
	defer f.Close()
	return ListArchiveContentFromReader(f)
}

func ListArchiveContentFromReader(r io.Reader) ([]string, error) {
	in, _, err := compression.AutoDecompress(r)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot determine compression")
	}

	var result []string

	tr := tar.NewReader(in)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return result, nil
			}
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			result = append(result, header.Name)
		case tar.TypeSymlink, tar.TypeLink:
			result = append(result, header.Name)
		case tar.TypeReg:
			result = append(result, header.Name)
		}
	}
}
