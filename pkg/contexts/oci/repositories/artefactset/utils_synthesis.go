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

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer"
)

const SynthesizedBlobFormat = "+tar+gzip"

type ArtefactBlob interface {
	accessio.TemporaryFileSystemBlobAccess
}

type Producer func(set *ArtefactSet) error

func SythesizeArtefactSet(mime string, producer Producer) (ArtefactBlob, error) {
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
	err = producer(set)
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

	return SythesizeArtefactSet(blob.MimeType(), func(set *ArtefactSet) error {
		err = TransferArtefact(art, set)
		if err != nil {
			return fmt.Errorf("failed to transfer artifact: %w", err)
		}

		if ok, _ := artdesc.IsDigest(ref); !ok {
			err = set.AddTags(digest, ref)
			if err != nil {
				return fmt.Errorf("failed to add tag: %w", err)
			}
		}

		set.Annotate(MAINARTEFACT_ANNOTATION, digest.String())

		return nil
	})
}
