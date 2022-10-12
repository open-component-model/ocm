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

package artefactset

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

const SynthesizedBlobFormat = "+tar+gzip"

type ArtefactBlob interface {
	accessio.TemporaryFileSystemBlobAccess
}

type Producer func(set *ArtefactSet) (string, error)

func SythesizeArtefactSet(producer Producer) (ArtefactBlob, error) {
	fs := osfs.New()
	temp, err := accessio.NewTempFile(fs, "", "artefactblob*.tgz")
	if err != nil {
		return nil, err
	}
	defer temp.Close()

	set, err := Create(accessobj.ACC_CREATE, "", 0o600, accessio.File(temp.Writer().(vfs.File)), accessobj.FormatTGZ)
	if err != nil {
		return nil, err
	}
	mime, err := producer(set)
	err2 := set.Close()
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, err2
	}

	return temp.AsBlob(artdesc.ToContentMediaType(mime) + SynthesizedBlobFormat), nil
}

func TransferArtefact(art cpi.ArtefactAccess, set cpi.ArtefactSink, tags ...string) error {
	return transfer.TransferArtefact(art, set, tags...)
}

// SynthesizeArtefactBlob synthesizes an artefact blob incorporating all side artefacts.
// To support extensions like cosign, we need the namespace access her to find
// additionally objects associated by tags.
func SynthesizeArtefactBlob(ns cpi.NamespaceAccess, ref string) (ArtefactBlob, error) {
	art, err := ns.GetArtefact(ref)
	if err != nil {
		return nil, GetArtifactError{Original: err, Ref: ref}
	}

	defer art.Close()

	blob, err := art.Blob()
	if err != nil {
		return nil, err
	}
	digest := blob.Digest()

	return SythesizeArtefactSet(func(set *ArtefactSet) (string, error) {
		err = TransferArtefact(art, set)
		if err != nil {
			return "", fmt.Errorf("failed to transfer artifact: %w", err)
		}

		if ok, _ := artdesc.IsDigest(ref); !ok {
			err = set.AddTags(digest, ref)
			if err != nil {
				return "", fmt.Errorf("failed to add tag: %w", err)
			}
		}

		set.Annotate(MAINARTEFACT_ANNOTATION, digest.String())
		set.Annotate(LEGACY_MAINARTEFACT_ANNOTATION, digest.String())

		return blob.MimeType(), nil
	})
}

// ArtefactFactory add an artefact to the given set and provides descriptor metadata.
type ArtefactFactory func(set *ArtefactSet) (digest.Digest, string, error)

// ArtefactIterator provides a sequence of artefact factories by successive calls.
// The sequence is finished if nil is returned for the factory.
type ArtefactIterator func() (ArtefactFactory, bool, error)

// ArtefactFeedback is called after an artefact has successfully be added.
type ArtefactFeedback func(blob accessio.BlobAccess, art cpi.ArtefactAccess) error

// ArtefactTransferCreator provides an ArtefactFactory transferring the given artefact.
func ArtefactTransferCreator(art cpi.ArtefactAccess, finalizer *utils.Finalizer, feedback ...ArtefactFeedback) ArtefactFactory {
	return func(set *ArtefactSet) (digest.Digest, string, error) {
		var f utils.Finalizer
		defer f.Finalize()

		f.Include(finalizer)

		blob, err := art.Blob()
		if err != nil {
			return "", "", errors.Wrapf(err, "cannot access artefact manifest blob")
		}
		f.Close(blob)

		err = TransferArtefact(art, set)
		if err != nil {
			return "", "", fmt.Errorf("failed to transfer artifact: %w", err)
		}

		list := errors.ErrListf("add artefact")
		for _, fb := range feedback {
			list.Add(fb(blob, art))
		}
		return blob.Digest(), blob.MimeType(), list.Result()
	}
}

// SynthesizeArtefactBlobFor synthesizes an artefact blob incorporating all artefacts
// provided ba a factory.
func SynthesizeArtefactBlobFor(tag string, iter ArtefactIterator) (ArtefactBlob, error) {
	return SythesizeArtefactSet(func(set *ArtefactSet) (string, error) {
		mime := artdesc.MediaTypeImageManifest
		for {
			art, main, err := iter()
			if err != nil || art == nil {
				return mime, err
			}

			digest, _mime, err := art(set)
			if err != nil {
				return "", err
			}
			if main {
				if mime != "" {
					mime = _mime
				}
				set.Annotate(MAINARTEFACT_ANNOTATION, digest.String())
				set.Annotate(LEGACY_MAINARTEFACT_ANNOTATION, digest.String())
				if tag != "" {
					err = set.AddTags(digest, tag)
					if err != nil {
						return "", fmt.Errorf("failed to add tag: %w", err)
					}
				}
			}
		}
	})
}
