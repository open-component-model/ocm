// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ocireg

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/index"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

/*
   A common transport archive is just a folder with artefact archives.
   in tar format and an index.json file. The name of the archive
   is the digest of the artefact descriptor.

   The artefact archive is a filesystem structure with a file
   artefact-descriptor.json and a folder blobs containing
   the flat blob files with the name according to the blob digest.

   Digests used as filename will replace the ":" by a "-"
*/

const ArtefactDescriptor = "artefact-descriptor.json"
const ArtefactIndex = "artefact-index.json"

type Repository struct {
	fs     vfs.FileSystem
	ctx    cpi.Context
	mode   os.FileMode
	closer RepositoryCloser

	artefacts *index.RepositoryIndex
}

// NewRepository returns a new representation based repository
func NewRepository(ctx cpi.Context, fs vfs.FileSystem, closer RepositoryCloser, mode os.FileMode) (*Repository, error) {
	repo := &Repository{
		fs:     fs,
		ctx:    ctx,
		mode:   mode,
		closer: closer,
	}
	if err := repo.BuildIndex(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Write(path string, mode os.FileMode, opts ...CTFOption) error {
	o := CTFOptions{}.ApplyOptions(opts...).Default()
	f := GetRepositoryFormat(*o.FileFormat)
	if f == nil {
		return errors.ErrUnknown("repository format", string(*o.FileFormat))
	}
	return f.Write(r, path, o, mode)
}

func (r *Repository) BuildIndex() error {
	data, err := vfs.ReadFile(r.fs, filepath.Join("/", ArtefactIndex))
	if err != nil {
		return fmt.Errorf("unable to read the artefact index from %s: %w", ArtefactIndex, err)
	}
	idx := index.ArtefactIndex{}
	err = json.Unmarshal(data, &idx)
	if err != nil {
		return fmt.Errorf("unable to parse artefact index read from %s: %w", ArtefactIndex, err)
	}

	r.artefacts = index.NewRepositoryIndex()
	for _, a := range idx.Index {
		r.artefacts.AddArtefact(&a)
	}
	return nil
}

func ExtractArtefactDescriptorFromTAR(name string, reader io.Reader) (*artdesc.ArtefactDescriptor, error) {
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil, errors.ErrNotFound("file", ArtefactDescriptor, name)
			}
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			if header.Name == ArtefactDescriptor {
				data, err := ioutil.ReadAll(tr)
				if err != nil {
					return nil, fmt.Errorf("unable to read artefact descriptor: %w", err)
				}
				return artdesc.Decode(data)
			}
		}
	}
}
